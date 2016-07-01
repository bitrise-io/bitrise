package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func shareAudit(c *cli.Context) error {
	return tools.StepmanShareAudit()
}
