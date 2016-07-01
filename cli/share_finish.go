package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func finish(c *cli.Context) error {
	return tools.StepmanShareFinish()
}
