package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func shareAudit(c *cli.Context) {
	if err := tools.StepmanShareAudit(); err != nil {
		log.Fatalf("Bitrise share audit failed, error: %s", err)
	}
}
