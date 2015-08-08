package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

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
		if err := stepman.ReGenerateStepSpec(URI); err != nil {
			log.Fatalf("Failed to update collection:%s error:%v", URI, err)
		}
	}

	log.Info("[STEPMAN] - Updated")
}
