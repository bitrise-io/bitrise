package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func update(c *cli.Context) {
	log.Info("[STEPMAN] - Update")

	stepCollectionPath := stepman.GetCurrentStepCollectionPath()
	if exists, err := pathutil.IsPathExists(stepCollectionPath); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return
	} else if exists == false {
		log.Error("[STEPMAN] - Not initialized")
		return
	}

	if err := stepman.DoGitPull(stepCollectionPath); err != nil {
		log.Error("[STEPMAN] - Failed to do git update:", err)
		return
	}

	if err := stepman.WriteStepSpecToFile(); err != nil {
		log.Error("[STEPMAN] - Failed to save step spec:", err)
		return
	}

	log.Info("[STEPMAN] - Updated")
}
