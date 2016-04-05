package cmdex

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// PrintableCommandArgs ...
func PrintableCommandArgs(isQuoteFirst bool, fullCommandArgs []string) string {
	cmdArgsDecorated := []string{}
	for idx, anArg := range fullCommandArgs {
		quotedArg := strconv.Quote(anArg)
		if idx == 0 && !isQuoteFirst {
			quotedArg = anArg
		}
		cmdArgsDecorated = append(cmdArgsDecorated, quotedArg)
	}

	return strings.Join(cmdArgsDecorated, " ")
}

// RunCommandWithReaderAndWriters ...
func RunCommandWithReaderAndWriters(inReader io.Reader, outWriter, errWriter io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = inReader
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	return cmd.Run()
}

// RunCommandWithWriters ...
func RunCommandWithWriters(outWriter, errWriter io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	return cmd.Run()
}

// RunCommandInDirWithEnvsAndReturnExitCode ...
func RunCommandInDirWithEnvsAndReturnExitCode(envs []string, dir, name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}
	if len(envs) > 0 {
		cmd.Env = envs
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

// RunCommandInDirAndReturnExitCode ...
func RunCommandInDirAndReturnExitCode(dir, name string, args ...string) (int, error) {
	return RunCommandInDirWithEnvsAndReturnExitCode([]string{}, dir, name, args...)
}

// RunCommandWithEnvsAndReturnExitCode ...
func RunCommandWithEnvsAndReturnExitCode(envs []string, name string, args ...string) (int, error) {
	return RunCommandInDirWithEnvsAndReturnExitCode(envs, "", name, args...)
}

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

// RunCommand ...
func RunCommand(name string, args ...string) error {
	return RunCommandInDir("", name, args...)
}

// RunCommandAndReturnStdout ..
func RunCommandAndReturnStdout(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunCommandInDirAndReturnCombinedStdoutAndStderr ...
func RunCommandInDirAndReturnCombinedStdoutAndStderr(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunCommandAndReturnCombinedStdoutAndStderr ..
func RunCommandAndReturnCombinedStdoutAndStderr(name string, args ...string) (string, error) {
	return RunCommandInDirAndReturnCombinedStdoutAndStderr("", name, args...)
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
