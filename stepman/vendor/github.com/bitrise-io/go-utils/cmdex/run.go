package cmdex

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/errorutil"
)

// ----------

// CommandModel ...
type CommandModel struct {
	cmd *exec.Cmd
}

// NewCommand ...
func NewCommand(name string, args ...string) *CommandModel {
	return &CommandModel{
		cmd: exec.Command(name, args...),
	}
}

// NewCommandFromSlice ...
func NewCommandFromSlice(cmdSlice []string) (*CommandModel, error) {
	if len(cmdSlice) == 0 {
		return nil, errors.New("no command provided")
	} else if len(cmdSlice) == 1 {
		return NewCommand(cmdSlice[0]), nil
	}

	return NewCommand(cmdSlice[0], cmdSlice[1:]...), nil
}

// NewCommandWithCmd ...
func NewCommandWithCmd(cmd *exec.Cmd) *CommandModel {
	return &CommandModel{
		cmd: cmd,
	}
}

// GetCmd ...
func (command *CommandModel) GetCmd() *exec.Cmd {
	return command.cmd
}

// SetDir ...
func (command *CommandModel) SetDir(dir string) *CommandModel {
	command.cmd.Dir = dir
	return command
}

// SetEnvs ...
func (command *CommandModel) SetEnvs(envs []string) *CommandModel {
	command.cmd.Env = envs
	return command
}

// SetStdin ...
func (command *CommandModel) SetStdin(in io.Reader) *CommandModel {
	command.cmd.Stdin = in
	return command
}

// SetStdout ...
func (command *CommandModel) SetStdout(out io.Writer) *CommandModel {
	command.cmd.Stdout = out
	return command
}

// SetStderr ...
func (command *CommandModel) SetStderr(err io.Writer) *CommandModel {
	command.cmd.Stderr = err
	return command
}

// Run ...
func (command CommandModel) Run() error {
	return command.cmd.Run()
}

// RunAndReturnExitCode ...
func (command CommandModel) RunAndReturnExitCode() (int, error) {
	return RunCmdAndReturnExitCode(command.cmd)
}

// RunAndReturnTrimmedOutput ...
func (command CommandModel) RunAndReturnTrimmedOutput() (string, error) {
	return RunCmdAndReturnTrimmedOutput(command.cmd)
}

// RunAndReturnTrimmedCombinedOutput ...
func (command CommandModel) RunAndReturnTrimmedCombinedOutput() (string, error) {
	return RunCmdAndReturnTrimmedCombinedOutput(command.cmd)
}

// ----------

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

// RunCmdAndReturnExitCode ...
func RunCmdAndReturnExitCode(cmd *exec.Cmd) (int, error) {
	err := cmd.Run()
	if err != nil {
		exitCode, castErr := errorutil.CmdExitCodeFromError(err)
		if castErr != nil {
			return 1, fmt.Errorf("failed get exit code from error: %s, error: %s", err, castErr)
		}

		return exitCode, err
	}

	return 0, nil
}

// RunCmdAndReturnTrimmedOutput ...
func RunCmdAndReturnTrimmedOutput(cmd *exec.Cmd) (string, error) {
	outBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if outBytes == nil {
		return "", nil
	}
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunCmdAndReturnTrimmedCombinedOutput ...
func RunCmdAndReturnTrimmedCombinedOutput(cmd *exec.Cmd) (string, error) {
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	if outBytes == nil {
		return "", nil
	}
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
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

	return RunCmdAndReturnExitCode(cmd)
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
	return RunCmdAndReturnTrimmedOutput(cmd)
}

// RunCommandInDirAndReturnCombinedStdoutAndStderr ...
func RunCommandInDirAndReturnCombinedStdoutAndStderr(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return RunCmdAndReturnTrimmedCombinedOutput(cmd)
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
