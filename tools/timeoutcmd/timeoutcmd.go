package timeoutcmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

// Command controls the command run.
type Command struct {
	cmd     *exec.Cmd
	timeout time.Duration
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
func (c *Command) SetTimeout(t time.Duration) {
	c.timeout = t
}

// AppendEnv appends and env to the command's env list.
func (c *Command) AppendEnv(env string) {
	if c.cmd.Env != nil {
		c.cmd.Env = append(c.cmd.Env, env)
		return
	}
	c.cmd.Env = append(os.Environ(), env)
}

// SetStandardIO sets the input and outputs of the command.
func (c *Command) SetStandardIO(in io.Reader, out, err io.Writer) {
	c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr = in, out, err
}

// Start starts the command run.
func (c *Command) Start() error {
	// setting up notification for signals so we can have
	// separated logic to end the process
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, os.Kill)
	var interrupted bool
	go func() {
		<-interruptChan
		interrupted = true
	}()

	// start the process
	if err := c.cmd.Start(); err != nil {
		return err
	}

	// Wait for the process to finish
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
		return fmt.Errorf("timed out")
	case err := <-done:
		if interrupted {
			os.Exit(ExitStatus(err))
		}
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
