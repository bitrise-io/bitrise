package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func download(c *cli.Context) {
	log.Info("[STEPMAN] - Download")

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Fatal("[STEPMAN] - Missing step version")
	}

	collection, err := stepman.ReadStepSpec()
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	exist, step := collection.GetStep(id, version)
	if exist == false {
		log.Fatalf("[STEPMAN] - Step: %s - (%s) dos not exist", id, version)
	}

	if err := stepman.DownloadStep(collection, step); err != nil {
		log.Fatal("[STEPMAN] - Failed to download step")
	}
}
