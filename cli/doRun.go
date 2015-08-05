package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/bitrise-io/bitrise-cli/colorstring"
	"github.com/bitrise-io/bitrise-cli/dependencies"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-pathutil/pathutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

	depManagerBrew                 = "brew"
	stepRunSummaryBoxMaxWidthChars = 50

	stepRunResultCodeSuccess            = 0
	stepRunResultCodeFailed             = 1
	stepRunResultCodeFailedNotImportant = 2
	stepRunResultCodeSkipped            = 3
	stepRunResultCodeSkippedWithRunIf   = 4
)

var (
	inventoryPath   string
	startTime       time.Time
	buildRunResults models.BuildRunResultsModel
)

func printRunningWorkflow(title string) {
	fmt.Println()
	log.Info(colorstring.Bluef("Running workflow (%s)", title))
	fmt.Println()
}

func printRunningStep(title string, idx int) {
	content := fmt.Sprintf("(%d) %s", idx, title)
	if len(content) > stepRunSummaryBoxMaxWidthChars {
		dif := stepRunSummaryBoxMaxWidthChars - len(content)
		title = title[0 : len(content)-dif]
		content = fmt.Sprintf("(%d) %s", idx, title)
	}
	sep := strings.Repeat("-", len(content)+4)
	log.Info(sep)
	log.Infof("| (%d) %s |", idx, title)
	log.Info(sep)
}

func printStepSummary(title string, resultCode int, duration time.Duration, exitCode int) {
	runTime := bitrise.TimeToFormattedSeconds(duration, " sec")
	content := fmt.Sprintf("%s | .... | %s", title, runTime)
	if resultCode == stepRunResultCodeFailed || resultCode == stepRunResultCodeFailedNotImportant {
		content = fmt.Sprintf("%s | .... | exit code: %d | %s", title, exitCode, runTime)
	}
	if len(content) > stepRunSummaryBoxMaxWidthChars {
		dif := len(content) - stepRunSummaryBoxMaxWidthChars
		title = title[0:(len(title) - dif)]
		content = fmt.Sprintf("%s | .... | %s", title, runTime)
		if resultCode == stepRunResultCodeFailed || resultCode == stepRunResultCodeFailedNotImportant {
			content = fmt.Sprintf("%s | .... | exit code: %d | %s", title, exitCode, runTime)
		}
	}

	sep := strings.Repeat("-", len(content)+2)
	log.Info(sep)
	switch resultCode {
	case stepRunResultCodeSuccess:
		runStateIcon := "✅ "
		content = fmt.Sprintf("%s | %s | %s", runStateIcon, colorstring.Green(title), runTime)
		break
	case stepRunResultCodeFailed:
		runStateIcon := "❌ "
		content = fmt.Sprintf("%s | %s | %s | exit code: %d", runStateIcon, colorstring.Red(title), runTime, exitCode)
		break
	case stepRunResultCodeFailedNotImportant:
		runStateIcon := "❌ "
		content = fmt.Sprintf("%s | %s | %s | exit code: %d", runStateIcon, colorstring.Yellow(title), runTime, exitCode)
		break
	case stepRunResultCodeSkipped, stepRunResultCodeSkippedWithRunIf:
		runStateIcon := "➡ "
		content = fmt.Sprintf("%s | %s | %s", runStateIcon, colorstring.White(title), runTime)
		break
	default:
		log.Error("Unkown result code")
		return
	}

	log.Infof("| " + content + " |")
	log.Info(sep)
	fmt.Println()
}

func buildFailedFatal(err error) {
	runTime := time.Now().Sub(startTime)
	log.Fatal("Build failed error: " + err.Error() + " total run time: " + runTime.String())
}

func printSummary() {
	totalStepCount := 0
	successStepCount := 0
	failedStepCount := 0
	failedNotImportantStepCount := 0
	skippedStepCount := 0

	successStepCount += len(buildRunResults.SuccessSteps)
	failedStepCount += len(buildRunResults.FailedSteps)
	failedNotImportantStepCount += len(buildRunResults.FailedNotImportantSteps)
	skippedStepCount += len(buildRunResults.SkippedSteps)
	totalStepCount = successStepCount + failedStepCount + failedNotImportantStepCount + skippedStepCount

	fmt.Println()
	log.Infoln("==> Summary:")
	runTime := time.Now().Sub(startTime)
	log.Info("Total run time: " + bitrise.TimeToFormattedSeconds(runTime, " seconds"))

	if totalStepCount > 0 {
		log.Infof("Out of %d steps:", totalStepCount)

		if successStepCount > 0 {
			log.Info(colorstring.Greenf(" * %d was successful", successStepCount))
		}
		if failedStepCount > 0 {
			log.Info(colorstring.Redf(" * %d failed", failedStepCount))
		}
		if failedNotImportantStepCount > 0 {
			log.Info(colorstring.Yellowf(" * %d failed but was marked as skippable and", failedNotImportantStepCount))
		}
		if skippedStepCount > 0 {
			log.Info(colorstring.Whitef(" * %d was skipped", skippedStepCount))
		}

		fmt.Println()
		if failedStepCount > 0 {
			log.Fatal("FINISHED but a couple of steps failed - Ouch")
		} else {
			log.Info("DONE - Congrats!!")
			if failedNotImportantStepCount > 0 {
				log.Warn("P.S.: a couple of non imporatant steps failed")
			}
		}
	}
}

