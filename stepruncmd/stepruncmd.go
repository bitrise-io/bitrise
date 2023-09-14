package stepruncmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/bitrise-io/bitrise/stepruncmd/timeoutcmd"
	"github.com/bitrise-io/go-utils/v2/log"
)

type Cmd struct {
	cmd    timeoutcmd.Command
	stdout StdoutWriter
	logger log.Logger
}

func New(name string, args []string, workDir string, envs, secrets []string, timeout, noOutputTimeout time.Duration, stdout io.Writer, logger log.Logger) Cmd {
	outWriter := NewStdoutWriter(secrets, stdout, logger)

	cmd := timeoutcmd.New(workDir, name, args...)
	cmd.SetTimeout(timeout)
	cmd.SetHangTimeout(noOutputTimeout)
	cmd.SetStandardIO(os.Stdin, outWriter, outWriter)
	cmd.SetEnv(append(envs, "PWD="+workDir))

	return Cmd{cmd: cmd, stdout: outWriter, logger: logger}
}

func (c *Cmd) Run() (int, error) {
	cmdErr := c.cmd.Start()

	if err := c.stdout.Close(); err != nil {
		c.logger.Warnf("Failed to close command output writer: %s", err)
	}

	if cmdErr == nil {
		return 0, nil
	}

	var exitErr *exec.ExitError
	if !errors.As(cmdErr, &exitErr) {
		return 1, fmt.Errorf("executing command failed: %w", cmdErr)
	}

	exitCode := exitErr.ExitCode()

	errorMessages := c.stdout.ErrorMessages()
	if len(errorMessages) > 0 {
		lastErrorMessage := errorMessages[len(errorMessages)-1]
		return exitCode, errors.New(lastErrorMessage)
	}

	return exitCode, exitErr
}
