package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func update(c *cli.Context) error {
	collectionURIs := []string{}

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Infof("No StepLib specified, update all...")
		collectionURIs = stepman.GetAllStepCollectionPath()
	} else {
		collectionURIs = []string{collectionURI}
	}

	if len(collectionURIs) == 0 {
		log.Infof("No local StepLib found, nothing to update...")
	}

	for _, URI := range collectionURIs {
		log.Infof("Update StepLib (%s)...", URI)
		if err := UpdateLibrary(URI, log.NewDefaultLogger(false)); err != nil {
			return fmt.Errorf("Failed to update StepLib (%s), error: %s", URI, err)
		}
	}

	return nil
}

// UpdateLibrary ...
func UpdateLibrary(uri string, log stepman.Logger) error {
	_, err := stepman.UpdateLibrary(uri, log)
	return err
}
