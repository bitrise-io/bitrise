package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func shareAudit(c *cli.Context) {
	if err := bitrise.StepmanShareAudit(); err != nil {
		log.Fatalf("Bitrise share audit failed, err: %s", err)
	}
}
