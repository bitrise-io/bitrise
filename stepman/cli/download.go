package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func download(c *cli.Context) {
	log.Info("[STEPMAN] - Download")

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Debug("[STEPMAN] - Missing step version -- Use latest version")

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatal("[STEPMAN] - Failed to get step latest version:", err)
		}
		log.Debug("[STEPMAN] - Latest version of step: %s", latest)
		version = latest
	}

	if err := stepman.DownloadStep(collection, id, version); err != nil {
		log.Fatal("[STEPMAN] - Failed to download step")
	}
}
