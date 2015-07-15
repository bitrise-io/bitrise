package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	log.Debugln("[STEPMAN] - Setup")

	if exist, err := stepman.RootExistForCurrentCollection(); err != nil {
		log.Fatal("[STEPMAN] - Failed to check routing:", err)
	} else if exist {
		log.Debugln("[STEPMAN] - Nothing to setup, everything's ready.")
		return
	}

	if err := stepman.SetupCurrentRouting(); err != nil {
		log.Fatal("[STEPMAN] - Failed to setup routing:", err)
	}

	pth := stepman.GetCurrentStepCollectionPath()
	if err := stepman.DoGitClone(stepman.CollectionURI, pth); err != nil {
		log.Fatal("[STEPMAN] - Failed to get step spec path:", err)
	}

	specPth := pth + "steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	if err := stepman.WriteStepSpecToFile(collection); err != nil {
		log.Fatal("[STEPMAN] - Failed to save step spec:", err)
	}

	log.Info("[STEPMAN] - Initialized")
}
