package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func share(c *cli.Context) {
	if err := tools.StepmanShare(); err != nil {
		log.Fatalf("Bitrise share failed, err: %s", err)
	}
}
