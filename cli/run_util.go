package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
)

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
	environments = append(environments, step.Inputs...)

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

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel, isLastWorkflow bool) models.BuildRunResultsModel {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	var stepStartTime time.Time

	registerStepRunResults := func(step stepmanModels.StepModel, resultCode, exitCode int, err error, isLastStepInWorkflow bool) {
		if step.Title == nil {
			log.Error("Step title is nil, should not happend!")
			step.Title = pointers.NewStringPtr("ERROR! Step title is nil!")
		}

		stepResults := models.StepRunResultsModel{
			StepName: *step.Title,
			Status:   resultCode,
			Idx:      buildRunResults.ResultsCount(),
			RunTime:  time.Now().Sub(stepStartTime),
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case models.StepRunStatusCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case models.StepRunStatusCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", *step.Title, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case models.StepRunStatusCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", *step.Title, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case models.StepRunStatusCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", *step.Title)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case models.StepRunStatusCodeSkippedWithRunIf:
			log.Warn("The step's (" + *step.Title + ") Run-If expression evaluated to false - skipping")
			log.Info("The Run-If expression was: ", colorstring.Blue(*step.RunIf))
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(stepResults, isLastStepInWorkflow)
	}

	registerStepListItemRunResults := func(stepListItem models.StepListItemModel, resultCode, exitCode int, err error, isLastStepInWorkflow bool) {
		name := ""
		for key := range stepListItem {
			name = key
			break
		}

		stepResults := models.StepRunResultsModel{
			StepName: name,
			Status:   resultCode,
			Idx:      buildRunResults.ResultsCount(),
			RunTime:  time.Now().Sub(stepStartTime),
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case models.StepRunStatusCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case models.StepRunStatusCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", name, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case models.StepRunStatusCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", name, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case models.StepRunStatusCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", name)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case models.StepRunStatusCodeSkippedWithRunIf:
			log.Warn("The step's (" + name + ") Run-If expression evaluated to false - skipping")
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(stepResults, isLastStepInWorkflow)
	}

	for idx, stepListItm := range workflow.Steps {
		stepStartTime = time.Now()
		isLastStepInWorkflow := isLastWorkflow && (idx == len(workflow.Steps)-1)

		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}

		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
			continue
		}

		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := path.Join(bitrise.BitriseWorkDirPath, "current_step.yml")

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)

			if err := cmdex.CopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepAbsLocalPth, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepDir, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}
		} else if stepIDData.SteplibSource == "_" {
			log.Debugf("[BITRISE_CLI] - Steplib independent step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

			// Steplib independent steps are completly defined in workflow
			stepYMLPth = ""

			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.StepmanSetup(stepIDData.SteplibSource); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}

			if err := bitrise.StepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			}
		} else {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData), isLastStepInWorkflow)
			continue
		}

		mergedStep := workflowStep
		if stepYMLPth != "" {
			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}

			mergedStep, err = models.MergeStepWith(specStep, workflowStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}
		}

		// Run step
		bitrise.PrintRunningStep(*mergedStep.Title, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, buildRunResults)
			if err != nil {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeFailed, 1, err, isLastStepInWorkflow)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeSkippedWithRunIf, 0, err, isLastStepInWorkflow)
				continue
			}
		}
		outEnvironments := []envmanModels.EnvironmentItemModel{}
		if buildRunResults.IsBuildFailed() && !*mergedStep.IsAlwaysRun {
			registerStepRunResults(mergedStep, models.StepRunStatusCodeSkipped, 0, err, isLastStepInWorkflow)
		} else {
			exit, out, err := runStep(mergedStep, stepIDData, stepDir, *environments)
			outEnvironments = out
			if err != nil {
				if *mergedStep.IsSkippable {
					registerStepRunResults(mergedStep, models.StepRunStatusCodeFailedSkippable, exit, err, isLastStepInWorkflow)
				} else {
					registerStepRunResults(mergedStep, models.StepRunStatusCodeFailed, exit, err, isLastStepInWorkflow)
				}
			} else {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeSuccess, 0, nil, isLastStepInWorkflow)
				*environments = append(*environments, outEnvironments...)
			}
		}
	}

	return buildRunResults
}

