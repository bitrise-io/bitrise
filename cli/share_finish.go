package cli

import (
	"log"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func finish(c *cli.Context) error {
	if err := tools.StepmanShareFinish(); err != nil {
		log.Fatalf("Bitrise share finish failed, error: %s", err)
	}

	return nil
}
