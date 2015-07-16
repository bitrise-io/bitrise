package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/codegangsta/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"
)

var (
	failedSteps   []string
	inventoryPath string
)

// StepIDData ...
type StepIDData struct {
	ID            string
	Version       string
	SteplibSource string
}

func isBuildFailed() bool {
	if len(failedSteps) > 0 {
		return true
	}
	return false
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

func exportEnvironmentsList(envsList []models.InputModel) error {
	log.Debugln("[BITRISE_CLI] - Exporting environments:", envsList)

	for _, env := range envsList {
		envKey := env.MappedTo
		envValue := env.Value
		if envValue != "" {
			if err := bitrise.RunEnvmanAdd(envKey, envValue, *env.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

func activateAndRunSteps(workflow models.WorkflowModel) error {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	for idx, stepListItm := range workflow.Steps {
		// TODO: first arg should be 'stepCompositeID'
		//  which can contain the step-collection, step-id, version, etc.
		//  in one string!
		compositeStepIDStr, workflowStep, err := stepListItm.GetStepIDStepDataPair()
		if err != nil {
			return err
		}
		stepIDData, err := createStepIDDataFromString(compositeStepIDStr)
		if err != nil {
			return err
		}
		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		fmt.Println()
		log.Infof("========== (%d) %s ==========", idx, workflowStep.Name)
		fmt.Println()
		stepDir := bitrise.BitriseWorkStepsDirPath + "/" + stepIDData.ID + "/" + stepIDData.Version + "/"

		if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
			log.Error("Failed to setup stepman:", err)
		}

		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"
		if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.ID, stepIDData.Version, stepDir, stepYMLPth); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run stepman activate")
			failedSteps = append(failedSteps, compositeStepIDStr)
		} else {
			log.Debugf("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.ID, stepIDData.Version)

			if specStep, err := bitrise.ReadSpecStep(stepYMLPth); err != nil {
				log.Fatal("Failed to read spec step:", err)
			} else {
				specStep.MergeWith(workflowStep)

				if err := bitrise.RemoveFile(stepYMLPth); err != nil {
					log.Fatal("Failed to remove step yml:", err)
				}

				if err := runStep(specStep, stepIDData); err != nil {
					log.Errorln("[BITRISE_CLI] - Failed to run step:", err)
					failedSteps = append(failedSteps, compositeStepIDStr)
				}
			}
		}
	}
	return nil
}

func runStep(step models.StepModel, stepIDData StepIDData) error {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.ID, stepIDData.Version)

	// Add step envs
	for _, input := range step.Inputs {
		envKey := input.MappedTo
		envValue := input.Value
		if envValue != "" {
			log.Info("Input:", input)
			if err := bitrise.RunEnvmanAdd(envKey, envValue, *input.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}

	stepDir := bitrise.BitriseWorkStepsDirPath + "/" + stepIDData.ID + "/" + stepIDData.Version + "/"
	stepCmd := stepDir + "/" + "step.sh"
	cmd := []string{"bash", stepCmd}
	if err := bitrise.RunEnvmanRunInDir(bitrise.CurrentDir, cmd); err != nil {
		log.Errorln("[BITRISE_CLI] - Failed to run envman run")
		return err
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.ID, stepIDData.Version)
	return nil
}

func doRun(c *cli.Context) {
	log.Debugln("[BITRISE_CLI] - Run")

	// Cleanup
	if err := bitrise.CleanupBitriseWorkPath(); err != nil {
		log.Fatal("Failed to cleanup bitrise work dir:", err)
	}
	failedSteps = []string{}

	// Input validation
	bitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" {
		log.Debugln("[BITRISE_CLI] - Workflow path not defined, searching for " + DefaultBitriseConfigFileName + " in current folder...")

		if exist, err := pathutil.IsPathExists("./" + DefaultBitriseConfigFileName); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to check path:", err)
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow yml found")
		}
		bitriseConfigPath = "./" + DefaultBitriseConfigFileName
	}

	inventoryPath = c.String(InventoryKey)
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = bitrise.CurrentDir + "/" + DefaultSecretsFileName

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to check path:", err)
		} else if !exist {
			log.Debugln("[BITRISE_CLI] - No inventory yml found")
			inventoryPath = ""
		}
	} else {
		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to check path:", err)
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No inventory yml found")
		}
	}
	if inventoryPath != "" {
		if err := bitrise.RunEnvmanEnvstoreTest(inventoryPath); err != nil {
			log.Fatal("Invalid invetory format:", err)
		}

		if err := bitrise.RunCopy(inventoryPath, bitrise.EnvstorePath); err != nil {
			log.Fatal("Failed to copy inventory:", err)
		}
	}

	// Workflow selection
	workflowToRunName := ""
	if len(c.Args()) < 1 {
		log.Infoln("No workfow specified!")
	} else {
		workflowToRunName = c.Args()[0]
	}

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.EnvstorePath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if inventoryPath == "" {
		if err := bitrise.RunEnvmanInit(); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to run envman init")
		}
	}

	// Run work flow
	bitriseConfig, err := bitrise.ReadBitriseConfigYML(bitriseConfigPath)
	if err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to read Workflow:", err)
	}

	// check workflow
	if workflowToRunName == "" {
		// no workflow specified
		//  list all the available ones and then exit
		log.Infoln("The following workflows are available:")
		for wfName := range bitriseConfig.Workflows {
			log.Infoln(" * " + wfName)
		}
		os.Exit(1)
	}

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunName]
	if !exist {
		log.Fatalln("[BITRISE_CLI] - Specified Workflow (" + workflowToRunName + ") does not exist!")
	}
	log.Infoln("[BITRISE_CLI] - Running Workflow:", workflowToRunName)

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

	log.Debugln("Failed steps:", failedSteps)
	log.Infoln("")
	log.Infoln("DONE - Congrats!!")
}
