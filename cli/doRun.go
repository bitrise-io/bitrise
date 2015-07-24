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
	inventoryPath string
	startTime     time.Time
)

// FailedStepModel ...
type FailedStepModel struct {
	StepName string
	Error    error
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	TotalStepCount          int
	FailedSteps             []FailedStepModel
	FailedNotImportantSteps []FailedStepModel
	SkippedSteps            []FailedStepModel
}

func buildFailedFatal(err error) {
	runTime := time.Now().Sub(startTime)
	log.Fatal("Build failed error: " + err.Error() + " total run time: " + runTime.String())
}

func printStepStatus(stepRunResults StepRunResultsModel) {
	failedCount := len(stepRunResults.FailedSteps)
	failedNotImportantCount := len(stepRunResults.FailedNotImportantSteps)
	skippedCount := len(stepRunResults.SkippedSteps)
	successCount := stepRunResults.TotalStepCount - failedCount - failedNotImportantCount - skippedCount

	log.Infof("Out of %d steps, %d was successful, %d failed, %d failed but was marked as not important and %d was skipped",
		stepRunResults.TotalStepCount,
		successCount,
		failedCount,
		failedNotImportantCount,
		skippedCount)

	printStepStatusList("Failed steps:", stepRunResults.FailedSteps)
	printStepStatusList("Failed but not important steps:", stepRunResults.FailedNotImportantSteps)
	printStepStatusList("Skipped steps:", stepRunResults.SkippedSteps)
}

func printStepStatusList(header string, stepList []FailedStepModel) {
	if len(stepList) > 0 {
		log.Infof(header)
		for _, step := range stepList {
			if step.Error != nil {
				log.Infof(" * Step: (%s) | error: (%v)", step.StepName, step.Error)
			} else {
				log.Infof(" * Step: (%s)", step.StepName)
			}
		}
	}
}

func setBuildFailedEnv(failed bool) error {
	statusStr := "0"
	if failed {
		statusStr = "1"
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", statusStr); err != nil {
		return err
	}

	if err := os.Setenv("BITRISE_BUILD_STATUS", statusStr); err != nil {
		return err
	}
	return nil
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

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string) (stepRunResults StepRunResultsModel) {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	stepRunResults = StepRunResultsModel{
		TotalStepCount:          0,
		FailedSteps:             []FailedStepModel{},
		FailedNotImportantSteps: []FailedStepModel{},
		SkippedSteps:            []FailedStepModel{},
	}

	registerFailedStepListItem := func(stepListItem models.StepListItemModel, err error) {
		name := ""
		for key := range stepListItem {
			name = key
			break
		}

		failedStep := FailedStepModel{
			StepName: name,
			Error:    err,
		}
		stepRunResults.FailedSteps = append(stepRunResults.FailedSteps, failedStep)
		log.Errorf("Failed to execute step: (%v) error: (%v)", name, err)
	}
	registerFailedStep := func(step stepmanModels.StepModel, err error) {
		failedStep := FailedStepModel{
			StepName: *step.Title,
			Error:    err,
		}

		if *step.IsNotImportant {
			stepRunResults.FailedNotImportantSteps = append(stepRunResults.FailedNotImportantSteps, failedStep)
			log.Errorf("Failed to execute step: (%v) error: (%v), but it's marked as not important", *step.Title, err)
		} else {
			stepRunResults.FailedSteps = append(stepRunResults.FailedSteps, failedStep)
			log.Errorf("Failed to execute step: (%v) error: (%v)", *step.Title, err)
		}
	}
	isBuildFailed := func() bool {
		return len(stepRunResults.FailedSteps) > 0
	}
	stepRunResults.TotalStepCount = len(workflow.Steps)

	for idx, stepListItm := range workflow.Steps {
		if err := setBuildFailedEnv(isBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}
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

		if err := cleanupStepWorkDir(); err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"

		log.Debugf("StepIdData: %v", stepIDData)
		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: %s (%s)", stepIDData.ID, stepIDData.Version)

			if err := bitrise.RunCopyDir(stepIDData.ID, stepDir); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
			if err := bitrise.RunCopyFile(stepIDData.ID+"/step.yml", stepYMLPth); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {

		} else if stepIDData.SteplibSource != "" {
			log.Debug("[BITRISE_CLI] - Activating step")
			if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}

			if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.ID, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.ID, stepIDData.Version)
			}
		} else {
			registerFailedStepListItem(stepListItm, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData))
			continue
		}

		log.Debug("------------Step YML:", stepYMLPth)
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
			skippedStep := FailedStepModel{
				StepName: *specStep.Title,
			}
			stepRunResults.SkippedSteps = append(stepRunResults.SkippedSteps, skippedStep)
		} else {
			if err := runStep(specStep, stepIDData, stepDir); err != nil {
				registerFailedStep(specStep, err)
				continue
			}
		}
	}
	return stepRunResults
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string) error {
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

	stepCmd := stepDir + "/" + "step.sh"
	cmd := []string{"bash", stepCmd}
	fmt.Println("----------------------- OUTPUT ---------------------------")
	err := bitrise.RunEnvmanRunInDir(bitrise.CurrentDir, cmd, "panic")
	fmt.Println("----------------------------------------------------------")
	if err != nil {
		return err
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.ID, stepIDData.Version)
	return nil
}

func doRun(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()
	log.Debugln("[BITRISE_CLI] - Run")

	startTime = time.Now()

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
		fmt.Println()
		log.Infoln("You can run a selected workflow with:")
		log.Infoln("-> bitrise-cli run the-workflow-name")
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
	stepRunResults := activateAndRunSteps(workflowToRun, bitriseConfig.DefaultStepLibSource)

	// Build finished
	fmt.Println()
	log.Infoln("==> Summary:")
	runTime := time.Now().Sub(startTime)
	printStepStatus(stepRunResults)
	log.Info("Total run time: " + runTime.String())
	if len(stepRunResults.FailedSteps) > 0 {
		log.Fatal("FINISHED but a couple of steps failed - Ouch")
	} else {
		log.Info("DONE - Congrats!!")
		if len(stepRunResults.FailedNotImportantSteps) > 0 {
			log.Warn("P.S.: a couple of non imporatant steps failed")
		}
	}
}
