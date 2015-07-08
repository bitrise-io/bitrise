package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func activate(c *cli.Context) {
	log.Info("[STEPMAN] - Activate")

	// Input validation
	id := c.String(ID_KEY)
	if id == "" {
		log.Error("[STEPMAN] - Missing step id")
		return
	}

	version := c.String(VERSION_KEY)
	if version == "" {
		log.Error("[STEPMAN] - Missing step version")
		return
	}

	path := c.String(PATH_KEY)
	if path == "" {
		log.Error("[STEPMAN] - Missing destination path")
		return
	}

	// Get step
	stepCollection, err := stepman.ReadStepSpec()
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec")
		return
	}

	exist, step := stepCollection.GetStep(id, version)
	if exist == false {
		log.Errorf("[STEPMAN] - Step: %s - (%s) dos not exist", id, version)
		return
	}

	pth := stepman.GetStepPath(step)
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return
	} else if exist == false {
		log.Info("[STEPMAN] - Step dos not exist, download it")
		if err := stepman.DownloadStep(step); err != nil {
			log.Error("[STEPMAN] - Failed to download step:", err)
		}
	}

	// Copy to specified path
	srcFolder := pth
	destFolder := path

	if exist, err = pathutil.IsPathExists(destFolder); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return
	} else if exist == false {
		if err := os.MkdirAll(destFolder, 0777); err != nil {
			log.Error("[STEPMAN] - Failed to create path:", err)
			return
		}
	}

	if err = stepman.RunCommand("cp", []string{"-rf", srcFolder, destFolder}...); err != nil {
		log.Error("[STEPMAN] - Failed to copy step:", err)
	}
}
