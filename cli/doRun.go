package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

	depManagerBrew     = "brew"
	depManagerTryCheck = "_"
)

func printAboutUtilityWorkflos() {
	log.Infoln("Note about utility workflows:")
	log.Infoln("Utility workflow names start with '_' (example: _my_utility_workflow),")
	log.Infoln(" these can't be triggered directly but can be used by other workflows")
	log.Infoln(" in the before_run and after_run blocks.")
}

func printAvailableWorkflows(config models.BitriseDataModel) {
	workflowNames := []string{}
	utilityWorkflowNames := []string{}

	for wfName := range config.Workflows {
		if strings.HasPrefix(wfName, "_") {
			utilityWorkflowNames = append(utilityWorkflowNames, wfName)
		} else {
			workflowNames = append(workflowNames, wfName)
		}
	}

	if len(workflowNames) > 0 {
		log.Infoln("The following workflows are available:")
		for _, wfName := range workflowNames {
			log.Infoln(" * " + wfName)
		}

		fmt.Println()
		log.Infoln("You can run a selected workflow with:")
		log.Infoln("-> bitrise run the-workflow-name")
		fmt.Println()
	} else {
		log.Infoln("No workflows are available!")
	}

	if len(utilityWorkflowNames) > 0 {
		log.Infoln("The following utility workflows also defined:")
		for _, wfName := range utilityWorkflowNames {
			log.Infoln(" * " + wfName)
		}

		fmt.Println()
		printAboutUtilityWorkflos()
		fmt.Println()
	}

	os.Exit(1)
}

// GetBitriseConfigFilePath ...
func GetBitriseConfigFilePath(c *cli.Context) (string, error) {
	bitriseConfigPath := c.String(PathKey)

	if bitriseConfigPath == "" {
		log.Debugln("[BITRISE_CLI] - Workflow path not defined, searching for " + DefaultBitriseConfigFileName + " in current folder...")
		bitriseConfigPath = path.Join(bitrise.CurrentDir, DefaultBitriseConfigFileName)

		if exist, err := pathutil.IsPathExists(bitriseConfigPath); err != nil {
			return "", err
		} else if !exist {
			return "", errors.New("No workflow yml found")
		}
	}

	return bitriseConfigPath, nil
}