func printStepStatus(stepRunResults models.BuildRunResultsModel) {
	failedCount := len(stepRunResults.FailedSteps)
	failedNotImportantCount := len(stepRunResults.FailedNotImportantSteps)
	skippedCount := len(stepRunResults.SkippedSteps)
	successCount := len(stepRunResults.SuccessSteps)
	totalCount := successCount + failedCount + failedNotImportantCount + skippedCount

	log.Infof("Out of %d steps, %d was successful, %d failed, %d failed but was marked as skippable and %d was skipped",
		totalCount,
		successCount,
		failedCount,
		failedNotImportantCount,
		skippedCount)

	printStepStatusList("Failed steps:", stepRunResults.FailedSteps)
	printStepStatusList("Failed but skippable steps:", stepRunResults.FailedNotImportantSteps)
	printStepStatusList("Skipped steps:", stepRunResults.SkippedSteps)
}

func printStepStatusList(header string, stepList []models.StepRunResultsModel) {
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

func exportEnvironmentsList(envsList []envmanModels.EnvironmentItemModel) error {
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

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string) (workflowRunResults models.BuildRunResultsModel) {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	var stepStartTime time.Time

	registerStepRunResults := func(step stepmanModels.StepModel, resultCode int, exitCode int, err error) {
		stepResults := models.StepRunResultsModel{
			StepName: *step.Title,
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case stepRunResultCodeSuccess:
			workflowRunResults.SuccessSteps = append(workflowRunResults.SuccessSteps, stepResults)
			break
		case stepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", *step.Title, err)
			workflowRunResults.FailedSteps = append(workflowRunResults.FailedSteps, stepResults)
			break
		case stepRunResultCodeFailedNotImportant:
			log.Warnf("Step (%s) failed, but was marked as not important, error: (%v)", *step.Title, err)
			workflowRunResults.FailedNotImportantSteps = append(workflowRunResults.FailedNotImportantSteps, stepResults)
			break
		case stepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", *step.Title)
			workflowRunResults.SkippedSteps = append(workflowRunResults.SkippedSteps, stepResults)
			break
		case stepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + *step.Title + ") Run-If expression evaluated to false - skipping")
			log.Info("The Run-If expression was: ", colorstring.White(*step.RunIf))
			workflowRunResults.SkippedSteps = append(workflowRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		if resultCode != stepRunResultCodeSuccess {
			if err := setBuildFailedEnv(true); err != nil {
				log.Error("Failed to set Build Status envs")
			}
		}

		printStepSummary(*step.Title, resultCode, time.Now().Sub(stepStartTime), exitCode)
	}
	registerStepListItemRunResults := func(stepListItem models.StepListItemModel, resultCode int, exitCode int, err error) {
		name := ""
		for key := range stepListItem {
			name = key
			break
		}

		stepResults := models.StepRunResultsModel{
			StepName: name,
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case stepRunResultCodeSuccess:
			workflowRunResults.SuccessSteps = append(workflowRunResults.SuccessSteps, stepResults)
			break
		case stepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", name, err)
			workflowRunResults.FailedSteps = append(workflowRunResults.FailedSteps, stepResults)
			break
		case stepRunResultCodeFailedNotImportant:
			log.Warnf("Step (%s) failed, but was marked as not important, error: (%v)", name, err)
			workflowRunResults.FailedNotImportantSteps = append(workflowRunResults.FailedNotImportantSteps, stepResults)
			break
		case stepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", name)
			workflowRunResults.SkippedSteps = append(workflowRunResults.SkippedSteps, stepResults)
			break
		case stepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + name + ") Run-If expression evaluated to false - skipping")
			workflowRunResults.SkippedSteps = append(workflowRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		if resultCode != stepRunResultCodeSuccess {
			if err := setBuildFailedEnv(true); err != nil {
				log.Error("Failed to set Build Status envs")
			}
		}

		printStepSummary(name, resultCode, time.Now().Sub(stepStartTime), exitCode)
	}

	for idx, stepListItm := range workflow.Steps {
		stepStartTime = time.Now()

		if err := setBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}
		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
			continue
		}

		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		if err := cleanupStepWorkDir(); err != nil {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)
			if err := bitrise.RunCopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}
			if err := bitrise.RunCopyFile(stepAbsLocalPth+"/step.yml", stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.RunGitClone(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}
			if err := bitrise.RunCopyFile(stepDir+"/step.yml", stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			}

			if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			}
		} else {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData))
			continue
		}

		specStep, err := bitrise.ReadSpecStep(stepYMLPth)
		log.Debugf("Spec read from YML: %#v\n", specStep)
		if err != nil {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
			continue
		}

		mergedStep, err := models.MergeStepWith(specStep, workflowStep)
		if err != nil {
			registerStepListItemRunResults(stepListItm, stepRunResultCodeFailed, 1, err)
			continue
		}

		printRunningStep(*mergedStep.Title, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, workflowRunResults, IsCIMode)
			if err != nil {
				registerStepRunResults(mergedStep, stepRunResultCodeFailed, 1, err)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, stepRunResultCodeSkippedWithRunIf, 0, err)
				continue
			}
		}
		if workflowRunResults.IsBuildFailed() && !*mergedStep.IsAlwaysRun {
			registerStepRunResults(mergedStep, stepRunResultCodeSkipped, 0, err)
			continue
		} else {
			exit, err := runStep(mergedStep, stepIDData, stepDir)
			if err != nil {
				registerStepRunResults(mergedStep, stepRunResultCodeFailed, exit, err)
				continue
			} else {
				registerStepRunResults(mergedStep, stepRunResultCodeSuccess, 0, nil)
			}
		}
	}
	return workflowRunResults
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string) (int, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check dependencies
	for _, dep := range step.Dependencies {
		switch dep.Manager {
		case depManagerBrew:
			err := dependencies.InstallWithBrewIfNeeded(dep.Name)
			if err != nil {
				return 1, err
			}
			break
		default:
			return 1, errors.New("Not supported dependency (" + dep.Manager + ") (" + dep.Name + ")")
		}
	}

	// Add step envs
	for _, input := range step.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return 1, err
		}

		opts, err := input.GetOptions()
		if err != nil {
			return 1, err
		}

		if value != "" {
			log.Debugf("Input: %#v\n", input)
			if err := bitrise.RunEnvmanAdd(key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return 1, err
			}
		}
	}

	stepCmd := stepDir + "/" + "step.sh"
	cmd := []string{"bash", stepCmd}
	log.Debug("OUTPUT:")
	if exit, err := bitrise.RunEnvmanRunInDir(bitrise.CurrentDir, cmd, "panic"); err != nil {
		return exit, err
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)
	return 0, nil
}

