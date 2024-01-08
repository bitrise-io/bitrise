package command

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/v2/env"
)

// ErrorFinder ...
type ErrorFinder func(out string) []string

// Opts ...
type Opts struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
	Env         []string
	Dir         string
	ErrorFinder ErrorFinder
}

// Factory ...
type Factory interface {
	Create(name string, args []string, opts *Opts) Command
}

type factory struct {
	envRepository env.Repository
}

// NewFactory ...
func NewFactory(envRepository env.Repository) Factory {
	return factory{envRepository: envRepository}
}

// Create ...
func (f factory) Create(name string, args []string, opts *Opts) Command {
	cmd := exec.Command(name, args...)
	var collector *errorCollector

	if opts != nil {
		if opts.ErrorFinder != nil {
			collector = &errorCollector{errorFinder: opts.ErrorFinder}
		}

		cmd.Stdout = opts.Stdout
		cmd.Stderr = opts.Stderr
		cmd.Stdin = opts.Stdin

		// If Env is nil, the new process uses the current process's
		// environment.
		// If we pass env vars we want to append them to the
		// current process's environment.
		cmd.Env = append(f.envRepository.List(), opts.Env...)
		cmd.Dir = opts.Dir
	}
	return &command{
		cmd:            cmd,
		errorCollector: collector,
	}
}

// Command ...
type Command interface {
	PrintableCommandArgs() string
	Run() error
	RunAndReturnExitCode() (int, error)
	RunAndReturnTrimmedOutput() (string, error)
	RunAndReturnTrimmedCombinedOutput() (string, error)
	Start() error
	Wait() error
}

type command struct {
	cmd            *exec.Cmd
	errorCollector *errorCollector
}

// PrintableCommandArgs ...
func (c command) PrintableCommandArgs() string {
	return printableCommandArgs(false, c.cmd.Args)
}

// Run ...
func (c *command) Run() error {
	c.wrapOutputs()

	if err := c.cmd.Run(); err != nil {
		return c.wrapError(err)
	}

	return nil
}

// RunAndReturnExitCode ...
func (c command) RunAndReturnExitCode() (int, error) {
	c.wrapOutputs()
	err := c.cmd.Run()
	if err != nil {
		err = c.wrapError(err)
	}

	exitCode := c.cmd.ProcessState.ExitCode()
	return exitCode, err
}

// RunAndReturnTrimmedOutput ...
func (c command) RunAndReturnTrimmedOutput() (string, error) {
	outBytes, err := c.cmd.Output()
	outStr := string(outBytes)
	if err != nil {
		if c.errorCollector != nil {
			c.errorCollector.collectErrors(outStr)
		}
		err = c.wrapError(err)
	}

	return strings.TrimSpace(outStr), err
}

// RunAndReturnTrimmedCombinedOutput ...
func (c command) RunAndReturnTrimmedCombinedOutput() (string, error) {
	outBytes, err := c.cmd.CombinedOutput()
	outStr := string(outBytes)
	if err != nil {
		if c.errorCollector != nil {
			c.errorCollector.collectErrors(outStr)
		}
		err = c.wrapError(err)
	}

	return strings.TrimSpace(outStr), err
}

// Start ...
func (c command) Start() error {
	c.wrapOutputs()
	return c.cmd.Start()
}

// Wait ...
func (c command) Wait() error {
	err := c.cmd.Wait()
	if err != nil {
		err = c.wrapError(err)
	}

	return err
}

func printableCommandArgs(isQuoteFirst bool, fullCommandArgs []string) string {
	var cmdArgsDecorated []string
	for idx, anArg := range fullCommandArgs {
		quotedArg := strconv.Quote(anArg)
		if idx == 0 && !isQuoteFirst {
			quotedArg = anArg
		}
		cmdArgsDecorated = append(cmdArgsDecorated, quotedArg)
	}

	return strings.Join(cmdArgsDecorated, " ")
}

func (c command) wrapError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if c.errorCollector != nil && len(c.errorCollector.errorLines) > 0 {
			return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), c.PrintableCommandArgs(), errors.New(strings.Join(c.errorCollector.errorLines, "\n")))
		}
		return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), c.PrintableCommandArgs(), errors.New("check the command's output for details"))
	}
	return fmt.Errorf("executing command failed (%s): %w", c.PrintableCommandArgs(), err)
}

func (c command) wrapOutputs() {
	if c.errorCollector == nil {
		return
	}

	if c.cmd.Stdout != nil {
		outWriter := io.MultiWriter(c.errorCollector, c.cmd.Stdout)
		c.cmd.Stdout = outWriter
	} else {
		c.cmd.Stdout = c.errorCollector
	}

	if c.cmd.Stderr != nil {
		errWriter := io.MultiWriter(c.errorCollector, c.cmd.Stderr)
		c.cmd.Stderr = errWriter
	} else {
		c.cmd.Stderr = c.errorCollector
	}
}