func setPredefinedEnvironments() error {
	formattedOutputFilePath := path.Join(bitrise.BitriseWorkDirPath, "formatted_output.md")
	log.Debugln("=> formattedOutputFilePath: ", formattedOutputFilePath)
	if err := os.Setenv("BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH", formattedOutputFilePath); err != nil {
		return err
	}
	return nil
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string, environments []envmanModels.EnvironmentItemModel) (int, []envmanModels.EnvironmentItemModel, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check dependencies
	for _, dep := range step.Dependencies {
		switch dep.Manager {
		case depManagerBrew:
			err := bitrise.InstallWithBrewIfNeeded(dep.Name, IsCIMode)
			if err != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, err
			}
			break
		case depManagerTryCheck:
			err := bitrise.DependencyTryCheckTool(dep.Name)
			if err != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, err
			}
			break
		default:
			return 1, []envmanModels.EnvironmentItemModel{}, errors.New("Not supported dependency (" + dep.Manager + ") (" + dep.Name + ")")
		}

		log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", dep.Name)
	}

	// Collect step inputs
	environments = bitrise.AppendEnvironmentSlice(environments, step.Inputs)

	// Cleanup envstore
	if err := bitrise.EnvmanInitAtPath(bitrise.InputEnvstorePath); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}
	if err := bitrise.ExportEnvironmentsList(environments); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	// Run step
	stepCmd := path.Join(stepDir, "step.sh")
	log.Debug("OUTPUT:")
	cmd := []string{"bash", stepCmd}
	if exit, err := bitrise.EnvmanRun(bitrise.InputEnvstorePath, bitrise.CurrentDir, cmd, "panic"); err != nil {
		return exit, []envmanModels.EnvironmentItemModel{}, err
	}

	stepOutputs, err := bitrise.CollectEnvironmentsFromFile(bitrise.OutputEnvstorePath)
	if err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}
	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	return 0, stepOutputs, nil
}

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel) models.BuildRunResultsModel {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	var stepStartTime time.Time

	registerStepRunResults := func(step stepmanModels.StepModel, resultCode int, exitCode int, err error) {
		if step.Title == nil {
			log.Error("Step title is nil, should not happend!")
			step.Title = pointers.NewStringPtr("ERROR! Step title is nil!")
		}

		stepResults := models.StepRunResultsModel{
			StepName: *step.Title,
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case bitrise.StepRunResultCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", *step.Title, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", *step.Title, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", *step.Title)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + *step.Title + ") Run-If expression evaluated to false - skipping")
			log.Info("The Run-If expression was: ", colorstring.Blue(*step.RunIf))
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(*step.Title, resultCode, time.Now().Sub(stepStartTime), exitCode)
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
		case bitrise.StepRunResultCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", name, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", name, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", name)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + name + ") Run-If expression evaluated to false - skipping")
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(name, resultCode, time.Now().Sub(stepStartTime), exitCode)
	}

	for idx, stepListItm := range workflow.Steps {
		stepStartTime = time.Now()

		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}

		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := path.Join(bitrise.BitriseWorkDirPath, "current_step.yml")

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)

			if err := cmdex.CopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepAbsLocalPth, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepDir, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource == "_" {
			log.Debugf("[BITRISE_CLI] - Steplib independent step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

			// Steplib independent steps are completly defined in workflow
			stepYMLPth = ""

			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.StepmanSetup(stepIDData.SteplibSource); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			if err := bitrise.StepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			}
		} else {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData))
			continue
		}

		mergedStep := workflowStep
		if stepYMLPth != "" {
			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			mergedStep, err = models.MergeStepWith(specStep, workflowStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		}

		// Run step
		bitrise.PrintRunningStep(*mergedStep.Title, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, buildRunResults)
			if err != nil {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSkippedWithRunIf, 0, err)
				continue
			}
		}
		outEnvironments := []envmanModels.EnvironmentItemModel{}
		if buildRunResults.IsBuildFailed() && !*mergedStep.IsAlwaysRun {
			registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSkipped, 0, err)
		} else {
			exit, out, err := runStep(mergedStep, stepIDData, stepDir, *environments)
			outEnvironments = out
			if err != nil {
				if *mergedStep.IsSkippable {
					registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailedSkippable, exit, err)
				} else {
					registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailed, exit, err)
				}
			} else {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSuccess, 0, nil)
				*environments = bitrise.AppendEnvironmentSlice(*environments, outEnvironments)
			}
		}
	}

	return buildRunResults
}

func runWorkflow(workflow models.WorkflowModel, steplibSource string, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel) models.BuildRunResultsModel {
	bitrise.PrintRunningWorkflow(workflow.Title)

	*environments = bitrise.AppendEnvironmentSlice(*environments, workflow.Environments)
	return activateAndRunSteps(workflow, steplibSource, buildRunResults, environments)
}

func activateAndRunWorkflow(workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel) models.BuildRunResultsModel {
	// Run these workflows before running the target workflow
	for _, beforeWorkflowName := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowName]
		if !exist {
			bitrise.PrintBuildFailedFatal(buildRunResults.StartTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+beforeWorkflowName+") does not exist!"))
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowName
		}
		buildRunResults = activateAndRunWorkflow(beforeWorkflow, bitriseConfig, buildRunResults, environments)
	}

	// Run the target workflow
	buildRunResults = runWorkflow(workflow, bitriseConfig.DefaultStepLibSource, buildRunResults, environments)

	// Run these workflows after running the target workflow
	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			bitrise.PrintBuildFailedFatal(buildRunResults.StartTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+afterWorkflowName+") does not exist!"))
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowName
		}
		buildRunResults = activateAndRunWorkflow(afterWorkflow, bitriseConfig, buildRunResults, environments)
	}

	return buildRunResults
}

