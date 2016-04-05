package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func shareAudit(c *cli.Context) {
	if err := tools.StepmanShareAudit(); err != nil {
		log.Fatalf("Bitrise share audit failed, err: %s", err)
	}
}
