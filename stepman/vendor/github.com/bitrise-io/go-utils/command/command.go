package command

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

// Model ...
type Model struct {
	cmd *exec.Cmd
}

// New ...
func New(name string, args ...string) *Model {
	return &Model{
		cmd: exec.Command(name, args...),
	}
}

// NewWithStandardOuts - same as NewCommand, but sets the command's
// stdout and stderr to the standard (OS) out (os.Stdout) and err (os.Stderr)
func NewWithStandardOuts(name string, args ...string) *Model {
	return New(name, args...).SetStdout(os.Stdout).SetStderr(os.Stderr)
}

// NewWithParams ...
func NewWithParams(params ...string) (*Model, error) {
	if len(params) == 0 {
		return nil, errors.New("no command provided")
	} else if len(params) == 1 {
		return New(params[0]), nil
	}

	return New(params[0], params[1:]...), nil
}

// NewFromSlice ...
func NewFromSlice(slice []string) (*Model, error) {
	return NewWithParams(slice...)
}

// NewWithCmd ...
func NewWithCmd(cmd *exec.Cmd) *Model {
	return &Model{
		cmd: cmd,
	}
}

// GetCmd ...
func (m *Model) GetCmd() *exec.Cmd {
	return m.cmd
}

// SetDir ...
func (m *Model) SetDir(dir string) *Model {
	m.cmd.Dir = dir
	return m
}

// SetEnvs ...
func (m *Model) SetEnvs(envs ...string) *Model {
	m.cmd.Env = envs
	return m
}

// AppendEnvs - appends the envs to the current os.Environ()
// Calling this multiple times will NOT appens the envs one by one,
// only the last "envs" set will be appended to os.Environ()!
func (m *Model) AppendEnvs(envs ...string) *Model {
	return m.SetEnvs(append(os.Environ(), envs...)...)
}

// SetStdin ...
func (m *Model) SetStdin(in io.Reader) *Model {
	m.cmd.Stdin = in
	return m
}

// SetStdout ...
func (m *Model) SetStdout(out io.Writer) *Model {
	m.cmd.Stdout = out
	return m
}

// SetStderr ...
func (m *Model) SetStderr(err io.Writer) *Model {
	m.cmd.Stderr = err
	return m
}

// Run ...
func (m Model) Run() error {
	return m.cmd.Run()
}

// RunAndReturnExitCode ...
func (m Model) RunAndReturnExitCode() (int, error) {
	return RunCmdAndReturnExitCode(m.cmd)
}

// RunAndReturnTrimmedOutput ...
func (m Model) RunAndReturnTrimmedOutput() (string, error) {
	return RunCmdAndReturnTrimmedOutput(m.cmd)
}

// RunAndReturnTrimmedCombinedOutput ...
func (m Model) RunAndReturnTrimmedCombinedOutput() (string, error) {
	return RunCmdAndReturnTrimmedCombinedOutput(m.cmd)
}

// PrintableCommandArgs ...
func (m Model) PrintableCommandArgs() string {
	return PrintableCommandArgs(false, m.cmd.Args)
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
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// RunCmdAndReturnTrimmedCombinedOutput ...
func RunCmdAndReturnTrimmedCombinedOutput(cmd *exec.Cmd) (string, error) {
	outBytes, err := cmd.CombinedOutput()
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
