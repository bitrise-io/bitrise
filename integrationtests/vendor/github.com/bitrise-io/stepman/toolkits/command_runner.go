package toolkits

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/stepman/stepman"
)

type commandRunner interface {
	runForOutput(c *command.Model) (string, error)
}

type defaultRunner struct {
	logger stepman.Logger
}

func newDefaultRunner(logger stepman.Logger) *defaultRunner {
	return &defaultRunner{logger: logger}
}

func (r *defaultRunner) runForOutput(c *command.Model) (string, error) {
	r.logger.Debugf("$ %s", c.PrintableCommandArgs())

	out, err := c.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return out, fmt.Errorf("command `%s` failed, output: %s", c.PrintableCommandArgs(), out)
		}

		return out, fmt.Errorf("failed to run command `%s`: %v", c.PrintableCommandArgs(), err)
	}

	return out, nil
}