func runWorkflow(workflow models.WorkflowModel, steplibSource string, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel, isLastWorkflow bool) models.BuildRunResultsModel {
	bitrise.PrintRunningWorkflow(workflow.Title)

	*environments = append(*environments, workflow.Environments...)
	return activateAndRunSteps(workflow, steplibSource, buildRunResults, environments, isLastWorkflow)
}

func activateAndRunWorkflow(workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel, lastWorkflowTitle string) (models.BuildRunResultsModel, error) {
	var err error
	// Run these workflows before running the target workflow
	for _, beforeWorkflowName := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowName]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist!", beforeWorkflowName)
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowName
		}
		buildRunResults, err = activateAndRunWorkflow(beforeWorkflow, bitriseConfig, buildRunResults, environments, lastWorkflowTitle)
		if err != nil {
			return buildRunResults, err
		}
	}

	// Run the target workflow
	isLastWorkflow := (workflow.Title == lastWorkflowTitle)
	buildRunResults = runWorkflow(workflow, bitriseConfig.DefaultStepLibSource, buildRunResults, environments, isLastWorkflow)

	// Run these workflows after running the target workflow
	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist!", afterWorkflowName)
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowName
		}
		buildRunResults, err = activateAndRunWorkflow(afterWorkflow, bitriseConfig, buildRunResults, environments, lastWorkflowTitle)
		if err != nil {
			return buildRunResults, err
		}
	}

	return buildRunResults, nil
}

func lastWorkflowIDInConfig(workflowToRunID string, bitriseConfig models.BitriseDataModel) (string, error) {
	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunID]
	if !exist {
		return "", errors.New("No worfklow exist with ID: " + workflowToRunID)
	}

	if len(workflowToRun.AfterRun) > 0 {
		lastAfterID := workflowToRun.AfterRun[len(workflowToRun.AfterRun)-1]
		wfID, err := lastWorkflowIDInConfig(lastAfterID, bitriseConfig)
		if err != nil {
			return "", err
		}
		workflowToRunID = wfID
	}
	return workflowToRunID, nil
}

func runWorkflowWithConfiguration(
	startTime time.Time,
	workflowToRunID string,
	bitriseConfig models.BitriseDataModel,
	secretEnvironments []envmanModels.EnvironmentItemModel) (models.BuildRunResultsModel, error) {

	if err := bitrise.InitPaths(); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to initialize required paths: %s", err)
	}

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunID]
	if !exist {
		return models.BuildRunResultsModel{}, fmt.Errorf("Specified Workflow (%s) does not exist!", workflowToRunID)
	}

	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunID
	}

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.OutputEnvstorePath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := bitrise.EnvmanInit(); err != nil {
		return models.BuildRunResultsModel{}, errors.New("Failed to run envman init")
	}

	// App level environment
	environments := append(bitriseConfig.App.Environments, secretEnvironments...)

	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_ID", workflowToRunID); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_ID env: %s", err)
	}
	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_TITLE", workflowToRun.Title); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_TITLE env: %s", err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: startTime,
	}

	environments = append(environments, workflowToRun.Environments...)

	lastWorkflowTitle, err := lastWorkflowIDInConfig(workflowToRunID, bitriseConfig)
	if err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to get last workflow id: %s", err)
	}

	buildRunResults, err = activateAndRunWorkflow(workflowToRun, bitriseConfig, buildRunResults, &environments, lastWorkflowTitle)
	if err != nil {
		return buildRunResults, errors.New("[BITRISE_CLI] - Failed to activate and run workflow " + workflowToRunID)
	}

	// Build finished
	bitrise.PrintSummary(buildRunResults)
	if buildRunResults.IsBuildFailed() {
		return buildRunResults, errors.New("[BITRISE_CLI] - Workflow FINISHED but a couple of steps failed - Ouch")
	}
	if buildRunResults.HasFailedSkippableSteps() {
		log.Warn("[BITRISE_CLI] - Workflow FINISHED but a couple of non imporatant steps failed")
	}
	return buildRunResults, nil
}
