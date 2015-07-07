package stepman

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-pathutil"
)

func DoGitPull(pth string) error {
	return RunCommandInDir(pth, "git", []string{"pull"}...)
}

func DoGitClone(git, pth string) error {
	return RunCommand("git", []string{"clone", "--recursive", git, pth}...)
}

func DoGitUpdate(git, pth string) error {
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if exists == false {
		fmt.Println("Git path does not exist, do clone")
		return DoGitClone(git, pth)
	}

	fmt.Println("Git path exist, do pull")
	return DoGitPull(pth)
}

func clearPathIfExist(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(pth); err != nil {
			return err
		}
	}
	return nil
}
