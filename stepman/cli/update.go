package cli

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func updateCollection(steplibSource string) (models.StepCollectionModel, error) {
	route, found := stepman.ReadRoute(steplibSource)
	if !found {
		return models.StepCollectionModel{},
			fmt.Errorf("No collection found for lib, call 'stepman delete -c %s' for cleanup", steplibSource)
	}

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

	return stepman.ReadStepSpec(steplibSource)
}

func update(c *cli.Context) {
	log.Info("[STEPMAN] - Update")

	collectionURIs := []string{}

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Info("[STEPMAN] - No step collection specified, update all")
		collectionURIs = stepman.GetAllStepCollectionPath()
	} else {
		collectionURIs = []string{collectionURI}
	}

	for _, URI := range collectionURIs {
		if _, err := updateCollection(URI); err != nil {
			log.Fatalf("Failed to update collection (%s), err: %s", collectionURI, err)
		}
	}

	log.Info("[STEPMAN] - Updated")
}
