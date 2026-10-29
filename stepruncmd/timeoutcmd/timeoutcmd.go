package timeoutcmd

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/bitrise-io/bitrise/v3/stepruncmd/hangdetector"
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

	output, err := copyOutput(c.cmd)
	if err != nil {
		return err
	}

	startErr := c.cmd.Start() // start the process
	// The child holds its own copies of the write ends; closing ours lets the
	// copies reach EOF once the child, and anything it left behind, exits.
	output.closeWriteEnds()
	if startErr != nil {
		return startErr
	}

	// Wait reaps the process only. cmd.Wait would additionally block until the
	// child's stdout/stderr pipes reach EOF, but a step that leaves a daemon
	// behind (ssh-agent, adb, an emulator) keeps them open, and with no per-step
	// timeout and no BITRISE_NO_OUTPUT_TIMEOUT configured, which is the default,
	// nothing else in the select below can fire: the step would hang forever.
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
		// The copies can still be behind the child's last writes when it exits,
		// which would lose the end of the step's output and the error message
		// read from it. Wait for them, but only briefly: a daemon left behind
		// keeps the pipes open, and its output keeps being copied after this
		// returns, so it isn't killed by writing to a closed pipe.
		select {
		case <-output.done:
		case <-time.After(outputDrainTimeout):
		}
		return err
	}
}

// outputDrainTimeout bounds how long Start waits, once the process exits, for
// its remaining output to be copied.
const outputDrainTimeout = time.Second

// commandOutput copies a command's stdout and stderr through pipes it owns,
// instead of the ones exec.Cmd creates, so that waiting for the copy doesn't
// mean cmd.Wait, which waits for EOF or, given a WaitDelay, closes the pipes.
type commandOutput struct {
	writeEnds []*os.File
	done      chan struct{}
}

// copyOutput replaces cmd's stdout and stderr writers with pipes, and copies
// from them to the original writers. Like exec.Cmd, it uses one pipe for both
// when they are the same writer, which keeps their lines in order.
func copyOutput(cmd *exec.Cmd) (commandOutput, error) {
	dests := []io.Writer{cmd.Stdout}
	if !sameWriter(cmd.Stdout, cmd.Stderr) {
		dests = append(dests, cmd.Stderr)
	}

	output := commandOutput{done: make(chan struct{})}
	var copies sync.WaitGroup
	for _, dest := range dests {
		r, w, err := os.Pipe()
		if err != nil {
			output.closeWriteEnds()
			return commandOutput{}, err
		}
		output.writeEnds = append(output.writeEnds, w)

		copies.Go(func() {
			defer r.Close()
			_, _ = io.Copy(dest, r)
		})
	}
	go func() {
		copies.Wait()
		close(output.done)
	}()

	cmd.Stdout = output.writeEnds[0]
	cmd.Stderr = output.writeEnds[len(output.writeEnds)-1]
	return output, nil
}

func (o commandOutput) closeWriteEnds() {
	for _, w := range o.writeEnds {
		_ = w.Close()
	}
}

// sameWriter reports whether a and b are the same writer, the way exec.Cmd
// decides to share one pipe: a writer of an uncomparable type is never equal.
func sameWriter(a, b io.Writer) (same bool) {
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
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
