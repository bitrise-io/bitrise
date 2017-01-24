package command

import (
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
)

// CopyFile ...
func CopyFile(src, dst string) error {
	// replace with a pure Go implementation?
	//  Golang proposal was: https://go-review.googlesource.com/#/c/1591/5/src/io/ioutil/ioutil.go
	args := []string{src, dst}
	return RunCommand("rsync", args...)
}

// CopyDir ...
func CopyDir(src, dst string, isOnlyContent bool) error {
	if isOnlyContent && !strings.HasSuffix(src, "/") {
		src = src + "/"
	}
	args := []string{"-ar", src, dst}
	return RunCommand("rsync", args...)
}

// RemoveDir ...
func RemoveDir(dirPth string) error {
	if exist, err := pathutil.IsPathExists(dirPth); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(dirPth); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFile ...
func RemoveFile(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if exist {
		if err := os.Remove(pth); err != nil {
			return err
		}
	}
	return nil
}
