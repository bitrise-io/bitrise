package ziputil

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
)

// ZipDir ...
func ZipDir(sourceDirPth, destinationZipPth string, isContentOnly bool) error {
	if exist, err := pathutil.IsDirExists(sourceDirPth); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("dir (%s) not exist", sourceDirPth)
	}

	workDir := filepath.Dir(sourceDirPth)
	if isContentOnly {
		workDir = sourceDirPth
	}

	zipTarget := filepath.Base(sourceDirPth)
	if isContentOnly {
		zipTarget = "."
	}

	// -r - Travel the directory structure recursively
	// -T - Test the integrity of the new zip file
	// -y - Store symbolic links as such in the zip archive, instead of compressing and storing the file referred to by the link
	cmd := command.New("/usr/bin/zip", "-rTy", destinationZipPth, zipTarget)
	cmd.SetDir(workDir)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("command: (%s) failed, output: %s, error: %s", cmd.PrintableCommandArgs(), out, err)
	}

	return nil
}

// ZipFile ...
func ZipFile(sourceFilePth, destinationZipPth string) error {
	if exist, err := pathutil.IsPathExists(sourceFilePth); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("file (%s) not exist", sourceFilePth)
	}

	workDir := filepath.Dir(sourceFilePth)
	zipTarget := filepath.Base(sourceFilePth)

	// -T - Test the integrity of the new zip file
	// -y - Store symbolic links as such in the zip archive, instead of compressing and storing the file referred to by the link
	cmd := command.New("/usr/bin/zip", "-Ty", destinationZipPth, zipTarget)
	cmd.SetDir(workDir)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("command: (%s) failed, output: %s, error: %s", cmd.PrintableCommandArgs(), out, err)
	}

	return nil
}

// UnZip ...
func UnZip(zip, intoDir string) error {
	cmd := command.New("/usr/bin/unzip", zip, "-d", intoDir)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("command: (%s) failed, output: %s, error: %s", cmd.PrintableCommandArgs(), out, err)
	}

	return nil
}
