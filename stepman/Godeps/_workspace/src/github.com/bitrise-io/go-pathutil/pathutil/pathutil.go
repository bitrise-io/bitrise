package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsRelativePath ...
func IsRelativePath(pth string) bool {
	if strings.HasPrefix(pth, "./") {
		return true
	}

	if strings.HasPrefix(pth, "/") {
		return false
	}

	if strings.HasPrefix(pth, "$") {
		return false
	}

	return true
}

func genericIsPathExists(pth string) (os.FileInfo, bool, error) {
	if pth == "" {
		return nil, false, errors.New("No path provided")
	}
	fileInf, err := os.Stat(pth)
	if err == nil {
		return nil, true, nil
	}
	if os.IsNotExist(err) {
		return fileInf, false, nil
	}
	return fileInf, false, err
}

// IsPathExists ...
func IsPathExists(pth string) (bool, error) {
	_, isExists, err := genericIsPathExists(pth)
	return isExists, err
}

// IsDirExists ...
func IsDirExists(pth string) (bool, error) {
	fileInf, isExists, err := genericIsPathExists(pth)
	if err != nil {
		return false, err
	}
	if !isExists {
		return false, nil
	}
	if fileInf == nil {
		return false, errors.New("No file info available.")
	}
	return fileInf.IsDir(), nil
}

// AbsPath expands ENV vars and the ~ character
//	then call Go's Abs
func AbsPath(pth string) (string, error) {
	if pth == "" {
		return "", errors.New("No Path provided")
	}
	if len(pth) >= 2 && pth[:2] == "~/" {
		pth = strings.Replace(pth, "~/", "$HOME/", 1)
	}
	return filepath.Abs(os.ExpandEnv(pth))
}

// CurrentWorkingDirectoryAbsolutePath ...
func CurrentWorkingDirectoryAbsolutePath() (string, error) {
	return filepath.Abs("./")
}

// UserHomeDir ...
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
