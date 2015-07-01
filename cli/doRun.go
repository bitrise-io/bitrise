package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/codegangsta/cli"
)

func getWorkFlowPathInCurrentFolder() (string, error) {
	fileInfos, err := ioutil.ReadDir("./")
	if err != nil {
		return "", err
	}

	matches := 0
	workFlowName := ""
	for _, fileInfo := range fileInfos {
		match, e := regexp.MatchString("([a-z]+).json", fileInfo.Name())
		if e != nil {
			return "", err
		}
		if match {
			matches = matches + 1
			workFlowName = fileInfo.Name()
		}
	}

	if matches == 0 {
		return "", errors.New("No workflow json found")
	}
	if matches > 1 {
		return "", errors.New("More then one possible workflow json found")
	}

	return workFlowName, nil
}

func activateSteps(workFlow bitrise.WorkFlowJsonStruct) error {
	for _, step := range workFlow.Steps {
		stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"
		activateCmd := fmt.Sprintf("stepman activate -i %s -v %s -p %s", step.Id, step.VersionTag, stepDir)
		err := bitrise.RunBashComman(activateCmd)
		if err != nil {
			log.Errorln("Failed to execute cmd:", activateCmd)
			return err
		}

		log.Infof("Step activated: %s (%s)", step.Id, step.VersionTag)

		stepCmd := "step.sh"
		err = bitrise.RunBashCommanInDir(stepDir, stepCmd)
		if err != nil {
			log.Errorln("Failed to execute cmd:", stepCmd)
			return err
		}

		log.Infof("Step executed: %s (%s)", step.Id, step.VersionTag)
	}
	return nil
}

func doRun(c *cli.Context) {
	log.Info("Run")

	// Input validation
	workFlowJsonPath := c.String(PATH_KEY)
	if workFlowJsonPath == "" {
		log.Infoln("Workflow json path not defined, try search in current folder")

		workFlowName, err := getWorkFlowPathInCurrentFolder()
		if err != nil {
			log.Errorln("Failed to find workflow json:", err)
			return
		}
		workFlowJsonPath = "./" + workFlowName
	}

	workFlow, err := bitrise.ReadWorkFlowJson(workFlowJsonPath)
	if err != nil {
		log.Errorln("Failed to read work flow:", err)
		return
	}

	err = activateSteps(workFlow)
	if err != nil {
		log.Errorln("Failed to activate steps:", err)
		return
	}
}
