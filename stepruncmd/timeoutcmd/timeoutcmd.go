package timeoutcmd

import (
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/stepruncmd/hangdetector"
)

// Command controls the command run.
type Command struct {
	cmd          *exec.Cmd
	timeout      time.Duration
	hangTimeout  time.Duration
	hangDetector hangdetector.HangDetector
}

// New creates a command model.
func New(dir, name string, args ...string) Command {
	c := Command{
		cmd: exec.Command(name, args...),
	}
	c.cmd.Dir = dir

	return c
}

// SetTimeout sets the max runtime of the command.
func (c *Command) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SetHangTimeout sets the timeout after which the command is killed when no output is received on either stdout or stderr.
func (c *Command) SetHangTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.hangTimeout = timeout
		c.hangDetector = hangdetector.NewDefaultHangDetector(timeout)
	}
}

// SetEnv sets the command's env list.
func (c *Command) SetEnv(env []string) {
	c.cmd.Env = env
}

// SetStandardIO sets the input and outputs of the command.
func (c *Command) SetStandardIO(in io.Reader, out, err io.Writer) {
	if c.hangDetector == nil {
		c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = in, out, err
		return
	}

	c.cmd.Stdin = in
	c.cmd.Stdout = c.hangDetector.WrapOutWriter(out)
	c.cmd.Stderr = c.hangDetector.WrapErrWriter(err)
}

// Start starts the command run.
func (c *Command) Start() error {
	var hanged <-chan bool
	if c.hangDetector != nil {
		c.hangDetector.Start()
		defer c.hangDetector.Stop()
		hanged = c.hangDetector.C()
	}

	if err := c.cmd.Start(); err != nil { // start the process
		return err
	}

	// Wait reaps the process only. cmd.Wait would additionally block until the
	// goroutines exec.Cmd spawns to copy the child's stdout/stderr have drained
	// their pipes, which sounds better — it is what stops a large final burst
	// being truncated — but the pipes only reach EOF once every process holding
	// the write end exits. A step that leaves a daemon behind (ssh-agent, adb,
	// an emulator) keeps them open, and with no per-step timeout and no
	// BITRISE_NO_OUTPUT_TIMEOUT configured, which is the default, nothing else
	// in the select below can fire: the step would hang forever. Truncation is
	// the lesser failure, so this stays on Process.Wait.
	done := make(chan error, 1)
	go func() {
		switch p, err := c.cmd.Process.Wait(); {
		case err != nil:
			done <- err
		case p != nil:
			if !p.Success() {
				done <- &exec.ExitError{ProcessState: p}
			} else {
				done <- nil
			}
		}
	}()

	// or kill it after a timeout (whichever happens first)
	var timeoutChan <-chan time.Time
	if c.timeout > 0 {
		timeoutChan = time.After(c.timeout)
	}

	// exiting the method for the two supported cases: finish/error or timeout
	select {
	case <-timeoutChan:
		if err := c.cmd.Process.Kill(); err != nil {
			log.Warnf("Failed to kill process: %s", err)
		}

		return NewTimeoutError(c.timeout)
	case <-hanged:
		if err := c.cmd.Process.Kill(); err != nil {
			log.Warnf("Failed to kill process: %s", err)
		}

		return NewNoOutputTimeout(c.hangTimeout)
	case err := <-done:
		return err
	}
}

// ExitStatus returns the error's exit status
// if the error is an exec.ExitError
// if the error is nil it return 0
// otherwise returns 1.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}

	code := 1
	if exiterr, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			code = waitStatus.ExitStatus()
		}
	}
	return code
}
