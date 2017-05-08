package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	ver "github.com/hashicorp/go-version"
)

//=======================================
// Util
//=======================================

func filterVersionTags(tagList []string) []*ver.Version {
	versionTags := []*ver.Version{}
	for _, tag := range tagList {
		versionTag, err := ver.NewVersion(tag)
		if err == nil && versionTag != nil {
			versionTags = append(versionTags, versionTag)
		}
	}
	return versionTags
}

//=======================================
// Git
//=======================================

func createError(prinatableCmd, cmdOut string, cmdErr error) error {
	message := fmt.Sprintf("command (%s) failed", prinatableCmd)
	if len(cmdOut) > 0 {
		message += "\nout: " + cmdOut
	}
	if !errorutil.IsExitStatusError(cmdErr) {
		message += "\nerr: " + cmdErr.Error()
	}
	return errors.New(message)
}

func runAndHandle(cmd *command.Model) error {
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return createError(cmd.PrintableCommandArgs(), out, err)
	}
	return nil
}

func runForOutputAndHandle(cmd *command.Model) (string, error) {
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", createError(cmd.PrintableCommandArgs(), out, err)
	}
	return out, nil
}

func commitHashOfTag(cloneIntoDir, tag string) (string, error) {
	cmd := command.New("git", "show-ref", "--hash", tag)
	cmd.SetDir(cloneIntoDir)
	return runForOutputAndHandle(cmd)
}

func gitRemoteTagList(cloneIntoDir string) ([]string, error) {
	cmd := command.New("git", "ls-remote", "--tags")
	cmd.SetDir(cloneIntoDir)
	out, err := runForOutputAndHandle(cmd)
	if err != nil {
		return []string{}, err
	}
	if out == "" {
		return []string{}, nil
	}

	var exp = regexp.MustCompile(`(^[a-z0-9]+)+.*refs/tags/([0-9.]+)`)
	versionMap := map[string]bool{}
	outSplit := strings.Split(out, "\n")

	for _, line := range outSplit {
		result := exp.FindAllStringSubmatch(line, -1)
		if len(result) > 0 {
			matches := result[0]

			if len(matches) == 3 {
				version := matches[2]
				versionMap[version] = true
			}
		}
	}

	versions := []string{}
	for key := range versionMap {
		versions = append(versions, key)
	}

	return versions, nil
}

func gitInit(cloneIntoDir string) error {
	cmd := command.New("git", "init")
	cmd.SetDir(cloneIntoDir)
	return runAndHandle(cmd)
}

func gitAddRemote(cloneIntoDir, repositoryURL string) error {
	cmd := command.New("git", "remote", "add", "origin", repositoryURL)
	cmd.SetDir(cloneIntoDir)
	return runAndHandle(cmd)
}

func gitFetch(cloneIntoDir string) error {
	cmd := command.New("git", "fetch")
	cmd.SetDir(cloneIntoDir)
	return runAndHandle(cmd)
}

func gitCheckout(cloneIntoDir, gitCheckoutParam string) error {
	cmd := command.New("git", "checkout", gitCheckoutParam)
	cmd.SetDir(cloneIntoDir)
	return runAndHandle(cmd)
}

func gitLog(cloneIntoDir, formatParam string) (string, error) {
	cmd := command.New("git", "log", "-1", "--format="+formatParam)
	cmd.SetDir(cloneIntoDir)
	return runForOutputAndHandle(cmd)
}

func gitInitWithRemote(cloneIntoDir, repositoryURL string) error {
	gitCheckPath := filepath.Join(cloneIntoDir, ".git")
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

// ByVersion ..
type ByVersion []*ver.Version

func (s ByVersion) Len() int {
	return len(s)
}
func (s ByVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByVersion) Less(i, j int) bool {
	return s[i].LessThan(s[j])
}

// GitVersionTags ...
func GitVersionTags(gitRepoDir string) ([]*ver.Version, error) {
	tagList, err := gitRemoteTagList(gitRepoDir)
	if err != nil {
		return []*ver.Version{}, fmt.Errorf("Could not get version tag list, error: %s", err)
	}

	tags := filterVersionTags(tagList)

	sort.Sort(ByVersion(tags))

	return tags, nil
}

// GitCloneAndCheckoutVersionOrLatestVersion ...
func GitCloneAndCheckoutVersionOrLatestVersion(cloneIntoDir, repositoryURL, checkoutVersion string) (string, error) {
	if err := gitInitWithRemote(cloneIntoDir, repositoryURL); err != nil {
		return "", fmt.Errorf("git init failed, error: %s", err)
	}

	if checkoutVersion == "" {
		versionTagList, err := GitVersionTags(cloneIntoDir)
		if err != nil {
			return "", fmt.Errorf("could not get version tag list, error: %s", err)
		}

		if len(versionTagList) == 0 {
			return "", fmt.Errorf("no version tag found")
		}

		versionPtr := versionTagList[len(versionTagList)-1]
		if versionPtr == nil {
			return "", fmt.Errorf("uninitialized version found")
		}

		checkoutVersion = versionPtr.String()
	}

	if err := gitCheckout(cloneIntoDir, checkoutVersion); err != nil {
		return "", fmt.Errorf("could not checkout (%s), err :%s", checkoutVersion, err)
	}

	return checkoutVersion, nil
}
