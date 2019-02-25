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
	cmd           *exec.Cmd
	timeout       time.Duration
	timeoutTimer  *time.Timer
	interruptChan chan os.Signal
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
	// Setpgid: true creates a new process group for cmd and its subprocesses
	// this way cmd will not belong to its parent process group,
	// cmd will not be killed when you hit ^C in your terminal
	// to fix this, we listen and handle Interrupt signal manually
	c.interruptChan = make(chan os.Signal)
	signal.Notify(c.interruptChan, os.Interrupt, os.Kill)
	c.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	if c.timeout > 0 {
		// terminate the process after the given timeout
		c.timeoutTimer = time.AfterFunc(c.timeout, func() {
			if err := c.Stop(os.Kill); err != nil {
				log.Warnf("Failed to kill process, error: %s", err)
			}
		})
	}

	var interrupted bool

	go func() {
		for {
			sig := <-c.interruptChan
			interrupted = true

			if err := c.Stop(sig); err != nil {
				log.Warnf("Failed to stop process, error: %s", err)
			}
		}
	}()

	if err := c.cmd.Wait(); err != nil {
		if interrupted {
			os.Exit(ExitStatus(err))
		}
		return err
	}
	return nil
}

// Stop terminates the command run.
func (c *Command) Stop(sig os.Signal) error {
	if c.cmd.Process == nil {
		// not yet started or there was an error starting the command
		return nil
	}

	if c.timeout > 0 {
		// stop the timeout timer if it was set
		c.timeoutTimer.Stop()
	}

	signum := syscall.SIGKILL
	if sig == os.Interrupt {
		signum = syscall.SIGINT
	}

	// all child processes will be included in the current process' group and
	// this group ID will be the original PID of the child because of Setpgid
	// - is to kill the whole group not the only children
	return syscall.Kill(-c.cmd.Process.Pid, signum)
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
