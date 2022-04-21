package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func start(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatal("No step collection specified")
	}

	if err := tools.StepmanShareStart(collectionURI); err != nil {
		log.Fatalf("Bitrise share start failed, error: %s", err)
	}

	return nil
}
