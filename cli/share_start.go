package cli

import (
	"errors"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func start(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		return errors.New("No step collection specified")
	}

	return tools.StepmanShareStart(collectionURI)
}
