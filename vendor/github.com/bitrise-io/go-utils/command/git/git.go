package git

import (
	"fmt"

	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
)

func setStandardOutAndErr(cmd *command.Model) {
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)
}

// CloneCommand ...
func CloneCommand(uri, destination string) *command.Model {
	return command.New("git", "clone", "--recursive", uri, destination)
}

// Clone ...
func Clone(uri, destination string) error {
	cmd := CloneCommand(uri, destination)
	setStandardOutAndErr(cmd)
	return cmd.Run()
}

// CloneTagOrBranchCommand ...
func CloneTagOrBranchCommand(uri, destination, tagOrBranch string) *command.Model {
	return command.New("git", "clone", "--recursive", "--branch", tagOrBranch, uri, destination)
}

// CloneTagOrBranch ...
func CloneTagOrBranch(uri, destination, tagOrBranch string) error {
	cmd := CloneTagOrBranchCommand(uri, destination, tagOrBranch)
	setStandardOutAndErr(cmd)
	return cmd.Run()
}

// BranchCommand ...
func BranchCommand() *command.Model {
	return command.New("git", "branch")
}

// CloneTagAndEnsureHead ...
func CloneTagAndEnsureHead(uri, destination, tag string) error {
	cmd := CloneTagOrBranchCommand(uri, destination, tag)
	setStandardOutAndErr(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("(%s) failed, error: %s", cmd.PrintableCommandArgs(), err)
	}

	cmd = BranchCommand()
	cmd.SetDir(destination)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("(%s) failed, error: %s", cmd.PrintableCommandArgs(), err)
	}

	if out != "* (no branch)" {
		return fmt.Errorf("current HEAD is not detached head, current branch should be: '* (no branch)', got: %s", out)
	}

	return nil
}

// CloneTagOrBranchAndValidateCommitHash ...
func CloneTagOrBranchAndValidateCommitHash(uri, destination, tagOrBranch, commithash string) error {
	cmd := CloneTagOrBranchCommand(uri, destination, tagOrBranch)
	setStandardOutAndErr(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("(%s) failed, error: %s", cmd.PrintableCommandArgs(), err)
	}

	cmd = GetCommitHashOfHeadCommand()
	cmd.SetDir(destination)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("(%s) failed, error: %s", cmd.PrintableCommandArgs(), err)
	}
	if commithash != out {
		return fmt.Errorf("commit hash doesn't match the one specified for the version tag. (version tag: %s) (expected commit hash: %s) (got: %s)", tagOrBranch, out, commithash)
	}

	return nil
}

// PullCommand ...
func PullCommand() *command.Model {
	return command.New("git", "pull")
}

// Pull ...
func Pull(sourceDir string) error {
	cmd := PullCommand()
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// Update ...
func Update(uri, sourceDir string) error {
	if exists, err := pathutil.IsPathExists(sourceDir); err != nil {
		return err
	} else if !exists {
		return Clone(uri, sourceDir)
	}

	return Pull(sourceDir)
}

// CheckoutCommand ...
func CheckoutCommand(branchOrTag string) *command.Model {
	return command.New("git", "checkout", branchOrTag)
}

// Checkout ...
func Checkout(sourceDir, branchOrTag string) error {
	cmd := CheckoutCommand(branchOrTag)
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// CreateAndCheckoutBranchCommand ...
func CreateAndCheckoutBranchCommand(branch string) *command.Model {
	return command.New("git", "checkout", "-b", branch)
}

// CreateAndCheckoutBranch ...
func CreateAndCheckoutBranch(sourceDir, branch string) error {
	cmd := CreateAndCheckoutBranchCommand(branch)
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// AddCommand ...
func AddCommand(pth string) *command.Model {
	return command.New("git", "add", pth)
}

// AddFile ...
func AddFile(sourceDir, filePth string) error {
	cmd := AddCommand(filePth)
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// PushToOriginCommand ...
func PushToOriginCommand(branch string) *command.Model {
	return command.New("git", "push", "-u", "origin", branch)
}

// PushToOrigin ...
func PushToOrigin(sourceDir, branch string) error {
	cmd := PushToOriginCommand(branch)
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// StatusPorcelainCommand ...
func StatusPorcelainCommand() *command.Model {
	return command.New("git", "status", "--porcelain")
}

// CheckIsNoChanges ...
func CheckIsNoChanges(sourceDir string) error {
	cmd := StatusPorcelainCommand()
	cmd.SetDir(sourceDir)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("(%s) failed, error: %s", cmd.PrintableCommandArgs(), err)
	}

	if out != "" {
		return fmt.Errorf("uncommited changes: %s", out)
	}
	return nil
}

// CommitCommand ...
func CommitCommand(message string) *command.Model {
	return command.New("git", "commit", "-m", message)
}

// Commit ...
func Commit(sourceDir string, message string) error {
	cmd := CommitCommand(message)
	setStandardOutAndErr(cmd)
	cmd.SetDir(sourceDir)
	return cmd.Run()
}

// GetCommitHashOfHeadCommand ...
func GetCommitHashOfHeadCommand() *command.Model {
	return command.New("git", "rev-parse", "HEAD")
}

// GetCommitHashOfHead ...
func GetCommitHashOfHead(pth string) (string, error) {
	cmd := GetCommitHashOfHeadCommand()
	cmd.SetDir(pth)
	return cmd.RunAndReturnTrimmedCombinedOutput()
}
