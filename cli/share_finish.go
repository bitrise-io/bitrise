package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func finish(c *cli.Context) {
	if err := bitrise.StepmanShareFinish(); err != nil {
		log.Fatalf("Bitrise share finish failed, err: %s", err)
	}
}
