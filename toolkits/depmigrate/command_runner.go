package depmigrate

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

// CommandRunner ...
type CommandRunner interface {
	Run(c *command.Model) (string, error)
}

// DefaultRunner ...
type DefaultRunner struct {
}

// RunForOutput ...
func (r DefaultRunner) Run(c *command.Model) (string, error) {
	log.Debugf("$ %s", c.PrintableCommandArgs())

	out, err := c.RunAndReturnTrimmedCombinedOutput()
	if err != nil && errorutil.IsExitStatusError(err) {
		return out, fmt.Errorf("command `%s` failed, output: %s", c.PrintableCommandArgs(), out)
	}

	return out, fmt.Errorf("failed to run command `%s`: %v", err)
}
