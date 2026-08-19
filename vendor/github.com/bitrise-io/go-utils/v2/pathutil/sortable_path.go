package pathutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SortablePath pairs a path with its absolute form and component breakdown,
// enabling sorts that compare entries by directory depth.
type SortablePath struct {
	Pth        string
	AbsPth     string
	Components []string
}

// NewSortablePath expands pth to its absolute form and splits it into
// non-empty components suitable for depth-based sorting.
func NewSortablePath(pth string) (SortablePath, error) {
	absPth, err := pathModifier{}.AbsPath(pth)
	if err != nil {
		return SortablePath{}, err
	}

	var components []string
	for _, c := range strings.Split(absPth, string(os.PathSeparator)) {
		if c != "" {
			components = append(components, c)
		}
	}

	return SortablePath{
		Pth:        pth,
		AbsPth:     absPth,
		Components: components,
	}, nil
}

// BySortablePathComponents sorts SortablePath values by component depth
// (shallowest first), breaking ties alphabetically on the base name and
// falling back to the absolute and original paths to keep the order
// deterministic when same-depth same-base entries exist.
type BySortablePathComponents []SortablePath

func (s BySortablePathComponents) Len() int      { return len(s) }
func (s BySortablePathComponents) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s BySortablePathComponents) Less(i, j int) bool {
	if len(s[i].Components) != len(s[j].Components) {
		return len(s[i].Components) < len(s[j].Components)
	}
	baseI, baseJ := filepath.Base(s[i].AbsPth), filepath.Base(s[j].AbsPth)
	if baseI != baseJ {
		return baseI < baseJ
	}
	if s[i].AbsPth != s[j].AbsPth {
		return s[i].AbsPth < s[j].AbsPth
	}
	return s[i].Pth < s[j].Pth
}

// SortPathsByComponents returns paths sorted by directory depth.
func SortPathsByComponents(paths []string) ([]string, error) {
	sortables := make([]SortablePath, 0, len(paths))
	for _, p := range paths {
		sp, err := NewSortablePath(p)
		if err != nil {
			return nil, err
		}
		sortables = append(sortables, sp)
	}

	sort.Sort(BySortablePathComponents(sortables))

	sorted := make([]string, 0, len(sortables))
	for _, sp := range sortables {
		sorted = append(sorted, sp.Pth)
	}
	return sorted, nil
}

// ListPathInDirSortedByComponents walks searchDir recursively and returns
// every path found, sorted by directory depth. If relPath is true the paths
// are returned relative to searchDir.
func ListPathInDirSortedByComponents(searchDir string, relPath bool) ([]string, error) {
	absDir, err := filepath.Abs(searchDir)
	if err != nil {
		return nil, err
	}

	var paths []string
	err = filepath.WalkDir(absDir, func(p string, _ fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if relPath {
			rel, err := filepath.Rel(absDir, p)
			if err != nil {
				return err
			}
			p = rel
		}
		paths = append(paths, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return SortPathsByComponents(paths)
}
