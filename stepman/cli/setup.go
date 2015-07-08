package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	log.Info("[STEPMAN] - Setup")

	if err := stepman.SetupCurrentRouting(); err != nil {
		log.Error("[STEPMAN] - Failed to setup routing:", err)
		return
	}

	pth := stepman.GetCurrentStepCollectionPath()
	if err := stepman.DoGitClone(stepman.CollectionURI, pth); err != nil {
		log.Error("[STEPMAN] - Failed to get step spec path:", err)
		return
	}

	if err := stepman.WriteStepSpecToFile(); err != nil {
		log.Error("[STEPMAN] - Failed to save step spec:", err)
		return
	}

	log.Info("[STEPMAN] - Initialized")
}
