package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/urfave/cli"
)

func shareAudit(c *cli.Context) error {
	logCommandParameters(c)

	if err := tools.StepmanShareAudit(); err != nil {
		failf("Bitrise share audit failed, error: %s", err)
	}

	return nil
}
