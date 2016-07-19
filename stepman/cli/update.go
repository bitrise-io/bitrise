package cli

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func updateCollection(steplibSource string) (models.StepCollectionModel, error) {
	route, found := stepman.ReadRoute(steplibSource)
	if !found {
		log.Warnf("No route found for collection: %s, cleaning up routing..", steplibSource)
		if err := stepman.CleanupDanglingLib(steplibSource); err != nil {
			log.Errorf("Error cleaning up lib: %s", steplibSource)
		}
		log.Infof("Call 'stepman setup -c %s' for a clean setup", steplibSource)
		return models.StepCollectionModel{}, fmt.Errorf("No route found for StepLib: %s", steplibSource)
	}

	isLocalSteplib := strings.HasPrefix(steplibSource, "file://")

	if isLocalSteplib {
		if err := stepman.CleanupRoute(route); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("Failed to cleanup route for StepLib: %s", steplibSource)
		}

		if err := setupSteplib(steplibSource, false); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("Failed to setup StepLib: %s", steplibSource)
		}
	} else {
		pth := stepman.GetCollectionBaseDirPath(route)
		if exists, err := pathutil.IsPathExists(pth); err != nil {
			return models.StepCollectionModel{}, err
		} else if !exists {
			return models.StepCollectionModel{}, errors.New("Not initialized")
		}

		if err := cmdex.GitPull(pth); err != nil {
			return models.StepCollectionModel{}, err
		}

		if err := stepman.ReGenerateStepSpec(route); err != nil {
			return models.StepCollectionModel{}, err
		}
	}

	return stepman.ReadStepSpec(steplibSource)
}

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
		if _, err := updateCollection(URI); err != nil {
			return fmt.Errorf("Failed to update StepLib (%s), error: %s", URI, err)
		}
	}

	return nil
}
