package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func share(c *cli.Context) {
	if err := bitrise.StepmanShare(); err != nil {
		log.Fatalf("Bitrise share failed, err: %s", err)
	}
}
