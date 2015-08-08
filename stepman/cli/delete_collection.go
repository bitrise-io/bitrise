package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func deleteCollection(c *cli.Context) {
	log.Debugln("[STEPMAN] - Delete collection")

	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		log.Warn("No route found for lib: " + collectionURI)
		return
	}

	if err := stepman.CleanupRoute(route); err != nil {
		log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
	}
}
