package stepman

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
)

// DoGitPull ...
func DoGitPull(pth string) error {
	return RunCommandInDir(pth, "git", []string{"pull"}...)
}

// DoGitClone ...
func DoGitClone(uri, pth string) error {
	if uri == "" {
		return errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return errors.New("Git Clone 'pth' missing")
	}
	return RunCommand("git", []string{"clone", "--recursive", uri, pth}...)
}

// DoGitCheckout ...
func DoGitCheckout(commithash string) error {
	if commithash == "" {
		return errors.New("Git Clone 'hash' missing")
	}
	return RunCommand("git", []string{"checkout", commithash}...)
}

// DoGitCloneWithCommit ...
func DoGitCloneWithCommit(uri, pth, version, commithash string) error {
	if uri == "" {
		return errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return errors.New("Git Clone 'pth' missing")
	}
	if commithash == "" {
		return errors.New("Git Clone 'hash' missing")
	}
	if err := RunCommand("git", []string{"clone", "--recursive", uri, pth, "--branch", version}...); err != nil {
		return err
	}

	cmd := exec.Command("git", []string{"rev-parse", "HEAD"}...)
	cmd.Dir = pth
	bytes, err := cmd.CombinedOutput()
	cmdOutput := string(bytes)
	if err != nil {
		log.Error(cmdOutput)
		return err
	}
	if commithash != cmdOutput {
		return fmt.Errorf("Commit hash doesn't match the one specified for the version tag. (version tag: %s) (expected: %s) (got: %s)", version, cmdOutput, commithash)
	}

	return DoGitCheckout(commithash)
}

// DoGitUpdate ...
func DoGitUpdate(git, pth string) error {
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exists {
		log.Info("[STEPMAN] - Git path does not exist, do clone")
		return DoGitClone(git, pth)
	}

	log.Info("[STEPMAN] - Git path exist, do pull")
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
