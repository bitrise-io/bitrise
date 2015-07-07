package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

func IsPathExists(pth string) (bool, error) {
	if pth == "" {
		return false, errors.New("No path provided")
	}
	_, err := os.Stat(pth)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Expands ENV vars and the ~ character
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
