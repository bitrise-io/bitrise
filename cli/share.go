package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func share(c *cli.Context) error {
	return tools.StepmanShare()
}
