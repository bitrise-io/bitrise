package cli

import (
	"errors"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/codegangsta/cli"
)

const (
	defaultBitriseConfigFileName = "bitrise.yml"
)

// StepIDData ...
type StepIDData struct {
	ID            string
	Version       string
	SteplibSource string
}

func createStepIDDataFromString(s string) (StepIDData, error) {
	libsourceStepSplits := strings.Split(s, "::")
	if len(libsourceStepSplits) != 2 {
		return StepIDData{}, errors.New("Steplib should be separated with a '::' separator from the step ID (" + s + ")")
	}
	stepidVersionSplits := strings.Split(libsourceStepSplits[1], "@")
	if len(stepidVersionSplits) != 2 {
		return StepIDData{}, errors.New("Step ID and version should be separated with a '@' separator (" + libsourceStepSplits[1] + ")")
	}

	return StepIDData{
		SteplibSource: libsourceStepSplits[0],
		ID:            stepidVersionSplits[0],
		Version:       stepidVersionSplits[1],
	}, nil
}

func exportEnvironmentsList(envsList []models.EnvironmentItemModel) error {
	log.Info("[BITRISE_CLI] - Exporting workflow environments")

	for _, env := range envsList {
		envKey, envValue, err := env.GetKeyValue()
		if err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to get environment key-value pair from env:", env)
			return err
		}
		if envValue != "" {
			expand := bitrise.ParseBool(env["is_expand"], true)
			if err := bitrise.RunEnvmanAdd(envKey, envValue, !expand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

func activateAndRunSteps(workflow models.WorkflowModel) error {
	log.Info("[BITRISE_CLI] - Activating and running steps")

	for _, stepListItm := range workflow.Steps {
		// TODO: first arg should be 'stepCompositeID'
		//  which can contain the step-collection, step-id, version, etc.
		//  in one string!
		compositeStepIDStr, step, err := stepListItm.GetStepIDStepDataPair()
		if err != nil {
			return err
		}
		stepIDData, err := createStepIDDataFromString(compositeStepIDStr)
		if err != nil {
			return err
		}
		log.Infof("Running Step: %#v", step)
		stepDir := "./steps/" + stepIDData.ID + "/" + stepIDData.Version + "/"

		if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
			log.Error("Failed to setup stepman:", err)
		}

		if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.ID, stepIDData.Version, stepDir); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run stepman activate")
			return err
		}

		log.Infof("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.ID, stepIDData.Version)

		if err := runStep(step, stepIDData); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run step")
			return err
		}
	}
	return nil
}

func runStep(step models.StepModel, stepIDData StepIDData) error {
	log.Infof("[BITRISE_CLI] - Running step: %s (%s)", stepIDData.ID, stepIDData.Version)

	// Add step envs
	for _, input := range step.Inputs {
		envKey, envValue, err := input.GetKeyValue()
		if err != nil {
			return err
		}
		if envValue != "" {
			expand := bitrise.ParseBool(input["is_expand"], true)
			if err := bitrise.RunEnvmanAdd(envKey, envValue, expand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}

	stepDir := "./steps/" + stepIDData.ID + "/" + stepIDData.Version + "/"
	//stepCmd := fmt.Sprintf("%sstep.sh", stepDir)
	stepCmd := "step.sh"
	cmd := []string{"bash", stepCmd}

	if err := bitrise.RunEnvmanRunInDir(stepDir, cmd); err != nil {
		log.Errorln("[BITRISE_CLI] - Failed to run envman run")
		return err
	}

	log.Infof("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.ID, stepIDData.Version)
	return nil
}

func doRun(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Run")

	// Input validation
	bitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" {
		log.Infoln("[BITRISE_CLI] - Workflow path not defined, searching for " + defaultBitriseConfigFileName + " in current folder...")

		if exist, err := pathutil.IsPathExists("./" + defaultBitriseConfigFileName); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to check path:", err)
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow yml found")
		}
		bitriseConfigPath = "./" + defaultBitriseConfigFileName
	}

	// Workflow selection
	if len(c.Args()) < 1 {
		log.Fatalln("No workfow specified!")
	}
	workflowToRunName := c.Args()[0]

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.EnvstorePath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := bitrise.RunEnvmanInit(); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to run envman init")
	}

	// Run work flow
	bitriseConfig, err := bitrise.ReadBitriseConfigYML(bitriseConfigPath)
	if err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to read Workflow:", err)
	}
	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunName]
	if !exist {
		log.Fatalln("Specified Workflow (" + workflowToRunName + ") does not exist!")
	}
	log.Infoln("Running Workflow:", workflowToRunName)

	// App level environment
	if err := exportEnvironmentsList(bitriseConfig.App.Environments); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to export App environments:", err)
	}

	// Workflow level environments
	if err := exportEnvironmentsList(workflowToRun.Environments); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to export Workflow environments:", err)
	}

	// Run the Workflow
	if err := activateAndRunSteps(workflowToRun); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to activate steps:", err)
	}
}
