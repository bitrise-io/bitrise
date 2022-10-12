package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func share(c *cli.Context) error {
	if err := tools.StepmanShare(); err != nil {
		failf("Bitrise share failed, error: %s", err)
	}

	return nil
}
