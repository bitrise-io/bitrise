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

func buildFailedFatal(err error) {
	runTime := time.Now().Sub(startTime)
	log.Fatal("Build failed error: " + err.Error() + " total run time: " + runTime.String())
}

func printStepStatus(stepRunResults models.StepRunResultsModel) {
	failedCount := len(stepRunResults.FailedSteps)
	failedNotImportantCount := len(stepRunResults.FailedNotImportantSteps)
	skippedCount := len(stepRunResults.SkippedSteps)
	successCount := stepRunResults.TotalStepCount - failedCount - failedNotImportantCount - skippedCount

	log.Infof("Out of %d steps, %d was successful, %d failed, %d failed but was marked as skippable and %d was skipped",
		stepRunResults.TotalStepCount,
		successCount,
		failedCount,
		failedNotImportantCount,
		skippedCount)

	printStepStatusList("Failed steps:", stepRunResults.FailedSteps)
	printStepStatusList("Failed but skippable steps:", stepRunResults.FailedNotImportantSteps)
	printStepStatusList("Skipped steps:", stepRunResults.SkippedSteps)
}

func printStepStatusList(header string, stepList []models.FailedStepModel) {
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

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string) (stepRunResults models.StepRunResultsModel) {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	stepRunResults = models.StepRunResultsModel{
		TotalStepCount:          0,
		FailedSteps:             []models.FailedStepModel{},
		FailedNotImportantSteps: []models.FailedStepModel{},
		SkippedSteps:            []models.FailedStepModel{},
	}

	registerFailedStepListItem := func(stepListItem models.StepListItemModel, err error) {
		name := ""
		for key := range stepListItem {
			name = key
			break
		}

		failedStep := models.FailedStepModel{
			StepName: name,
			Error:    err,
		}
		stepRunResults.FailedSteps = append(stepRunResults.FailedSteps, failedStep)
		log.Errorf("Failed to execute step: (%v) error: (%v)", name, err)
	}
	registerFailedStep := func(step stepmanModels.StepModel, err error) {
		failedStep := models.FailedStepModel{
			StepName: *step.Title,
			Error:    err,
		}

		if *step.IsSkippable {
			stepRunResults.FailedNotImportantSteps = append(stepRunResults.FailedNotImportantSteps, failedStep)
			log.Warnf("Failed to execute step: (%v) error: (%v), but it's marked as skippable", *step.Title, err)
		} else {
			stepRunResults.FailedSteps = append(stepRunResults.FailedSteps, failedStep)
			log.Errorf("Failed to execute step: (%v) error: (%v)", *step.Title, err)
		}
	}

	stepRunResults.TotalStepCount = len(workflow.Steps)

	for idx, stepListItm := range workflow.Steps {
		if err := setBuildFailedEnv(stepRunResults.IsBuildFailed()); err != nil {
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

		if stepIDData.SteplibSource == "path" {
			log.Infof("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}

			if err := bitrise.RunCopyDir(stepAbsLocalPth, stepDir); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
			if err := bitrise.RunCopyFile(stepAbsLocalPth+"/step.yml", stepYMLPth); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Infof("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.RunGitClone(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
			if err := bitrise.RunCopyFile(stepDir+"/step.yml", stepYMLPth); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			log.Infof("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			}

			if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerFailedStepListItem(stepListItm, err)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: %s (%s)", stepIDData.IDorURI, stepIDData.Version)
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

		mergedStep, err := models.MergeStepWith(specStep, workflowStep)
		if err != nil {
			registerFailedStepListItem(stepListItm, err)
			continue
		}

		fmt.Println()
		log.Infof("========== (%d) %s ==========", idx, *mergedStep.Title)
		fmt.Println()

		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, stepRunResults, IsCIMode)
			if err != nil {
				registerFailedStep(mergedStep, err)
				continue
			}
			if !isRun {
				log.Warn("The step's Is-Run expression evaluated to false - skipping")
				log.Info(" The Is-Run expression was: ", *mergedStep.RunIf)
				skippedStep := models.FailedStepModel{
					StepName: *mergedStep.Title,
				}
				stepRunResults.SkippedSteps = append(stepRunResults.SkippedSteps, skippedStep)
				continue
			}
		}
		if stepRunResults.IsBuildFailed() && !*mergedStep.IsAlwaysRun {
			log.Warnf("A previous step failed and this step was not marked to IsAlwaysRun - skipping step (id:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			skippedStep := models.FailedStepModel{
				StepName: *mergedStep.Title,
			}
			stepRunResults.SkippedSteps = append(stepRunResults.SkippedSteps, skippedStep)
			continue
		} else {
			if err := runStep(mergedStep, stepIDData, stepDir); err != nil {
				registerFailedStep(mergedStep, err)
				continue
			}
		}
	}
	return stepRunResults
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string) error {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

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

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)
	return nil
}

func activateAndRunWorkflow(workflowToRun models.WorkflowModel, bitriseConfig models.BitriseDataModel) models.StepRunResultsModel {
	// App level environment
	if err := exportEnvironmentsList(bitriseConfig.App.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export App environments: " + err.Error()))
	}

	// Workflow level environments
	if err := exportEnvironmentsList(workflowToRun.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export Workflow environments: " + err.Error()))
	}

	// Run the Workflow
	return activateAndRunSteps(workflowToRun, bitriseConfig.DefaultStepLibSource)
}

func doRun(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()
	log.Debugln("[BITRISE_CLI] - Run")

	startTime = time.Now()
	workflowMap := map[string]bool{}

	// Cleanup
	if err := bitrise.CleanupBitriseWorkPath(); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to cleanup bitrise work dir: " + err.Error()))
	}

	// Input validation
	bitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" {
		log.Debugln("[BITRISE_CLI] - Workflow path not defined, searching for " + DefaultBitriseConfigFileName + " in current folder...")
		bitriseConfigPath = bitrise.CurrentDir + "/" + DefaultBitriseConfigFileName

		if exist, err := pathutil.IsPathExists(bitriseConfigPath); err != nil {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to check path:" + err.Error()))
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow yml found")
			buildFailedFatal(errors.New("[BITRISE_CLI] - No workflow yml found"))
		}
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
	workflowMap[workflowToRunName] = true

	// Run beforeWorkflow
	if workflowToRun.BeforeWorkflow != "" {
		if workflowMap[workflowToRun.BeforeWorkflow] {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow refer cycle found: (" + workflowToRunName + ") - before: ( " + workflowToRun.BeforeWorkflow + ")"))
		}

		log.Info("[BITRISE_CLI] - Before workflow defined: ", workflowToRun.BeforeWorkflow)
		beforeWorkflowToRun, exist := bitriseConfig.Workflows[workflowToRun.BeforeWorkflow]
		if !exist {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + workflowToRun.BeforeWorkflow + ") does not exist!"))
		}
		log.Infoln("[BITRISE_CLI] - Running Workflow:", workflowToRun.BeforeWorkflow)
		log.Infof("++++++++++++++ Running before workflow: %s ++++++++++++++++", workflowToRun.BeforeWorkflow)
		fmt.Println()
		stepRunResults := activateAndRunWorkflow(beforeWorkflowToRun, bitriseConfig)
		fmt.Println()
		log.Infof("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		printStepStatus(stepRunResults)
		if len(stepRunResults.FailedSteps) > 0 {
			log.Fatal("[BITRISE_CLI] - Before workflow FINISHED but a couple of steps failed - Ouch")
		} else {
			if len(stepRunResults.FailedNotImportantSteps) > 0 {
				log.Warn("[BITRISE_CLI] - Before workflow FINISHED but a couple of non imporatant steps failed")
			}
		}

		workflowMap[workflowToRun.BeforeWorkflow] = true
	}

	log.Infoln("[BITRISE_CLI] - Running Workflow:", workflowToRunName)
	log.Infof("+++++++++++++++++ Running workflow: %s +++++++++++++++++++", workflowToRunName)
	fmt.Println()
	stepRunResults := activateAndRunWorkflow(workflowToRun, bitriseConfig)
	fmt.Println()
	log.Info("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	if workflowToRun.AfterWorkflow != "" {
		if workflowMap[workflowToRun.AfterWorkflow] {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow refer cycle found: (" + workflowToRunName + ") - after: ( " + workflowToRun.AfterWorkflow + ")"))
		}

		log.Info("[BITRISE_CLI] - After workflow defined: ", workflowToRun.AfterWorkflow)
		afterWorkflowToRun, exist := bitriseConfig.Workflows[workflowToRun.AfterWorkflow]
		if !exist {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + workflowToRun.AfterWorkflow + ") does not exist!"))
		}
		log.Infoln("[BITRISE_CLI] - Running Workflow:", workflowToRun.AfterWorkflow)
		log.Infof("+++++++++++++ Running after workflow: %s +++++++++++++++", workflowToRun.AfterWorkflow)
		fmt.Println()
		stepRunResults := activateAndRunWorkflow(afterWorkflowToRun, bitriseConfig)
		fmt.Println()
		log.Info("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		printStepStatus(stepRunResults)
		if len(stepRunResults.FailedSteps) > 0 {
			log.Fatal("[BITRISE_CLI] - After workflow FINISHED but a couple of steps failed - Ouch")
		} else {
			if len(stepRunResults.FailedNotImportantSteps) > 0 {
				log.Warn("[BITRISE_CLI] - After workflow FINISHED but a couple of non imporatant steps failed")
			}
		}
	}

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
