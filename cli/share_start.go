package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func start(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		failf("No step collection specified")
	}

	if err := tools.StepmanShareStart(collectionURI); err != nil {
		failf("Bitrise share start failed, error: %s", err)
	}

	return nil
}
