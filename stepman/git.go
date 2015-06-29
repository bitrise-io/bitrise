package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-pathutil"
)

func doGitUpdate(git, pth string) error {
	exists, err := pathutil.IsPathExists(pth)
	if err != nil {
		return err
	}
	if exists == false {
		fmt.Println("Git path does not exist, do clone")
		return doGitClone(git, pth)
	}

	fmt.Println("Git path exist, do pull")
	return doGitPull(pth)
}

func doGitPull(pth string) error {
	if err := runCommandInDir(pth, "git", []string{"pull"}...); err != nil {
		return err
	}
	return nil
}

func doGitClone(git, pth string) error {
	if err := runCommand("git", []string{"clone", "--recursive", git, pth}...); err != nil {
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
