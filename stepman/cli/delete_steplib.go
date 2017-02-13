package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func deleteStepLib(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		return fmt.Errorf("Missing required input: collection")
	}

	log.Infof("Delete StepLib: %s", collectionURI)

	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		log.Warnf("No route found for collection: %s, cleaning up routing..", collectionURI)
		if err := stepman.CleanupDanglingLibrary(collectionURI); err != nil {
			log.Errorf("Error cleaning up lib: %s", collectionURI)
		}
		log.Infof("Call 'stepman setup -c %s' for a clean setup", collectionURI)
		return nil
	}

	if err := stepman.CleanupRoute(route); err != nil {
		return fmt.Errorf("Failed to cleanup route for StepLib: %s", collectionURI)
	}

	return nil
}
