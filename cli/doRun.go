package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

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
	failedSteps            []FailedStepModel
	failedNotInporentSteps []FailedStepModel
	inventoryPath          string
	startTime              time.Time
)

// FailedStepModel ...
type FailedStepModel struct {
	StepName string
	Error    error
}

func isBuildFailed() bool {
	if len(failedSteps) > 0 {
		return true
	}
	return false
}

func registerFailedStepListItem(stepListItem models.StepListItemModel, err error) {
	name := ""
	for key := range stepListItem {
		name = key
		break
	}

	failedStep := FailedStepModel{
		StepName: name,
		Error:    err,
	}
	failedSteps = append(failedSteps, failedStep)
	log.Errorf("Failed to execute step: (%v) error: (%v)", name, err)
}

func registerFailedStep(step stepmanModels.StepModel, err error) {
	failedStep := FailedStepModel{
		StepName: *step.Title,
		Error:    err,
	}

	if *step.IsNotImportant {
		failedNotInporentSteps = append(failedNotInporentSteps, failedStep)
		log.Errorf("Failed to execute step: (%v) error: (%v), but it's marked as not important", *step.Title, err)
		fmt.Println()
	} else {
		failedSteps = append(failedSteps, failedStep)
		log.Errorf("Failed to execute step: (%v) error: (%v)", *step.Title, err)
		fmt.Println()
	}
}

func buildFailedFatal(err error) {
	runTime := time.Now().Sub(startTime)
	printStepStatus()
	log.Fatal("Build failed error: " + err.Error() + " total run time: " + runTime.String())
}

func printStepStatus() {
	printFailedStepsIfExist()
	printFailedNotInportentStepsIsExist()
}

func printFailedNotInportentStepsIsExist() {
	if len(failedNotInporentSteps) > 0 {
		log.Infof("%d not important step(s) failed:", len(failedNotInporentSteps))
		for _, failedNotImportantStep := range failedNotInporentSteps {
			log.Infof(" * Step: (%s) | error: (%v)", failedNotImportantStep.StepName, failedNotImportantStep.Error)
		}
	}
}

func printFailedStepsIfExist() {
	if len(failedSteps) > 0 {
		log.Infof("%d step(s) failed:", len(failedSteps))
		for _, failedStep := range failedSteps {
			log.Infof(" * Step: (%s) | error: (%v)", failedStep.StepName, failedStep.Error)
		}
	}
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
			registerFailedStepListItem(stepListItm, err)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		}
		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		stepDir := bitrise.BitriseWorkStepsDirPath

		if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		}

		if err := cleanupStepWorkDir(); err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		}

		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"
		if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.ID, stepIDData.Version, stepDir, stepYMLPth); err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		} else {
			log.Debugf("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.ID, stepIDData.Version)

			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}

			if err := models.MergeStepWith(specStep, workflowStep); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}

			fmt.Println()
			log.Infof("========== (%d) %s ==========", idx, *specStep.Title)
			fmt.Println()

			if isBuildFailed() && !*specStep.IsAlwaysRun {
				log.Infof("A previous step failed and this step was not marked to IsAlwaysRun - skipping step (id:%s) (version:%s)", stepIDData.ID, stepIDData.Version)
			} else {
				if err := runStep(specStep, stepIDData); err != nil {
					registerFailedStep(specStep, err)
					continue
				}
			}

			fmt.Println()
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

	startTime = time.Now()
	failedSteps = []FailedStepModel{}

	// Cleanup
	if err := bitrise.CleanupBitriseWorkPath(); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to cleanup bitrise work dir: " + err.Error()))
	}

	// Input validation
	bitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" {
		log.Debugln("[BITRISE_CLI] - Workflow path not defined, searching for " + DefaultBitriseConfigFileName + " in current folder...")

		if exist, err := pathutil.IsPathExists("./" + DefaultBitriseConfigFileName); err != nil {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to check path:" + err.Error()))
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow yml found")
			buildFailedFatal(errors.New("[BITRISE_CLI] - No workflow yml found"))
		}
		bitriseConfigPath = "./" + DefaultBitriseConfigFileName
	}

	inventoryPath = c.String(InventoryKey)
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = bitrise.CurrentDir + "/" + DefaultSecretsFileName

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to check path: " + err.Error()))
		} else if !exist {
			log.Debugln("[BITRISE_CLI] - No inventory yml found")
			inventoryPath = ""
		}
	} else {
		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to check path: " + err.Error()))
		} else if !exist {
			buildFailedFatal(errors.New("[BITRISE_CLI] - No inventory yml found"))
		}
	}
	if inventoryPath != "" {
		if err := bitrise.RunEnvmanEnvstoreTest(inventoryPath); err != nil {
			buildFailedFatal(errors.New("Invalid invetory format: " + err.Error()))
		}

		if err := bitrise.RunCopy(inventoryPath, bitrise.EnvstorePath); err != nil {
			buildFailedFatal(errors.New("Failed to copy inventory: " + err.Error()))
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
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to add env: " + err.Error()))
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to add env: " + err.Error()))
	}

	if inventoryPath == "" {
		if err := bitrise.RunEnvmanInit(); err != nil {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to run envman init"))
		}
	}

	// Run work flow
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to read Workflow: " + err.Error()))
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
		buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + workflowToRunName + ") does not exist!"))
	}
	log.Infoln("[BITRISE_CLI] - Running Workflow:", workflowToRunName)

	// App level environment
	if err := exportEnvironmentsList(bitriseConfig.App.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export App environments: " + err.Error()))
	}

	// Workflow level environments
	if err := exportEnvironmentsList(workflowToRun.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export Workflow environments: " + err.Error()))
	}

	// Run the Workflow
	if err := activateAndRunSteps(workflowToRun, bitriseConfig.DefaultStepLibSource); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to activate steps:", err)
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to activate steps: " + err.Error()))
	}

	// Build finished
	fmt.Println()
	log.Infoln("==> Summary:")
	runTime := time.Now().Sub(startTime)
	printStepStatus()
	log.Info("Total run time: " + runTime.String())
	if len(failedSteps) > 0 {
		log.Fatal("FINISHED but a couple of steps failed - Ouch")
	} else {
		log.Info("DONE - Congrats!!")
		if len(failedNotInporentSteps) > 0 {
			log.Warn("P.S.: a couple of non imporatant steps failed")
		}
	}
}
