package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-pathutil"
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
	if err := runCommandInDir(pth, "git", []string{"pull"}...); err != nil {
		return err
	}
	return nil
}

func DoGitClone(git, pth string) error {
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
