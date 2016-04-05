package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func start(c *cli.Context) {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("No step collection specified")
	}

	if err := tools.StepmanShareStart(collectionURI); err != nil {
		log.Fatalf("Bitrise share start failed, err: %s", err)
	}
}
