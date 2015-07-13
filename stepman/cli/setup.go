package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	log.Info("[STEPMAN] - Setup")

	if exist, err := stepman.RootExistForCurrentCollection(); err != nil {
		log.Error("[STEPMAN] - Failed to check routing:", err)
	} else if exist {
		return
	}

	if err := stepman.SetupCurrentRouting(); err != nil {
		log.Error("[STEPMAN] - Failed to setup routing:", err)
		return
	}

	pth := stepman.GetCurrentStepCollectionPath()
	if err := stepman.DoGitClone(stepman.CollectionURI, pth); err != nil {
		log.Error("[STEPMAN] - Failed to get step spec path:", err)
		return
	}

	specPth := pth + "steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		log.Error("[STEPMAN] - Failed to read step spec:", err)
		return
	}

	if err := stepman.WriteStepSpecToFile(collection); err != nil {
		log.Error("[STEPMAN] - Failed to save step spec:", err)
		return
	}

	log.Info("[STEPMAN] - Initialized")
}
