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
		if match, err := regexp.MatchString("([a-z]+).json", fileInfo.Name()); err != nil {
			return "", err
		} else if match {
			matches = matches + 1
			workFlowName = fileInfo.Name()
		}
	}

	if matches == 0 {
		return "", errors.New("No workflow json found")
	} else if matches > 1 {
		return "", errors.New("More then one possible workflow json found")
	}
	return workFlowName, nil
}

func activateAndRunSteps(workFlow models.WorkFlowModel) error {
	for _, step := range workFlow.Steps {
		stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"

		if err := bitrise.RunStepmanActivate(step.Id, step.VersionTag, stepDir); err != nil {
			log.Errorln("Failed to run stepman activate")
			return err
		}

		log.Infof("Step activated: %s (%s)", step.Id, step.VersionTag)

		if err := runStep(step); err != nil {
			log.Errorln("Failed to run step")
			return err
		}
	}
	return nil
}

func runStep(step models.StepModel) error {
	// Add step envs
	for _, input := range step.Inputs {
		if input.Value != nil {
			if err := bitrise.RunEnvmanAdd(*input.MappedTo, *input.Value); err != nil {
				log.Errorln("Failed to run envman add")
				return err
			}
		}
	}

	stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"
	stepCmd := fmt.Sprintf("%sstep.sh", stepDir)
	cmd := []string{"bash", stepCmd}

	if err := bitrise.RunEnvmanRun(cmd); err != nil {
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

		if workFlowName, err := getWorkFlowPathInCurrentFolder(); err != nil {
			log.Errorln("Failed to find workflow json:", err)
			return
		} else {
			workFlowJsonPath = "./" + workFlowName
		}
	}

	// Envman setup
	if err := os.Setenv(ENVSTORE_PATH_ENV_KEY, ENVSTORE_PATH); err != nil {
		log.Errorln("Failed to add env:", err)
		return
	}

	if err := os.Setenv(FORMATTED_OUTPUT_PATH_ENV_KEY, FORMATTED_OUTPUT_PATH); err != nil {
		log.Errorln("Failed to add env:", err)
		return
	}

	if err := bitrise.RunEnvmanInit(); err != nil {
		log.Error("Failed to run envman init")
		return
	}

	// Run work flow
	if workFlow, err := bitrise.ReadWorkFlowJson(workFlowJsonPath); err != nil {
		log.Errorln("Failed to read work flow:", err)
		return
	} else {
		if err := activateAndRunSteps(workFlow); err != nil {
			log.Errorln("Failed to activate steps:", err)
			return
		}
	}
}
