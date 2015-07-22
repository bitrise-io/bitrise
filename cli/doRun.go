package cli

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
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

func isBuildFailed() bool {
	if len(failedSteps) > 0 {
		return true
	}
	return false
}

func exportEnvironmentsList(envsList []stepmanModels.EnvironmentItemModel) error {
	log.Debugln("[BITRISE_CLI] - Exporting environments:", envsList)

	for _, env := range envsList {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return err
		}

		if value != "" {
			if err := bitrise.RunEnvmanAdd(key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

func cleanupStepWorkDir() error {
	stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"
	if err := bitrise.RemoveFile(stepYMLPth); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step yml: ", err))
	}

	stepDir := bitrise.BitriseWorkStepsDirPath
	if err := bitrise.RemoveDir(stepDir); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step work dir: ", err))
	}
	return nil
}

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string) error {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	for idx, stepListItm := range workflow.Steps {
		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			return err
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			return err
		}
		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		stepDir := bitrise.BitriseWorkStepsDirPath

		if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
			log.Error("Failed to setup stepman:", err)
		}

		if err := cleanupStepWorkDir(); err != nil {
			return err
		}

		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"
		if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.ID, stepIDData.Version, stepDir, stepYMLPth); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run stepman activate")
			failedSteps = append(failedSteps, compositeStepIDStr)
		} else {
			log.Debugf("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.ID, stepIDData.Version)

			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				return err
			}

			if err := models.MergeStepWith(specStep, workflowStep); err != nil {
				return err
			}

			fmt.Println()
			log.Infof("========== (%d) %s ==========", idx, *specStep.Title)
			fmt.Println()

			if isBuildFailed() && !*specStep.IsAlwaysRun {
				log.Infof("A previous step failed and this step was not marked to IsAlwaysRun - skipping %s (%s)", stepIDData.ID, stepIDData.Version)
			} else {
				if err := runStep(specStep, stepIDData); err != nil {
					log.Errorln("[BITRISE_CLI] - Failed to run step:", err)
					failedSteps = append(failedSteps, compositeStepIDStr)
				}
			}
		}
	}
	return nil
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData) error {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.ID, stepIDData.Version)

	// Add step envs
	for _, input := range step.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := input.GetOptions()
		if err != nil {
			return err
		}

		if value != "" {
			log.Debugf("Input: %#v\n", input)
			if err := bitrise.RunEnvmanAdd(key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}

	stepDir := bitrise.BitriseWorkStepsDirPath
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
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
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
	if err := activateAndRunSteps(workflowToRun, bitriseConfig.DefaultStepLibSource); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to activate steps:", err)
	}

	log.Debugln("Failed steps:", failedSteps)
	log.Infoln("")
	log.Infoln("DONE - Congrats!!")
}
