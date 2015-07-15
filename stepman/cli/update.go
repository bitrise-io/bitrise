package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func update(c *cli.Context) {
	log.Info("[STEPMAN] - Update")

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if collectionURI == "" {
		log.Info("[STEPMAN] - No step collection specified, update all")
	}

	pth := stepman.GetStepCollectionPath(collectionURI)
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		log.Fatal("[STEPMAN] - Failed to check path:", err)
	} else if !exists {
		log.Fatal("[STEPMAN] - Not initialized")
	}

	if err := stepman.DoGitPull(pth); err != nil {
		log.Fatal("[STEPMAN] - Failed to do git update:", err)
	}

	specPth := pth + "steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	if err := stepman.WriteStepSpecToFile(collectionURI, collection); err != nil {
		log.Fatal("[STEPMAN] - Failed to save step spec:", err)
	}

	log.Info("[STEPMAN] - Updated")
}
