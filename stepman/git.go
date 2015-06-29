package main

import (
	"os"

	"github.com/bitrise-io/go-pathutil"
)

func pullStepsSpec(stepsSpecDir string) error {
	if err := runCommandInDir(stepsSpecDir, "git", []string{"pull"}...); err != nil {
		return err
	}
	return nil
}

func cloneStepsSpecs(stepsSpecDir string) error {
	if err := runCommand("git", []string{"clone", "--recursive", STEP_COLLECTION_GIT, stepsSpecDir}...); err != nil {
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
