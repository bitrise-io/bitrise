package stepman

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-pathutil"
)

func DoGitUpdate(git, pth string) error {
	exists, err := pathutil.IsPathExists(pth)
	if err != nil {
		return err
	}
	if exists == false {
		fmt.Println("Git path does not exist, do clone")
		return DoGitClone(git, pth)
	}

	fmt.Println("Git path exist, do pull")
	return DoGitPull(pth)
}

func DoGitPull(pth string) error {
	if err := RunCommandInDir(pth, "git", []string{"pull"}...); err != nil {
		return err
	}
	return nil
}

func DoGitClone(git, pth string) error {
	if err := RunCommand("git", []string{"clone", "--recursive", git, pth}...); err != nil {
		return err
	}
	return nil
}

func clearPathIfExist(pth string) error {
	isStorePathExists, err := pathutil.IsPathExists(pth)
	if err != nil {
		return err
	}
	if isStorePathExists {
		err := os.RemoveAll(pth)
		if err != nil {
			return err
		}
	}
	return nil
}
