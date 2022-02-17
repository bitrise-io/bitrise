package pathutil

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListPathInDirSortedByComponents ...
func ListPathInDirSortedByComponents(searchDir string, relPath bool) ([]string, error) {
	searchDir, err := filepath.Abs(searchDir)
	if err != nil {
		return []string{}, err
	}

	var fileList []string

	if err := filepath.Walk(searchDir, func(path string, _ os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if relPath {
			rel, err := filepath.Rel(searchDir, path)
			if err != nil {
				return err
			}
			path = rel
		}

		fileList = append(fileList, path)

		return nil
	}); err != nil {
		return []string{}, err
	}
	return SortPathsByComponents(fileList)
}

// SortablePath ...
type SortablePath struct {
	Pth        string
	AbsPth     string
	Components []string
}

// NewSortablePath ...
func NewSortablePath(pth string) (SortablePath, error) {
	absPth, err := AbsPath(pth)
	if err != nil {
		return SortablePath{}, err
	}

	components := strings.Split(absPth, string(os.PathSeparator))
	fixedComponents := []string{}
	for _, comp := range components {
		if comp != "" {
			fixedComponents = append(fixedComponents, comp)
		}
	}

	return SortablePath{
		Pth:        pth,
		AbsPth:     absPth,
		Components: fixedComponents,
	}, nil
}

// BySortablePathComponents ..
type BySortablePathComponents []SortablePath

func (s BySortablePathComponents) Len() int {
	return len(s)
}
func (s BySortablePathComponents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s BySortablePathComponents) Less(i, j int) bool {
	path1 := s[i]
	path2 := s[j]

	d1 := len(path1.Components)
	d2 := len(path2.Components)

	if d1 < d2 {
		return true
	} else if d1 > d2 {
		return false
	}

	// if same component size,
	// do alphabetic sort based on the last component
	base1 := filepath.Base(path1.AbsPth)
	base2 := filepath.Base(path2.AbsPth)

	return base1 < base2
}

// SortPathsByComponents ...
func SortPathsByComponents(paths []string) ([]string, error) {
	sortableFiles := []SortablePath{}
	for _, pth := range paths {
		sortable, err := NewSortablePath(pth)
		if err != nil {
			return []string{}, err
		}
		sortableFiles = append(sortableFiles, sortable)
	}

	sort.Sort(BySortablePathComponents(sortableFiles))

	sortedFiles := []string{}
	for _, pth := range sortableFiles {
		sortedFiles = append(sortedFiles, pth.Pth)
	}

	return sortedFiles, nil
}