func doRun(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)
	log.Debugln("[BITRISE_CLI] - Run")

	if !bitrise.CheckIsSetupWasDoneForVersion(c.App.Version) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(c.App.Version, false); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	}

	startTime := time.Now()

	// Input validation
	bitriseConfigPath, err := GetBitriseConfigFilePath(c)
	if err != nil {
		bitrise.PrintBuildFailedFatal(startTime, fmt.Errorf("[BITRISE_CLI] - Failed to get config (bitrise.yml) path: %s", err))
	}
	if bitriseConfigPath == "" {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to get config (bitrise.yml) path: empty bitriseConfigPath"))
	}

	secretEnvironments := []envmanModels.EnvironmentItemModel{}
	inventoryPath := c.String(InventoryKey)
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = path.Join(bitrise.CurrentDir, DefaultSecretsFileName)

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to check path: "+err.Error()))
		} else if !exist {
			log.Debugln("[BITRISE_CLI] - No inventory yml found")
			inventoryPath = ""
		}
	} else {
		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to check path: "+err.Error()))
		} else if !exist {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - No inventory yml found"))
		}
	}
	if inventoryPath != "" {
		secretEnvironments, err = bitrise.CollectEnvironmentsFromFile(inventoryPath)
		if err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("Invalid invetory format: "+err.Error()))
		}
	}

	// Workflow selection
	workflowToRunName := ""
	if len(c.Args()) < 1 {
		log.Errorln("No workfow specified!")
	} else {
		workflowToRunName = c.Args()[0]
	}

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.OutputEnvstorePath); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to add env: "+err.Error()))
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to add env: "+err.Error()))
	}

	if err := bitrise.EnvmanInit(); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to run envman init"))
	}

	// Run workflow
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to read Workflow: "+err.Error()))
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(c.App.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI version: ", c.App.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI's version (%s).", bitriseConfig.FormatVersion, c.App.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	// Check workflow
	if workflowToRunName == "" {
		// no workflow specified
		//  list all the available ones and then exit
		printAvailableWorkflows(bitriseConfig)
	}

	// App level environment
	environments := bitrise.AppendEnvironmentSlice(bitriseConfig.App.Environments, secretEnvironments)

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunName]
	if !exist {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+workflowToRunName+") does not exist!"))
	}

	if strings.HasPrefix(workflowToRunName, "_") {
		log.Error("Utility workflows can't be triggered directly")
		fmt.Println()
		printAboutUtilityWorkflos()
		os.Exit(1)
	}

	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunName
	}
	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_ID", workflowToRunName); err != nil {
		log.Fatal("Failed to set BITRISE_TRIGGERED_WORKFLOW_ID env:", err)
	}
	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_TITLE", workflowToRun.Title); err != nil {
		log.Fatal("Failed to set BITRISE_TRIGGERED_WORKFLOW_TITLE env:", err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: startTime,
	}

	if err := setPredefinedEnvironments(); err != nil {
		log.Fatalln("Failed to register pre-defined Environment Variables: ", err)
	}
	environments = bitrise.AppendEnvironmentSlice(environments, workflowToRun.Environments)
	buildRunResults = activateAndRunWorkflow(workflowToRun, bitriseConfig, buildRunResults, &environments)

	// Build finished
	bitrise.PrintSummary(buildRunResults)
	if len(buildRunResults.FailedSteps) > 0 {
		log.Fatal("[BITRISE_CLI] - Workflow FINISHED but a couple of steps failed - Ouch")
	} else {
		if len(buildRunResults.FailedSkippableSteps) > 0 {
			log.Warn("[BITRISE_CLI] - Workflow FINISHED but a couple of non imporatant steps failed")
		}
	}
}
