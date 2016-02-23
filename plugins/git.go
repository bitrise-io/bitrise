package plugins

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bitrise-io/depman/pathutil"
	ver "github.com/hashicorp/go-version"
)

//=======================================
// Util
//=======================================

func filterVersionTags(tagList []string) []ver.Version {
	versionTags := []ver.Version{}
	for _, tag := range tagList {
		versionTag, err := ver.NewVersion(tag)
		if err == nil && versionTag != nil {
			versionTags = append(versionTags, *versionTag)
		}
	}
	return versionTags
}

//=======================================
// Git
//=======================================

func commitHashOfTag(cloneIntoDir, tag string) (string, error) {
	return commandOutput(cloneIntoDir, "git", "show-ref", "--hash", tag)
}

func gitTagList(cloneIntoDir string) ([]string, error) {
	out, err := commandOutput(cloneIntoDir, "git", "tag", "--list")
	if err != nil {
		return []string{}, err
	}

	versions := []string{}
	if out == "" {
		return versions, nil
	}

	outSplit := strings.Split(out, "\n")
	for _, line := range outSplit {
		strippedLine := strip(line)
		versions = append(versions, strippedLine)
	}
	return versions, nil
}

func gitInit(cloneIntoDir string) error {
	return command(cloneIntoDir, "git", "init")
}

func gitAddRemote(cloneIntoDir, repositoryURL string) error {
	return command(cloneIntoDir, "git", "remote", "add", "origin", repositoryURL)
}

func gitFetch(cloneIntoDir string) error {
	return command(cloneIntoDir, "git", "fetch")
}

func gitCheckout(cloneIntoDir, gitCheckoutParam string) error {
	return command(cloneIntoDir, "git", "checkout", gitCheckoutParam)
}

func gitLog(cloneIntoDir, formatParam string) (string, error) {
	return commandOutput(cloneIntoDir, "git", "log", "-1", "--format="+formatParam)
}

func gitInitWithRemote(cloneIntoDir, repositoryURL string) error {
	gitCheckPath := path.Join(cloneIntoDir, ".git")
	if exist, err := pathutil.IsPathExists(gitCheckPath); err != nil {
		return fmt.Errorf("Failed to file path (%s), err: %s", gitCheckPath, err)
	} else if exist {
		return fmt.Errorf(".git folder already exists in the destination dir (%s)", gitCheckPath)
	}

	if err := os.MkdirAll(cloneIntoDir, 0777); err != nil {
		return fmt.Errorf("Failed to create the clone_destination_dir at: %s", cloneIntoDir)
	}

	if err := gitInit(cloneIntoDir); err != nil {
		return fmt.Errorf("Could not init git repository, err: %s", cloneIntoDir)
	}

	if err := gitAddRemote(cloneIntoDir, repositoryURL); err != nil {
		return fmt.Errorf("Could not add remote, err: %s", err)
	}

	if err := gitFetch(cloneIntoDir); err != nil {
		return fmt.Errorf("Could not fetch from repository, err: %s", err)
	}

	return nil
}

//=======================================
// Main
//=======================================

// GitVersionTags ...
func GitVersionTags(gitRepoDir string) ([]ver.Version, error) {
	tagList, err := gitTagList(gitRepoDir)
	if err != nil {
		return []ver.Version{}, fmt.Errorf("Could not get version tag list, error: %s", err)
	}

	return filterVersionTags(tagList), nil
}

// GitCloneAndCheckoutVersion ...
func GitCloneAndCheckoutVersion(cloneIntoDir, repositoryURL, checkoutVersion string) (*ver.Version, string, error) {
	if err := gitInitWithRemote(cloneIntoDir, repositoryURL); err != nil {
		return nil, "", err
	}

	var version ver.Version

	if checkoutVersion == "" {
		versionTagList, err := GitVersionTags(cloneIntoDir)
		if err != nil {
			return nil, "", fmt.Errorf("Could not get version tag list, error: %s", err)
		}

		version = versionTagList[len(versionTagList)-1]
	} else {
		versionPtr, err := ver.NewVersion(checkoutVersion)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse version (%s), error: %s", checkoutVersion, err)
		}

		if versionPtr == nil {
			return nil, "", errors.New("failed to parse version (%s), error: nil version")
		}

		version = *versionPtr
	}

	if err := gitCheckout(cloneIntoDir, version.String()); err != nil {
		return nil, "", fmt.Errorf("Could not checkout, err :%s", err)
	}

	hash, err := commitHashOfTag(cloneIntoDir, version.String())
	if err != nil {
		return nil, "", fmt.Errorf("Could get commit hash of tag (%s), err :%s", version.String(), err)
	}

	return &version, hash, nil
}
