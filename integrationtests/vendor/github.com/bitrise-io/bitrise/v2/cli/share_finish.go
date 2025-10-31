package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/urfave/cli"
)

func finish(c *cli.Context) error {
	logCommandParameters(c)

	if err := tools.StepmanShareFinish(); err != nil {
		failf("Bitrise share finish failed, error: %s", err)
	}

	return nil
}
