package asynccmd

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

// Status ...
type Status struct {
	Code int
	Err  error
}

// Cmd ...
type Cmd struct {
	name string
	args []string

	dir     string
	secrets []string
	timeout time.Duration

	pid int
}

// New ...
func New(name string, args ...string) *Cmd {
	return &Cmd{name: name, args: args}
}

// SetDir ...
func (c *Cmd) SetDir(dir string) {
	c.dir = dir
}

// SetSecrets ...
func (c *Cmd) SetSecrets(secrets []string) {
	c.secrets = append([]string{}, secrets...)
}

// SetTimeout ...
func (c *Cmd) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Start starts the command asynchronous
// and returns a status chanel to observe the command's status
// and returns a log chanel to fetch the command's log.
func (c *Cmd) Start() (chan Status, chan string) {
	statusChan := make(chan Status)
	logChan := make(chan string)

	go func() {
		cmd := exec.Command(c.name, c.args...)
		cmd.Dir = c.dir

		if c.timeout > 0 {
			// Setpgid: true creates a new process group for cmd and its subprocesses
			// this way we can kill the whole process group
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		}

		combinedOut := newBuffer(c.secrets)
		cmd.Stdout = combinedOut
		cmd.Stderr = combinedOut
		cmd.Stdin = os.Stdin

		if err := cmd.Start(); err != nil {
			statusChan <- Status{Code: exitStatus(err), Err: err}
			return
		}

		c.pid = cmd.Process.Pid

		if c.timeout > 0 {
			// Setpgid: true creates a new process group for cmd and its subprocesses
			// this way cmd will not belong to its parent process group,
			// cmd will not be killed when you hit ^C in your terminal
			// to fix this, we listen and handle Interrupt signal manually
			c.listenOnInterrupt()
		}

		// Check Timeout
		var timer *time.Timer
		if c.timeout > 0 {
			timer = time.AfterFunc(c.timeout, func() {
				if err := c.Stop(); err != nil {
					log.Warnf("Failed to kill process, error: %s", err)
				}
			})
		}

		// Fetch combinedOut periodically
		ticker := time.NewTicker(time.Millisecond * 500)
		go func() {
			for range ticker.C {
				lines, err := combinedOut.ReadLines()
				if err != nil {
					statusChan <- Status{Code: 1, Err: err}
					return
				}

				if len(lines) == 0 && time.Since(combinedOut.lastWrite) > 2*time.Second {
					// may the command waits for user input
					if err := combinedOut.Flush(); err != nil {
						statusChan <- Status{Code: 1, Err: err}
						return
					}
				} else {
					for _, line := range lines {
						logChan <- line
					}
				}
			}
		}()

		exitErr := cmd.Wait()

		// Stop timers
		if timer != nil {
			timer.Stop()
		}
		ticker.Stop()

		// Fetch remaining lines from combinedOut
		if err := combinedOut.Flush(); err != nil {
			statusChan <- Status{Code: 1, Err: err}
			return
		}

		lines, err := combinedOut.ReadLines()
		if err != nil {
			statusChan <- Status{Code: 1, Err: err}
			return
		}
		for _, line := range lines {
			logChan <- line
		}

		// notify the receiver that no more logs will be sent
		close(logChan)

		statusChan <- Status{Code: exitStatus(exitErr), Err: exitErr}
	}()

	return statusChan, logChan
}

// Stop stops the command.
func (c *Cmd) Stop() error {
	if c.timeout > 0 {
		pgid, err := syscall.Getpgid(c.pid)
		if err != nil {
			return err
		}

		return syscall.Kill(-pgid, syscall.SIGKILL)
	}
	return syscall.Kill(c.pid, syscall.SIGKILL)
}

// listenOnInterrupt listens on Interrupt signal.
func (c *Cmd) listenOnInterrupt() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		signal.Stop(signalChan)
		if err := c.Stop(); err != nil {
			log.Warnf("Failed to kill process, error: %s", err)
		}
	}()
}

// exitStatus returns the error's exit status
// if the error is an exec.ExitError
// if the error is nil it return 0
// otherwise returns 1.
func exitStatus(err error) (code int) {
	if err == nil {
		return
	}

	code = 1
	if exiterr, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			code = waitStatus.ExitStatus()
		}
	}
	return
}
