package pathutil

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// FilterPaths applies filters to an explicit list of OS paths and returns
// the ones that pass every filter, preserving input order. Each path is
// stat-ed opportunistically so filters that consult fs.DirEntry (e.g.
// IsDirectoryFilter) work; paths that do not exist on disk are still fed
// to the filters with a nil DirEntry, matching v1's purely lexical
// behavior for path-only filters. Stat errors other than "not exist"
// surface to the caller.
//
// This adapter preserves the v1 go-utils pathutil.FilterPaths signature so
// callers can migrate by import-path rename only. Prefer FilterFS in new code.
func FilterPaths(paths []string, filters ...FilterFunc) ([]string, error) {
	var filtered []string
	for _, pth := range paths {
		var d fs.DirEntry
		info, err := os.Lstat(pth)
		switch {
		case err == nil:
			d = fs.FileInfoToDirEntry(info)
		case errors.Is(err, fs.ErrNotExist):
			// Leave d nil; path-only filters still work.
		default:
			return nil, err
		}

		keep, err := runFilters(filepath.ToSlash(pth), d, filters)
		if err != nil {
			return nil, err
		}
		if keep {
			filtered = append(filtered, pth)
		}
	}
	return filtered, nil
}

// ListEntries lists the immediate children of dir, applies filters, and
// returns the matching entries as absolute paths. It is non-recursive.
//
// This adapter preserves the v1 go-utils pathutil.ListEntries signature so
// callers can migrate by import-path rename only. Prefer FilterFS on an
// fs.FS in new code.
func ListEntries(dir string, filters ...FilterFunc) ([]string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, entry := range entries {
		pth := filepath.Join(absDir, entry.Name())
		keep, err := runFilters(filepath.ToSlash(pth), entry, filters)
		if err != nil {
			return nil, err
		}
		if keep {
			filtered = append(filtered, pth)
		}
	}
	return filtered, nil
}

// runFilters evaluates filters in order. fs.SkipDir from a filter is
// treated as "exclude this entry" rather than a walk directive, because the
// compat adapters do not walk a tree.
func runFilters(pth string, d fs.DirEntry, filters []FilterFunc) (bool, error) {
	for _, filter := range filters {
		matches, err := filter(pth, d)
		if err != nil {
			if errors.Is(err, fs.SkipDir) {
				return false, nil
			}
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	return true, nil
}
