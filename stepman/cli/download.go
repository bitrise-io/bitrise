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

	version := c.String(VersionKey)
	if version == "" {
		log.Fatal("[STEPMAN] - Missing step version")
	}

	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	exist, stepVersionData := collection.GetStep(id, version)
	if !exist {
		log.Fatalf("[STEPMAN] - Step: %s (v%s) failed to download from every avaiable download location.", id, version)
	}

	if err := stepman.DownloadStep(collection, stepVersionData); err != nil {
		log.Fatal("[STEPMAN] - Failed to download step")
	}
}
