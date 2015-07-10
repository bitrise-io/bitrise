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
		log.Error("[STEPMAN] - Missing step id")
		return
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Error("[STEPMAN] - Missing step version")
		return
	}

	collection, err := stepman.ReadStepSpec()
	if err != nil {
		log.Error("[STEPMAN] - Failed to read step spec:", err)
		return
	}

	exist, step := collection.GetStep(id, version)
	if exist == false {
		log.Errorf("[STEPMAN] - Step: %s - (%s) dos not exist", id, version)
		return
	}

	if err := stepman.DownloadStep(collection, step); err != nil {
		log.Error("[STEPMAN] - Failed to download step")
	}
}
