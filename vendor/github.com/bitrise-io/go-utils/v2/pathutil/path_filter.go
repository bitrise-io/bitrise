package pathutil

import (
	"errors"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

// FilterFunc decides whether an entry produced by an fs.FS walk should be
// kept. The path is slash-separated and rooted at the fs root ("." for the
// root itself). Returning fs.SkipDir as the error skips the current
// directory's subtree.
type FilterFunc func(pth string, d fs.DirEntry) (bool, error)

// FilterFS walks fsys recursively and returns every path for which all
// filters return true. Paths are slash-separated and relative to fsys.
func FilterFS(fsys fs.FS, filters ...FilterFunc) ([]string, error) {
	var filtered []string

	err := fs.WalkDir(fsys, ".", func(pth string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		for _, filter := range filters {
			matches, err := filter(pth, d)
			if err != nil {
				return err
			}
			if !matches {
				return nil
			}
		}

		filtered = append(filtered, pth)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filtered, nil
}

// BaseFilter matches entries whose base name equals base (case-insensitive).
// When allowed is false the match is inverted.
func BaseFilter(base string, allowed bool) FilterFunc {
	return func(pth string, _ fs.DirEntry) (bool, error) {
		return allowed == strings.EqualFold(path.Base(pth), base), nil
	}
}

// ExtensionFilter matches entries whose extension equals ext
// (case-insensitive, leading dot included, e.g. ".txt").
func ExtensionFilter(ext string, allowed bool) FilterFunc {
	return func(pth string, _ fs.DirEntry) (bool, error) {
		return allowed == strings.EqualFold(path.Ext(pth), ext), nil
	}
}

// RegexpFilter matches entries whose path contains a match for pattern.
// The pattern is compiled once when the filter is constructed.
func RegexpFilter(pattern string, allowed bool) FilterFunc {
	re := regexp.MustCompile(pattern)
	return func(pth string, _ fs.DirEntry) (bool, error) {
		return allowed == (re.FindString(pth) != ""), nil
	}
}

// ComponentFilter matches entries whose path contains component as one of
// its slash-separated components.
func ComponentFilter(component string, allowed bool) FilterFunc {
	return func(pth string, _ fs.DirEntry) (bool, error) {
		found := slices.Contains(strings.Split(pth, "/"), component)
		return allowed == found, nil
	}
}

// ComponentWithExtensionFilter matches entries whose path has at least one
// component with the given extension (case-insensitive, leading dot
// included, e.g. ".xcodeproj").
func ComponentWithExtensionFilter(ext string, allowed bool) FilterFunc {
	return func(pth string, _ fs.DirEntry) (bool, error) {
		found := slices.ContainsFunc(strings.Split(pth, "/"), func(c string) bool {
			return strings.EqualFold(path.Ext(c), ext)
		})
		return allowed == found, nil
	}
}

// IsDirectoryFilter matches entries based on whether they are directories.
func IsDirectoryFilter(allowed bool) FilterFunc {
	return func(_ string, d fs.DirEntry) (bool, error) {
		if d == nil {
			return false, errors.New("no directory entry available")
		}
		return allowed == d.IsDir(), nil
	}
}

// InDirectoryFilter matches entries whose direct parent directory equals dir.
func InDirectoryFilter(dir string, allowed bool) FilterFunc {
	return func(pth string, _ fs.DirEntry) (bool, error) {
		return allowed == (path.Dir(pth) == dir), nil
	}
}

// SkipDirectoryNameFilter keeps every non-directory entry. When it encounters
// a directory whose base name matches dirName (case-insensitive) it returns
// fs.SkipDir so the whole subtree is skipped.
func SkipDirectoryNameFilter(dirName string) FilterFunc {
	return func(pth string, d fs.DirEntry) (bool, error) {
		if d == nil || !d.IsDir() {
			return true, nil
		}
		if strings.EqualFold(path.Base(pth), dirName) {
			return false, fs.SkipDir
		}
		return true, nil
	}
}

// DirectoryContainsFileFilter matches directories that contain a regular
// file named fileName. fsys is used to stat the candidate path.
func DirectoryContainsFileFilter(fsys fs.FS, fileName string) FilterFunc {
	return func(pth string, d fs.DirEntry) (bool, error) {
		if d == nil || !d.IsDir() {
			return false, nil
		}
		info, err := fs.Stat(fsys, path.Join(pth, fileName))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return false, nil
			}
			return false, err
		}
		return !info.IsDir(), nil
	}
}

// FileContainsFilter matches regular files whose contents include content.
// Directories never match.
func FileContainsFilter(fsys fs.FS, content string) FilterFunc {
	return func(pth string, d fs.DirEntry) (bool, error) {
		if d != nil && d.IsDir() {
			return false, nil
		}
		data, err := fs.ReadFile(fsys, pth)
		if err != nil {
			return false, err
		}
		return strings.Contains(string(data), content), nil
	}
}
