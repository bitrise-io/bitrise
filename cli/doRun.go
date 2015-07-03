package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/bitrise-io/bitrise-cli/models"
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

func activateAndRunSteps(workFlow models.WorkFlowModel) error {
	for _, step := range workFlow.Steps {
		stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"

		err := bitrise.RunStepmanActivate(step.Id, step.VersionTag, stepDir)
		if err != nil {
			log.Errorln("Failed to run stepman activate")
			return err
		}

		log.Infof("Step activated: %s (%s)", step.Id, step.VersionTag)

		runStep(step)
	}
	return nil
}

func runStep(step models.StepModel) error {
	// Add step envs
	for _, input := range step.Inputs {
		if input.Value != nil {
			err := bitrise.RunPipedEnvmanAdd(*input.MappedTo, *input.Value)
			if err != nil {
				log.Errorln("Failed to run envman add")
				return err
			}
		}
	}

	stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"
	cmd := fmt.Sprintf("bash %sstep.sh", stepDir)
	err := bitrise.RunEnvmanRun(cmd)
	if err != nil {
		log.Errorln("Failed to run envman run")
		return err
	}

	log.Infof("Step executed: %s (%s)", step.Id, step.VersionTag)
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

	os.Setenv("ENVMAN_ENVSTORE_PATH", "/Users/godrei/develop/bitrise/bitrise-cli-test/envstore.yml")
	os.Setenv("BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH", "/Users/godrei/develop/bitrise/bitrise-cli-test/formout.md")
	err := bitrise.RunEnvmanInit()
	if err != nil {
		log.Error("Failed to run envman init")
		return
	}

	// Run work flow
	workFlow, err := bitrise.ReadWorkFlowJson(workFlowJsonPath)
	if err != nil {
		log.Errorln("Failed to read work flow:", err)
		return
	}

	err = activateAndRunSteps(workFlow)
	if err != nil {
		log.Errorln("Failed to activate steps:", err)
		return
	}
}
