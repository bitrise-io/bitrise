package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func list(c *cli.Context) {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	if exist, err := stepman.RootExistForCollection(collectionURI); err != nil {
		log.Fatal("[STEPMAN] - Failed to check routing:", err)
	} else if !exist {
		log.Fatalf("[STEPMAN] - Missing routing for collection, call 'stepman setup -c %s' before audit.", collectionURI)
	}

	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec (spec.json)")
	}

	// List
	for stepID := range collection.Steps {
		fmt.Println(stepID)
	}
}