func activateAndRunWorkflow(workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel) models.BuildRunResultsModel {
	// Workflow level environments
	if err := exportEnvironmentsList(workflow.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export Workflow environments: " + err.Error()))
	}

	// Run these workflows before running the target workflow
	for _, beforeWorkflowName := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowName]
		if !exist {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + beforeWorkflowName + ") does not exist!"))
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowName
		}
		activateAndRunWorkflow(beforeWorkflow, bitriseConfig)
	}

	// Run the target workflow
	printRunningWorkflow(workflow.Title)
	if err := exportEnvironmentsList(workflow.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export Workflow environments: " + err.Error()))
	}

	stepRunResults := activateAndRunSteps(workflow, bitriseConfig.DefaultStepLibSource)
	buildRunResults.Append(stepRunResults)

	// Run these workflows after running the target workflow
	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + afterWorkflowName + ") does not exist!"))
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowName
		}
		activateAndRunWorkflow(afterWorkflow, bitriseConfig)
	}

	return stepRunResults
}

func doRun(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()

	log.Debugln("[BITRISE_CLI] - Run")

	startTime = time.Now()
	buildRunResults = models.BuildRunResultsModel{}

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

	// Check workflow
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

	// App level environment
	if err := exportEnvironmentsList(bitriseConfig.App.Environments); err != nil {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Failed to export App environments: " + err.Error()))
	}

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunName]
	if !exist {
		buildFailedFatal(errors.New("[BITRISE_CLI] - Specified Workflow (" + workflowToRunName + ") does not exist!"))
	}
	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunName
	}

	activateAndRunWorkflow(workflowToRun, bitriseConfig)

	// // Build finished
	printSummary()
	if len(buildRunResults.FailedSteps) > 0 {
		log.Fatal("[BITRISE_CLI] - Workflow FINISHED but a couple of steps failed - Ouch")
	} else {
		if len(buildRunResults.FailedNotImportantSteps) > 0 {
			log.Warn("[BITRISE_CLI] - Workflow FINISHED but a couple of non imporatant steps failed")
		}
	}
}
