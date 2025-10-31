package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		printableCmd := PrintableCommandArgs(false, append([]string{name}, args...))

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), printableCmd, errors.New(string(out)))
		}
		return fmt.Errorf("executing command failed (%s): %w", printableCmd, err)
	}
	return nil
}

// CopyFile ...
func CopyFile(src, dst string) error {
	// replace with a pure Go implementation?
	// Golang proposal was: https://go-review.googlesource.com/#/c/1591/5/src/io/ioutil/ioutil.go
	isDir, err := pathutil.IsDirExists(src)
	if err != nil {
		return err
	}
	if isDir {
		return errors.New("source is a directory: " + src)
	}
	args := []string{src, dst}
	return runCommand("rsync", args...)
}

// CopyDir ...
func CopyDir(src, dst string, isOnlyContent bool) error {
	if isOnlyContent && !strings.HasSuffix(src, "/") {
		src = src + "/"
	}
	args := []string{"-ar", src, dst}
	return runCommand("rsync", args...)
}

// RemoveDir ...
// Deprecated: use RemoveAll instead.
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
// Deprecated: use RemoveAll instead.
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

// RemoveAll removes recursively every file on the given paths.
func RemoveAll(pths ...string) error {
	for _, pth := range pths {
		if err := os.RemoveAll(pth); err != nil {
			return err
		}
	}
	return nil
}
