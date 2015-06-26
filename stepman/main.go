package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-pathutil"
)

const (
	STEP_COLLECTION_GIT string = "https://github.com/steplib/steplib.git"
	STEP_COLLECTION_DIR string = "/.stepMan/step_collection"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInDir(workingDir, commandName string, args ...string) error {
	cmd := exec.Command(commandName, args...)
	cmd.Dir = workingDir
	return cmd.Run()
}

func pullStepsSpec(cloneDir string) error {
	if err := runCommandInDir(cloneDir, "git", []string{"pull"}...); err != nil {
		return err
	}
	return nil
}

func cloneStepsSpecs(cloneDir string) error {
	if err := runCommand("git", []string{"clone", "--recursive", STEP_COLLECTION_GIT, cloneDir}...); err != nil {
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

func doUpdateCommand(cloneDir string) error {
	isStorePathExists, err := pathutil.IsPathExists(cloneDir)
	if err != nil {
		return err
	}
	if isStorePathExists == false {
		fmt.Println("StepsSpec path does not exist, do clone")
		return cloneStepsSpecs(cloneDir)
	}

	fmt.Println("StepsSpec path exist, do pull")
	return pullStepsSpec(cloneDir)
}

func main() {
	fmt.Println("stepman")

	fmt.Println("start updating")

	cloneDir := pathutil.UserHomeDir() + STEP_COLLECTION_DIR
	err := doUpdateCommand(cloneDir)
	if err != nil {
		fmt.Println("Failed to update:", err)
	} else {
		fmt.Println("Update success!")
	}
}
