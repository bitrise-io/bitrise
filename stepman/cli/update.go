package cli

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func updateCollection(collectionURI string) error {
	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		return errors.New("No route found for lib: " + collectionURI)
	}

	pth := stepman.GetCollectionBaseDirPath(route)
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exists {
		return errors.New("[STEPMAN] - Not initialized")
	}

	if err := stepman.DoGitPull(pth); err != nil {
		return err
	}

	specPth := pth + "/steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		return err
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		return err
	}
	return nil
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
		if err := updateCollection(URI); err != nil {
			log.Fatalf("Failed to update collection:%s error:%v", URI, err)
		}
	}

	log.Info("[STEPMAN] - Updated")
}
