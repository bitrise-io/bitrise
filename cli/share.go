package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func share(c *cli.Context) error {
	if err := tools.StepmanShare(); err != nil {
		log.Fatalf("Bitrise share failed, error: %s", err)
	}

	return nil
}
