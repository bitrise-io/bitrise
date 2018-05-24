package timeoutcmd

import (
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
	cmd *exec.Cmd

	timeout       time.Duration
	timeoutTimer  *time.Timer
	interruptChan chan os.Signal
}

// New creates a command model.
func New(dir, name string, args ...string) Command {
	c := Command{}

	c.cmd = exec.Command(name, args...)
	c.cmd.Dir = dir

	return c
}

// SetTimeout sets the max runtime of the command.
func (c *Command) SetTimeout(t time.Duration) {
	c.timeout = t
}

// SetStandardIO sets the input and outputs of the command.
func (c *Command) SetStandardIO(in io.Reader, out, err io.Writer) {
	c.cmd.Stdin = in
	c.cmd.Stdout = out
	c.cmd.Stderr = err
}

// Start starts the command run.
func (c *Command) Start() error {
	if c.timeout > 0 {
		// Setpgid: true creates a new process group for cmd and its subprocesses
		// this way we can kill the whole process group
		c.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	if c.timeout > 0 {
		// terminate the process after the given timeout
		c.timeoutTimer = time.AfterFunc(c.timeout, func() {
			if err := c.Stop(); err != nil {
				log.Warnf("Failed to kill the process, error: %s", err)
			}
		})

		// Setpgid: true creates a new process group for cmd and its subprocesses
		// this way cmd will not belong to its parent process group,
		// cmd will not be killed when you hit ^C in your terminal
		// to fix this, we listen and handle Interrupt signal manually
		c.interruptChan = make(chan os.Signal, 1)
		signal.Notify(c.interruptChan, os.Interrupt)
		go func() {
			<-c.interruptChan
			signal.Stop(c.interruptChan)
			if err := c.Stop(); err != nil {
				log.Warnf("Failed to kill the process, error: %s", err)
			}
		}()
	}

	return c.cmd.Wait()
}

// Stop terminates the command run.
func (c *Command) Stop() error {
	if c.cmd.Process == nil {
		// not yet started
		return nil
	}

	pid := c.cmd.Process.Pid
	if c.timeout > 0 {
		// stop listening on os.Interrupt signal
		signal.Stop(c.interruptChan)
		// stop the timeout timer
		c.timeoutTimer.Stop()

		// use the negative process group id, to kill the whole process group
		pgid, err := syscall.Getpgid(c.cmd.Process.Pid)
		if err != nil {
			return err
		}
		pid = -1 * pgid
	}

	// kill the process
	return syscall.Kill(pid, syscall.SIGKILL)
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
