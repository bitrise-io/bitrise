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

// DoGitCloneWithCommit ...
func DoGitCloneWithCommit(uri, pth, version, commithash string) (err error) {
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

	if err = DoGitCheckout(commithash); err != nil {
		return
	}

	return
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
