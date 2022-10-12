package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func shareAudit(c *cli.Context) error {
	if err := tools.StepmanShareAudit(); err != nil {
		failf("Bitrise share audit failed, error: %s", err)
	}

	return nil
}
