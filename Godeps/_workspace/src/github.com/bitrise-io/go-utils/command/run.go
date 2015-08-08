package command

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ------------------
// --- Common

// RunCommandInDir ...
func RunCommandInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Run()
}

// RunCommandInDirWithExitCode ...
func RunCommandInDirWithExitCode(dir, name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}

	cmdExitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return 1, errors.New("Failed to cast exit status")
			}
			cmdExitCode = waitStatus.ExitStatus()
		}
		return cmdExitCode, err
	}
	return 0, nil
}

// RunCommand ...
func RunCommand(name string, args ...string) error {
	return RunCommandInDir("", name, args...)
}

// RunCommandAndReturnStdout ..
func RunCommandAndReturnStdout(cmdName string, cmdArgs ...string) (string, error) {
	cmd := exec.Command(cmdName, cmdArgs...)
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunCommandAndReturnCombinedStdoutAndStderr ..
func RunCommandAndReturnCombinedStdoutAndStderr(cmdName string, cmdArgs ...string) (string, error) {
	cmd := exec.Command(cmdName, cmdArgs...)
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunBashCommand ...
func RunBashCommand(cmdStr string) error {
	return RunCommand("bash", "-c", cmdStr)
}

// RunBashCommandLines ...
func RunBashCommandLines(cmdLines []string) error {
	for _, aLine := range cmdLines {
		if err := RunCommand("bash", "-c", aLine); err != nil {
			return err
		}
	}
	return nil
}

// RunCopy ...
func RunCopy(src, dst string) error {
	args := []string{src, dst}
	return RunCommand("rsync", args...)
}

// RunCopyFile ...
func RunCopyFile(src, dst string) error {
	args := []string{src, dst}
	return RunCommand("rsync", args...)
}

// RunCopyDir ...
func RunCopyDir(src, dst string, isOnlyContent bool) error {
	if isOnlyContent && !strings.HasSuffix(src, "/") {
		src = src + "/"
	}
	args := []string{"-r", src, dst}
	return RunCommand("rsync", args...)
}

// RunGitClone ...
func RunGitClone(uri, pth, tagOrBranch string) error {
	if uri == "" {
		return errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return errors.New("Git Clone 'path' missing")
	}
	if tagOrBranch == "" {
		return errors.New("Git Clone 'tag or branch' missing")
	}
	return RunCommand("git", []string{"clone", "--recursive", "--branch", tagOrBranch, uri, pth}...)
}
