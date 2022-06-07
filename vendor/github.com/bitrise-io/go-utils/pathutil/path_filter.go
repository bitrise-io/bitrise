package pathutil

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ListEntries filters contents of a directory using the provided filters
func ListEntries(dir string, filters ...FilterFunc) ([]string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return []string{}, err
	}

	entries, err := ioutil.ReadDir(absDir)
	if err != nil {
		return []string{}, err
	}

	var paths []string
	for _, entry := range entries {
		pth := filepath.Join(absDir, entry.Name())
		paths = append(paths, pth)
	}

	return FilterPaths(paths, filters...)
}

// FilterPaths ...
func FilterPaths(fileList []string, filters ...FilterFunc) ([]string, error) {
	var filtered []string

	for _, pth := range fileList {
		allowed := true
		for _, filter := range filters {
			if allows, err := filter(pth); err != nil {
				return []string{}, err
			} else if !allows {
				allowed = false
				break
			}
		}
		if allowed {
			filtered = append(filtered, pth)
		}
	}

	return filtered, nil
}

// FilterFunc ...
type FilterFunc func(string) (bool, error)

// BaseFilter ...
func BaseFilter(base string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		b := filepath.Base(pth)
		return allowed == strings.EqualFold(base, b), nil
	}
}

// ExtensionFilter ...
func ExtensionFilter(ext string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		e := filepath.Ext(pth)
		return allowed == strings.EqualFold(ext, e), nil
	}
}

// RegexpFilter ...
func RegexpFilter(pattern string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		re := regexp.MustCompile(pattern)
		found := re.FindString(pth) != ""
		return allowed == found, nil
	}
}

// ComponentFilter ...
func ComponentFilter(component string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		found := false
		pathComponents := strings.Split(pth, string(filepath.Separator))
		for _, c := range pathComponents {
			if c == component {
				found = true
			}
		}
		return allowed == found, nil
	}
}

// ComponentWithExtensionFilter ...
func ComponentWithExtensionFilter(ext string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		found := false
		pathComponents := strings.Split(pth, string(filepath.Separator))
		for _, c := range pathComponents {
			e := filepath.Ext(c)
			if e == ext {
				found = true
			}
		}
		return allowed == found, nil
	}
}

// IsDirectoryFilter ...
func IsDirectoryFilter(allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		fileInf, err := os.Lstat(pth)
		if err != nil {
			return false, err
		}
		if fileInf == nil {
			return false, errors.New("no file info available")
		}
		return allowed == fileInf.IsDir(), nil
	}
}

// InDirectoryFilter ...
func InDirectoryFilter(dir string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		in := filepath.Dir(pth) == dir
		return allowed == in, nil
	}
}

// DirectoryContainsFileFilter returns a FilterFunc that checks if a directory contains a file
func DirectoryContainsFileFilter(fileName string) FilterFunc {
	return func(pth string) (bool, error) {
		isDir, err := IsDirectoryFilter(true)(pth)
		if err != nil {
			return false, err
		}
		if !isDir {
			return false, nil
		}

		absPath := filepath.Join(pth, fileName)
		if _, err := os.Lstat(absPath); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
			return false, nil
		}
		return true, nil
	}
}

// FileContainsFilter ...
func FileContainsFilter(pth, str string) (bool, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(bytes), str), nil
}
