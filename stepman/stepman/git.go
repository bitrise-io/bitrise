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

// DoGitCheckoutBranch ...
func DoGitCheckoutBranch(repoPath, branch string) error {
	if branch == "" {
		return errors.New("Git checkout 'branch' missing")
	}
	if err := DoGitCheckout(repoPath, branch); err != nil {
		return RunCommandInDir(repoPath, "git", []string{"checkout", "-b", branch}...)
	}
	return nil
}

// DoGitCheckout ...
func DoGitCheckout(repoPath, checkout string) error {
	if checkout == "" {
		return errors.New("Git checkout 'hash' missing")
	}
	return RunCommandInDir(repoPath, "git", []string{"checkout", checkout}...)
}

// DoGitAddFile ...
func DoGitAddFile(repoPath, filePath string) error {
	if filePath == "" {
		return errors.New("Git add 'file' missing")
	}
	return RunCommandInDir(repoPath, "git", []string{"add", filePath}...)
}

// DoGitPushToOrigin ...
func DoGitPushToOrigin(repoPath, branch string) error {
	return RunCommandInDir(repoPath, "git", []string{"push", "-u", "origin", branch}...)
}

// CheckIsNoGitChanges ...
func CheckIsNoGitChanges(repoPath string) error {
	return RunCommandInDir(repoPath, "git", []string{"diff", "--cached", "--exit-code", "--quiet"}...)
}

// DoGitCommit ...
func DoGitCommit(repoPath string, message string) error {
	if message == "" {
		return errors.New("Git commit 'message' missing")
	}
	if err := CheckIsNoGitChanges(repoPath); err != nil {
		return RunCommandInDir(repoPath, "git", []string{"commit", "-m", message}...)
	}
	return nil
}

// GetLatestGitCommitHashOnHead ...
func GetLatestGitCommitHashOnHead(pth string) (string, error) {
	cmd := exec.Command("git", []string{"rev-parse", "HEAD"}...)
	cmd.Dir = pth
	bytes, err := cmd.CombinedOutput()
	cmdOutput := string(bytes)
	if err != nil {
		log.Error(cmdOutput)
		return "", err
	}
	return cmdOutput, nil
}

// DoGitCloneVersion ...
func DoGitCloneVersion(uri, pth, version string) (err error) {
	if uri == "" {
		return errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return errors.New("Git Clone 'pth' missing")
	}
	if version == "" {
		return errors.New("Git Clone 'version' missing")
	}
	return RunCommand("git", []string{"clone", "--recursive", uri, pth, "--branch", version}...)
}

// DoGitCloneVersionAndCommit ...
func DoGitCloneVersionAndCommit(uri, pth, version, commithash string) (err error) {
	if uri == "" {
		return errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return errors.New("Git Clone 'pth' missing")
	}
	if version == "" {
		return errors.New("Git Clone 'version' missing")
	}
	if commithash == "" {
		return errors.New("Git Clone 'commithash' missing")
	}
	if err = RunCommand("git", []string{"clone", "--recursive", uri, pth, "--branch", version}...); err != nil {
		return
	}

	// cleanup
	defer func() {
		if err != nil {
			if err := RemoveDir(pth); err != nil {
				log.Errorln("Failed to cleanup path: ", pth, " | err: ", err)
			}
		}
	}()

	latestCommit, err := GetLatestGitCommitHashOnHead(pth)
	if err != nil {
		return
	}
	if commithash != latestCommit {
		return fmt.Errorf("Commit hash doesn't match the one specified for the version tag. (version tag: %s) (expected: %s) (got: %s)", version, latestCommit, commithash)
	}

	if err = DoGitCheckout(pth, commithash); err != nil {
		return
	}

	return
}

// DoGitGetCommit ...
func DoGitGetCommit(pth string) (string, error) {
	cmd := exec.Command("git", []string{"rev-parse", "HEAD"}...)
	cmd.Dir = pth
	bytes, err := cmd.CombinedOutput()
	cmdOutput := string(bytes)
	if err != nil {
		log.Error(cmdOutput)
		return "", err
	}
	return cmdOutput, nil
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
