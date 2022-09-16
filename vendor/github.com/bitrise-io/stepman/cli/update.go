package cli

import (
	"fmt"

	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func update(c *cli.Context) error {
	collectionURIs := []string{}

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Info("No StepLib specified, update all...")
		collectionURIs = stepman.GetAllStepCollectionPath()
	} else {
		collectionURIs = []string{collectionURI}
	}

	if len(collectionURIs) == 0 {
		log.Info("No local StepLib found, nothing to update...")
	}

	for _, URI := range collectionURIs {
		log.Infof("Update StepLib (%s)...", URI)
		if _, err := stepman.UpdateLibrary(URI); err != nil {
			return fmt.Errorf("Failed to update StepLib (%s), error: %s", URI, err)
		}
	}

	return nil
}
