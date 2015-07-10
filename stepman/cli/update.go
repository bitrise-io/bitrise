package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func update(c *cli.Context) {
	log.Info("[STEPMAN] - Update")

	pth := stepman.GetCurrentStepCollectionPath()
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return
	} else if exists == false {
		log.Error("[STEPMAN] - Not initialized")
		return
	}

	if err := stepman.DoGitPull(pth); err != nil {
		log.Error("[STEPMAN] - Failed to do git update:", err)
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

	log.Info("[STEPMAN] - Updated")
}
