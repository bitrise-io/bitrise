package toolkits

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

// commandRunner ...
type commandRunner interface {
	run(c *command.Model) error
	runForOutput(c *command.Model) (string, error)
}

// defaultRunner ...
type defaultRunner struct {
}

// run ...
func (r *defaultRunner) run(c *command.Model) error {
	log.Debugf("$ %s", c.PrintableCommandArgs())

	if err := c.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("command `%s` failed: %v", c.PrintableCommandArgs(), err)
		}

		return fmt.Errorf("failed to run command `%s`: %v", c.PrintableCommandArgs(), err)
	}

	return nil
}

// runForOutput ...
func (r *defaultRunner) runForOutput(c *command.Model) (string, error) {
	log.Debugf("$ %s", c.PrintableCommandArgs())

	out, err := c.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return out, fmt.Errorf("command `%s` failed, output: %s", c.PrintableCommandArgs(), out)
		}

		return out, fmt.Errorf("failed to run command `%s`: %v", c.PrintableCommandArgs(), err)
	}

	return out, nil
}
