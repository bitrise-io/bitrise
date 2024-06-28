package cli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/cli/docker"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/stepruncmd"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/toolversions"
	envman "github.com/bitrise-io/envman/cli"
	"github.com/bitrise-io/envman/env"
	envmanEnv "github.com/bitrise-io/envman/env"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-steputils/v2/secretkeys"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	coreanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	commandV2 "github.com/bitrise-io/go-utils/v2/command"
	envV2 "github.com/bitrise-io/go-utils/v2/env"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/toolkits"
)

func (r WorkflowRunner) runWorkflow(
	plan models.WorkflowExecutionPlan,
	steplibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool, tracker analytics.Tracker, buildIDProperties coreanalytics.Properties,
) models.BuildRunResultsModel {
	bitrise.PrintRunningWorkflow(plan.WorkflowTitle)

	workflowIDProperties := coreanalytics.Properties{analytics.WorkflowExecutionID: plan.UUID}
	tracker.SendWorkflowStarted(buildIDProperties.Merge(workflowIDProperties), plan.WorkflowID, plan.WorkflowTitle)

	results := r.activateAndRunSteps(plan, steplibSource, buildRunResults, environments, secrets, isLastWorkflow, tracker, workflowIDProperties)

	tracker.SendWorkflowFinished(workflowIDProperties, results.IsBuildFailed())
	collectToolVersions(tracker)

	return results
}

func (r WorkflowRunner) activateAndRunSteps(
	plan models.WorkflowExecutionPlan,
	defaultStepLibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool,
	tracker analytics.Tracker,
	workflowIDProperties coreanalytics.Properties,
) models.BuildRunResultsModel {
	log.Debug("[BITRISE_CLI] - Activating and running steps")

	if len(plan.Steps) == 0 {
		log.Warnf("%s workflow has no steps to run, moving on to the next workflow...", plan.WorkflowTitle)
		return buildRunResults
	}

	runResultCollector := newBuildRunResultCollector(tracker)
	currentStepGroupID := ""

	// Global variables for restricting Step Bundle's environment variables for the given Step Bundle
	currentStepBundleUUID := ""
	// TODO: add the last step bundle's envs to environments
	var currentStepBundleEnvVars []envmanModels.EnvironmentItemModel

	// ------------------------------------------
	// Main - Preparing & running the steps
	for idx, stepPlan := range plan.Steps {
		if stepPlan.WithGroupUUID != currentStepGroupID {
			if stepPlan.WithGroupUUID != "" {
				if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
					r.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, *environments, stepPlan.WithGroupUUID, plan.WorkflowTitle)
				}
			}

			currentStepGroupID = stepPlan.WithGroupUUID
		}

		buildEnvironments := append([]envmanModels.EnvironmentItemModel{}, *environments...)

		if stepPlan.StepBundleUUID != currentStepBundleUUID {
			if stepPlan.StepBundleUUID != "" {
				currentStepBundleEnvVars = append(buildEnvironments, stepPlan.StepBundleEnvs...)
			}

			currentStepBundleUUID = stepPlan.StepBundleUUID
		}

		var envsForStepRun []envmanModels.EnvironmentItemModel
		if currentStepBundleUUID != "" {
			envsForStepRun = currentStepBundleEnvVars
		} else {
			envsForStepRun = buildEnvironments
		}

		stepStartTime := time.Now()
		stepIDProperties := coreanalytics.Properties{analytics.StepExecutionID: stepPlan.UUID}
		stepStartedProperties := workflowIDProperties.Merge(stepIDProperties)

		result := r.activateAndRunStep(
			stepPlan.Step,
			stepPlan.StepID,
			idx,
			defaultStepLibSource,
			stepPlan.UUID,
			tracker,
			envsForStepRun,
			secrets,
			buildRunResults,
			plan.IsSteplibOfflineMode,
			stepPlan.ContainerID,
			stepPlan.WithGroupUUID,
			stepStartTime,
			stepStartedProperties,
		)

		*environments = append(*environments, result.OutputEnvironments...)
		if currentStepBundleUUID != "" {
			currentStepBundleEnvVars = append(currentStepBundleEnvVars, result.OutputEnvironments...)
		}

		isLastStepInWorkflow := idx == len(plan.Steps)-1

		// Shut down containers if the step is in a 'With' group, and it's the last step in the group
		if currentStepGroupID != "" {
			doesStepGroupChange := idx < len(plan.Steps)-1 && currentStepGroupID != plan.Steps[idx+1].WithGroupUUID
			if isLastStepInWorkflow || doesStepGroupChange {
				r.stopContainersForStepGroup(currentStepGroupID, plan.WorkflowTitle)
			}
		}

		isLastStep := isLastWorkflow && isLastStepInWorkflow

		runResultCollector.registerStepRunResults(&buildRunResults, stepPlan.UUID, stepStartTime, stepmanModels.StepModel{}, result.StepInfoPtr, idx,
			result.StepRunStatus, result.StepRunExitCode, result.StepRunErr, isLastStep, result.PrintStepHeader, result.RedactedStepInputs, stepStartedProperties)

		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}
	}

	return buildRunResults
}

type activateAndRunStepResult struct {
	Step               stepmanModels.StepModel
	StepInfoPtr        stepmanModels.StepInfoModel
	PrintStepHeader    bool
	RedactedStepInputs map[string]string
	OutputEnvironments []envmanModels.EnvironmentItemModel
	StepRunStatus      models.StepRunStatus
	StepRunExitCode    int
	StepRunErr         error
}

func newActivateAndRunStepResult(step stepmanModels.StepModel, stepInfoPtr stepmanModels.StepInfoModel, stepRunStatus models.StepRunStatus, stepRunExitCode int, stepRunErr error, printStepHeader bool, redactedStepInputs map[string]string, outputEnvironments []envmanModels.EnvironmentItemModel) activateAndRunStepResult {
	return activateAndRunStepResult{Step: step, StepInfoPtr: stepInfoPtr, StepRunStatus: stepRunStatus, StepRunExitCode: stepRunExitCode, StepRunErr: stepRunErr, PrintStepHeader: printStepHeader, RedactedStepInputs: redactedStepInputs, OutputEnvironments: outputEnvironments}
}

func (r WorkflowRunner) activateAndRunStep(
	step stepmanModels.StepModel,
	stepID string,
	stepIDx int,
	defaultStepLibSource string,
	stepExecutionID string,
	tracker analytics.Tracker,
	environments []envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	buildRunResults models.BuildRunResultsModel,
	isStepLibOfflineMode bool,
	containerID, groupID string,
	stepStartTime time.Time,
	stepStartedProperties coreanalytics.Properties,
) activateAndRunStepResult {
	//
	// Activate step
	activateResult := r.activateStep(step, stepID, defaultStepLibSource, buildRunResults, isStepLibOfflineMode)
	if activateResult.Err != nil {
		return newActivateAndRunStepResult(activateResult.Step, activateResult.StepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, activateResult.Err, true, map[string]string{}, nil)
	}

	stepInfoPtr := activateResult.StepInfoPtr
	mergedStep := activateResult.Step
	stepIDData := activateResult.StepIDData
	stepDir := activateResult.StepDir

	//
	// Run step
	logStepStarted(stepInfoPtr, mergedStep, stepIDx, stepExecutionID, stepStartTime)

	// Evaluate run conditions
	if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
		buildFailedEnvs := bitrise.BuildFailedEnvs(buildRunResults.IsBuildFailed())
		runIfEnvs := append(environments, buildFailedEnvs...)
		runIfEnvList, err := envman.ConvertToEnvsJSONModel(runIfEnvs, true, false, &envmanEnv.DefaultEnvironmentSource{})
		if err != nil {
			err = fmt.Errorf("EnvmanReadEnvList failed, err: %s", err)
			return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, false, map[string]string{}, nil)
		}

		isRun, err := bitrise.EvaluateTemplateToBool(*mergedStep.RunIf, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, runIfEnvList)
		if err != nil {
			return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, false, map[string]string{}, nil)
		}
		if !isRun {
			return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeSkippedWithRunIf, 0, nil, false, map[string]string{}, nil)
		}
	}

	isAlwaysRun := stepmanModels.DefaultIsAlwaysRun
	if mergedStep.IsAlwaysRun != nil {
		isAlwaysRun = *mergedStep.IsAlwaysRun
	} else {
		log.Warnf("Step (%s) mergedStep.IsAlwaysRun is nil, should not!", stepIDData.IDorURI)
	}

	if buildRunResults.IsBuildFailed() && !isAlwaysRun {
		return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeSkipped, 0, nil, false, map[string]string{}, nil)
	}

	// Prepare envs for the step run
	prepareEnvsResult := r.prepareEnvsForStepRun(stepExecutionID, stepDir, mergedStep.Inputs, secrets, buildRunResults, environments)
	if prepareEnvsResult.Err != nil {
		return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, prepareEnvsResult.Err, false, map[string]string{}, nil)
	}

	redactedInputsWithType := prepareEnvsResult.RedactedInputsWithType
	redactedOriginalInputs := prepareEnvsResult.RedactedOriginalInputs
	stepDeclaredEnvironments := prepareEnvsResult.StepDeclaredEnvironments
	stepSecretValues := prepareEnvsResult.StepSecretValues
	stepTestDir := prepareEnvsResult.StepTestDir
	redactedStepInputs := prepareEnvsResult.RedactedStepInputs

	// Run the step
	tracker.SendStepStartedEvent(stepStartedProperties, prepareAnalyticsStepInfo(mergedStep, stepInfoPtr), redactedInputsWithType, redactedOriginalInputs)

	exit, outEnvironments, stepRunErr := r.runStep(stepExecutionID, mergedStep, stepIDData, stepDir, stepDeclaredEnvironments, stepSecretValues, containerID, groupID)

	if stepTestDir != "" {
		if err := addTestMetadata(stepTestDir, models.TestResultStepInfo{Number: stepIDx, Title: *mergedStep.Title, ID: stepIDData.IDorURI, Version: stepIDData.Version}); err != nil {
			log.Errorf("Failed to normalize test result dir, error: %s", err)
		}
	}

	if err := tools.EnvmanClear(configs.OutputEnvstorePath); err != nil {
		log.Errorf("Failed to clear output envstore, error: %s", err)
	}

	if stepRunErr != nil {
		if *mergedStep.IsSkippable {
			return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeFailedSkippable, exit, stepRunErr, false, redactedStepInputs, outEnvironments)
		} else {
			return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeFailed, exit, stepRunErr, false, redactedStepInputs, outEnvironments)
		}
	}

	return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeSuccess, 0, nil, false, redactedStepInputs, outEnvironments)
}

type activateStepResult struct {
	Step        stepmanModels.StepModel
	StepInfoPtr stepmanModels.StepInfoModel
	StepIDData  stepid.CanonicalID
	StepDir     string
	Err         error
}

func newActivateStepResult(step stepmanModels.StepModel, stepInfoPtr stepmanModels.StepInfoModel, stepIDData stepid.CanonicalID, stepDir string, err error) activateStepResult {
	return activateStepResult{Step: step, StepInfoPtr: stepInfoPtr, StepIDData: stepIDData, StepDir: stepDir, Err: err}
}

func (r WorkflowRunner) activateStep(
	step stepmanModels.StepModel,
	stepID string,
	defaultStepLibSource string,
	buildRunResults models.BuildRunResultsModel,
	isStepLibOfflineMode bool,
) activateStepResult {
	// TODO: stepInfoPtr.Step is not a real step, only stores presentation properties (printed in the step boxes)
	stepInfoPtr := stepmanModels.StepInfoModel{}

	compositeStepIDStr := stepID
	workflowStep := step

	stepInfoPtr.ID = compositeStepIDStr
	if workflowStep.Title != nil && *workflowStep.Title != "" {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(*workflowStep.Title)
	} else {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(compositeStepIDStr)
	}

	stepIDData, err := stepid.CreateCanonicalIDFromString(compositeStepIDStr, defaultStepLibSource)
	if err != nil {
		return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepid.CanonicalID{}, "", err)
	}
	stepInfoPtr.ID = stepIDData.IDorURI
	if stepInfoPtr.Step.Title == nil || *stepInfoPtr.Step.Title == "" {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(stepIDData.IDorURI)
	}
	stepInfoPtr.Version = stepIDData.Version
	stepInfoPtr.Library = stepIDData.SteplibSource

	//
	// Activating the step
	if err := bitrise.CleanupStepWorkDir(); err != nil {
		return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepIDData, "", err)
	}

	stepDir := configs.BitriseWorkStepsDirPath

	isStepLibUpdated := false
	if stepIDData.SteplibSource != "" {
		isStepLibUpdated = buildRunResults.IsStepLibUpdated(stepIDData.SteplibSource)
	}

	activator := newStepActivator()
	stepYMLPth, origStepYMLPth, didStepLibUpdate, err := activator.activateStep(stepIDData, isStepLibUpdated, stepDir, configs.BitriseWorkDirPath, &workflowStep, &stepInfoPtr, isStepLibOfflineMode)
	if didStepLibUpdate {
		buildRunResults.StepmanUpdates[stepIDData.SteplibSource]++
	}
	if err != nil {
		return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepIDData, stepDir, err)
	}

	// Fill step info with default step info, if exist
	mergedStep := workflowStep
	if stepYMLPth != "" {
		specStep, err := bitrise.ReadSpecStep(stepYMLPth)
		log.Debugf("Spec read from YML: %#v", specStep)
		if err != nil {
			ymlPth := stepYMLPth
			if origStepYMLPth != "" {
				// in case of local step (path:./) we use the original step definition path,
				// instead of the activated step's one.
				ymlPth = origStepYMLPth
			}
			err = fmt.Errorf("failed to parse step definition (%s): %s", ymlPth, err)
			return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepIDData, stepDir, err)
		}

		mergedStep, err = models.MergeStepWith(specStep, workflowStep)
		if err != nil {
			return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepIDData, stepDir, err)
		}
	}

	if mergedStep.SupportURL != nil {
		stepInfoPtr.Step.SupportURL = pointers.NewStringPtr(*mergedStep.SupportURL)
	}
	if mergedStep.SourceCodeURL != nil {
		stepInfoPtr.Step.SourceCodeURL = pointers.NewStringPtr(*mergedStep.SourceCodeURL)
	}

	if mergedStep.RunIf != nil {
		stepInfoPtr.Step.RunIf = pointers.NewStringPtr(*mergedStep.RunIf)
	}

	if mergedStep.Timeout != nil {
		stepInfoPtr.Step.Timeout = pointers.NewIntPtr(*mergedStep.Timeout)
	}

	if mergedStep.NoOutputTimeout != nil {
		stepInfoPtr.Step.NoOutputTimeout = pointers.NewIntPtr(*mergedStep.NoOutputTimeout)
	}

	// At this point we have a filled up step info model and also have a step model which is contains the merged step
	// data from the bitrise.yml and the steps step.yml.
	// If the step title contains the step id or the step library as a prefix then we will take the original steps
	// title instead.
	// Here are a couple of before and after examples:
	// git::https://github.com/bitrise-steplib/bitrise-step-simple-git-clone.git -> Simple Git Clone
	// certificate-and-profile-installer@1 -> Certificate and profile installer
	if stepInfoPtr.Step.Title != nil && (strings.HasPrefix(*stepInfoPtr.Step.Title, stepInfoPtr.ID) || strings.HasPrefix(*stepInfoPtr.Step.Title, stepInfoPtr.Library)) {
		if mergedStep.Title != nil && *mergedStep.Title != "" {
			*stepInfoPtr.Step.Title = *mergedStep.Title
		}
	}

	return newActivateStepResult(mergedStep, stepInfoPtr, stepIDData, stepDir, nil)
}

type prepareEnvsForStepRunResult struct {
	RedactedStepInputs       map[string]string
	RedactedInputsWithType   map[string]interface{}
	RedactedOriginalInputs   map[string]string
	StepDeclaredEnvironments []envmanModels.EnvironmentItemModel
	StepSecretValues         []string
	StepTestDir              string
	Err                      error
}

func newPrepareEnvsForStepRunResult(redactedStepInputs map[string]string, redactedInputsWithType map[string]interface{}, redactedOriginalInputs map[string]string, stepDeclaredEnvironments []envmanModels.EnvironmentItemModel, stepSecretValues []string, stepTestDir string, err error) prepareEnvsForStepRunResult {
	return prepareEnvsForStepRunResult{RedactedStepInputs: redactedStepInputs, RedactedInputsWithType: redactedInputsWithType, RedactedOriginalInputs: redactedOriginalInputs, StepDeclaredEnvironments: stepDeclaredEnvironments, StepSecretValues: stepSecretValues, StepTestDir: stepTestDir, Err: err}
}

func (r WorkflowRunner) prepareEnvsForStepRun(
	stepExecutionID string,
	stepDir string,
	stepInputs []envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	buildRunResults models.BuildRunResultsModel,
	environments []envmanModels.EnvironmentItemModel,
) prepareEnvsForStepRunResult {
	if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
		return newPrepareEnvsForStepRunResult(nil, nil, nil, nil, nil, "", err)
	}

	if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, environments); err != nil {
		return newPrepareEnvsForStepRunResult(nil, nil, nil, nil, nil, "", err)
	}

	// beside of the envs coming from the current parent process these will be added as an extra
	var additionalEnvironments []envmanModels.EnvironmentItemModel

	// add this environment variable so all child processes can connect their events to their step lifecycle events
	additionalEnvironments = append(additionalEnvironments, envmanModels.EnvironmentItemModel{
		analytics.StepExecutionIDEnvKey: stepExecutionID,
	})

	// add an extra env for the next step run to be able to access the step's source location
	additionalEnvironments = append(additionalEnvironments, envmanModels.EnvironmentItemModel{
		"BITRISE_STEP_SOURCE_DIR": stepDir,
	})

	testDeployDir := os.Getenv(configs.BitriseTestDeployDirEnvKey)
	// If testDeployDir is empty, MkdirTemp() will use the default temp dir. But if it points to a path,
	// we have to create it first.
	if testDeployDir != "" {
		err := os.MkdirAll(testDeployDir, 0755)
		if err != nil {
			log.Warnf("Failed to create %s, error: %s", configs.BitriseTestDeployDirEnvKey, err)
			testDeployDir = ""
		}
	}
	stepTestDir, err := os.MkdirTemp(testDeployDir, "step_test_result")

	if err != nil {
		log.Errorf("Failed to create per-step test result dir: %s", err)
	}

	if stepTestDir != "" {
		// managed to create the test dir, set the env for it for the next step run
		additionalEnvironments = append(additionalEnvironments, envmanModels.EnvironmentItemModel{
			configs.BitrisePerStepTestResultDirEnvKey: stepTestDir,
		})
	}

	environmentItemModels := append(environments, additionalEnvironments...)
	envSource := &env.DefaultEnvironmentSource{}
	stepDeclaredEnvironments, expandedStepEnvironment, redactedInputsWithType, err := prepareStepEnvironment(prepareStepInputParams{
		environment:       environmentItemModels,
		inputs:            stepInputs,
		buildRunResults:   buildRunResults,
		isCIMode:          configs.IsCIMode,
		isPullRequestMode: configs.IsPullRequestMode,
	}, envSource)
	if err != nil {
		err = fmt.Errorf("failed to prepare step environment variables: %s", err)
		return newPrepareEnvsForStepRunResult(nil, nil, nil, nil, nil, "", err)
	}

	stepSecretKeys, stepSecretValues := tools.GetSecretKeysAndValues(secrets)
	if configs.IsSecretEnvsFiltering {
		sensitiveEnvs, err := getSensitiveEnvs(stepDeclaredEnvironments, expandedStepEnvironment)
		if err != nil {
			err = fmt.Errorf("failed to get sensitive inputs: %s", err)
			return newPrepareEnvsForStepRunResult(nil, nil, nil, nil, nil, "", err)
		}

		sensitiveEnvKeys, sensitiveEnvValues := tools.GetSecretKeysAndValues(sensitiveEnvs)
		stepSecretKeys = append(stepSecretKeys, sensitiveEnvKeys...)
		stepSecretValues = append(stepSecretValues, sensitiveEnvValues...)
	}

	redactedStepInputs, redactedOriginalInputs, err := redactStepInputs(expandedStepEnvironment, stepInputs, stepSecretValues)
	if err != nil {
		err = fmt.Errorf("failed to redact step inputs: %s", err)
		return newPrepareEnvsForStepRunResult(nil, nil, nil, nil, nil, "", err)
	}

	for key, value := range redactedStepInputs {
		if _, ok := redactedInputsWithType[key]; !ok {
			redactedInputsWithType[key] = value
		}
	}

	secretKeysEnv := secretEnvKeysEnvironment(stepSecretKeys)
	stepDeclaredEnvironments = append(stepDeclaredEnvironments, secretKeysEnv)

	return newPrepareEnvsForStepRunResult(redactedStepInputs, redactedInputsWithType, redactedOriginalInputs, stepDeclaredEnvironments, stepSecretValues, stepTestDir, nil)
}

func (r WorkflowRunner) runStep(
	stepUUID string,
	step stepmanModels.StepModel,
	stepIDData stepid.CanonicalID,
	stepDir string,
	environments []envmanModels.EnvironmentItemModel,
	secrets []string,
	containerID string,
	groupID string,
) (int, []envmanModels.EnvironmentItemModel, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check & Install Step Dependencies
	// [!] Make sure this happens BEFORE the Toolkit Bootstrap,
	// so that if a Toolkit requires/allows the use of additional dependencies
	// required for the step (e.g. a brew installed OpenSSH) it can be done
	// with a Toolkit+Deps
	if err := retry.Times(2).Try(func(attempt uint) error {
		if attempt > 0 {
			log.Print()
			log.Warn("Installing Step dependency failed, retrying ...")
		}

		return checkAndInstallStepDependencies(step)
	}); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{},
			fmt.Errorf("Failed to install Step dependency, error: %s", err)
	}

	if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, environments); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	// Run step
	bitriseSourceDir, err := getCurrentBitriseSourceDir(environments)
	if err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}
	if bitriseSourceDir == "" {
		bitriseSourceDir = configs.CurrentDir
	}

	if exit, err := r.executeStep(stepUUID, step, stepIDData, stepDir, bitriseSourceDir, secrets, containerID, groupID); err != nil {
		stepOutputs, envErr := bitrise.CollectEnvironmentsFromFile(configs.OutputEnvstorePath)
		if envErr != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, envErr
		}

		updatedStepOutputs, updateErr := stepOutputs, error(nil)

		if configs.IsSecretEnvsFiltering {
			updatedStepOutputs, updateErr = bitrise.ApplySensitiveOutputs(updatedStepOutputs, step.Outputs)
			if updateErr != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, updateErr
			}
		}

		updatedStepOutputs, updateErr = bitrise.ApplyOutputAliases(updatedStepOutputs, step.Outputs)
		if updateErr != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, updateErr
		}

		return exit, updatedStepOutputs, err
	}

	stepOutputs, err := bitrise.CollectEnvironmentsFromFile(configs.OutputEnvstorePath)
	if err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	updatedStepOutputs, updateErr := stepOutputs, error(nil)

	if configs.IsSecretEnvsFiltering {
		updatedStepOutputs, updateErr = bitrise.ApplySensitiveOutputs(updatedStepOutputs, step.Outputs)
		if updateErr != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, updateErr
		}
	}

	updatedStepOutputs, updateErr = bitrise.ApplyOutputAliases(updatedStepOutputs, step.Outputs)
	if updateErr != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, updateErr
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	return 0, updatedStepOutputs, nil
}

func (r WorkflowRunner) executeStep(
	stepUUID string,
	step stepmanModels.StepModel, sIDData stepid.CanonicalID,
	stepAbsDirPath, bitriseSourceDir string,
	secrets []string,
	containerID string,
	groupID string,
) (int, error) {

	toolkitForStep := toolkits.ToolkitForStep(step)
	toolkitName := toolkitForStep.ToolkitName()

	if err := toolkitForStep.PrepareForStepRun(step, sIDData, stepAbsDirPath); err != nil {
		return 1, fmt.Errorf("Failed to prepare the step for execution through the required toolkit (%s), error: %s",
			toolkitName, err)
	}

	cmdArgs, err := toolkitForStep.StepRunCommandArguments(step, sIDData, stepAbsDirPath)
	if err != nil {
		return 1, fmt.Errorf("Toolkit (%s) rejected the step, error: %s",
			toolkitName, err)
	}

	timeout := time.Duration(-1)
	if step.Timeout != nil && *step.Timeout > 0 {
		timeoutSeconds := *step.Timeout
		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	noOutputTimeout := r.config.Modes.NoOutputTimeout
	if step.NoOutputTimeout != nil {
		noOutputTimeout = time.Duration(*step.NoOutputTimeout) * time.Second
	}

	var stepSecrets []string
	if r.config.Modes.SecretFilteringMode {
		stepSecrets = secrets
	}

	opts := log.GetGlobalLoggerOpts()
	opts.Producer = log.Step
	opts.ProducerID = stepUUID
	opts.DebugLogEnabled = true
	logger := log.NewLogger(opts)
	stdout := logwriter.NewLogWriter(logger)

	var name string
	var args []string
	var envs []string

	containerDef := r.ContainerDefinition(containerID)
	if containerDef != nil {
		envs, err = envman.ReadAndEvaluateEnvs(configs.InputEnvstorePath, &docker.EnvironmentSource{
			Logger: logger,
		})
		if err != nil {
			return 1, fmt.Errorf("failed to read command environment: %w", err)
		}

		name = "docker"
		runningContainer := r.dockerManager.GetContainerForStepGroup(groupID)
		if runningContainer == nil {
			return 1, fmt.Errorf("Docker container does not exist")
		}

		args = runningContainer.ExecuteCommandArgs(envs)
		args = append(args, cmdArgs...)

		cmd := stepruncmd.New(name, args, bitriseSourceDir, envs, stepSecrets, timeout, noOutputTimeout, stdout, logV2.NewLogger())

		logger.Infof("Step is running in container: %s", containerDef.Image)
		return cmd.Run()
	}

	envs, err = envman.ReadAndEvaluateEnvs(configs.InputEnvstorePath, &envmanEnv.DefaultEnvironmentSource{})
	if err != nil {
		return 1, fmt.Errorf("failed to read command environment: %w", err)
	}

	name = cmdArgs[0]
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	cmd := stepruncmd.New(name, args, bitriseSourceDir, envs, stepSecrets, timeout, noOutputTimeout, stdout, logV2.NewLogger())

	return cmd.Run()
}

func (r WorkflowRunner) startContainersForStepGroup(containerID string, serviceIDs []string, environments []envmanModels.EnvironmentItemModel, groupID, workflowTitle string) {
	if containerID == "" && len(serviceIDs) == 0 {
		return
	}

	if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
		log.Debugf("Couldn't initialize envman.")
	}
	if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, environments); err != nil {
		log.Debugf("Couldn't add envs.")
	}

	envList, err := tools.EnvmanReadEnvList(configs.InputEnvstorePath)
	if err != nil {
		log.Debugf("Couldn't read envs from envman.")
	}

	if containerID != "" {
		containerDef := r.ContainerDefinition(containerID)
		if containerDef != nil {
			log.Infof("ℹ️ Running workflow in docker container: %s", containerDef.Image)

			_, err := r.dockerManager.StartContainerForStepGroup(*containerDef, groupID, envList)
			if err != nil {
				log.Errorf("Could not start the specified docker image for workflow: %s", workflowTitle)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := r.ServiceDefinitions(serviceIDs...)
		_, err := r.dockerManager.StartServiceContainersForStepGroup(servicesDefs, groupID, envList)
		if err != nil {
			log.Errorf("❌ Some services failed to start properly!")
		}
	}
}

func (r WorkflowRunner) stopContainersForStepGroup(groupID, workflowTitle string) {
	if container := r.dockerManager.GetContainerForStepGroup(groupID); container != nil {
		// TODO: Feature idea, make this configurable, so that we can keep the container for debugging purposes.
		if err := container.Destroy(); err != nil {
			log.Errorf("Attempted to stop the docker container for workflow: %s: %s", workflowTitle, err)
		}
	}

	if services := r.dockerManager.GetServiceContainersForStepGroup(groupID); services != nil {
		for _, container := range services {
			if err := container.Destroy(); err != nil {
				log.Errorf("Attempted to stop the docker container for service: %s: %s", container.Name, err)
			}
		}
	}
}

func isPRMode(prGlobalFlagPtr *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
	if prGlobalFlagPtr != nil {
		return *prGlobalFlagPtr, nil
	}

	prIDEnv := os.Getenv(configs.PullRequestIDEnvKey)
	prModeEnv := os.Getenv(configs.PRModeEnvKey)

	if prIDEnv != "" || prModeEnv == "true" {
		return true, nil
	}

	for _, env := range inventoryEnvironments {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return false, err
		}

		if key == configs.PullRequestIDEnvKey && value != "" {
			return true, nil
		}
		if key == configs.PRModeEnvKey && value == "true" {
			return true, nil
		}
	}

	return false, nil
}

func registerPrMode(isPRMode bool) error {
	configs.IsPullRequestMode = isPRMode
	return os.Setenv(configs.PRModeEnvKey, strconv.FormatBool(isPRMode))
}

func isCIMode(ciGlobalFlagPtr *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
	if ciGlobalFlagPtr != nil {
		return *ciGlobalFlagPtr, nil
	}

	ciModeEnv := os.Getenv(configs.CIModeEnvKey)

	if ciModeEnv == "true" {
		return true, nil
	}

	for _, env := range inventoryEnvironments {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return false, err
		}

		if key == configs.CIModeEnvKey && value == "true" {
			return true, nil
		}
	}

	return false, nil
}

func registerCIMode(isCIMode bool) error {
	configs.IsCIMode = isCIMode
	return os.Setenv(configs.CIModeEnvKey, strconv.FormatBool(isCIMode))
}

func isSecretFiltering(filteringFlag *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
	if filteringFlag != nil {
		return *filteringFlag, nil
	}

	expandedEnvs, err := tools.ExpandEnvItems(inventoryEnvironments, os.Environ())
	if err != nil {
		return false, err
	}

	value, ok := expandedEnvs[configs.IsSecretFilteringKey]
	if ok {
		if value == "true" {
			return true, nil
		} else if value == "false" {
			return false, nil
		}
	}

	return true, nil
}

func registerSecretFiltering(filtering bool) error {
	configs.IsSecretFiltering = filtering
	return os.Setenv(configs.IsSecretFilteringKey, strconv.FormatBool(filtering))
}

func isSecretEnvsFiltering(filteringFlag *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
	if filteringFlag != nil {
		return *filteringFlag, nil
	}

	expandedEnvs, err := tools.ExpandEnvItems(inventoryEnvironments, os.Environ())
	if err != nil {
		return false, err
	}

	value, ok := expandedEnvs[configs.IsSecretEnvsFilteringKey]
	if ok {
		if value == "true" {
			return true, nil
		} else if value == "false" {
			return false, nil
		}
	}

	return true, nil
}

func registerSecretEnvsFiltering(filtering bool) error {
	configs.IsSecretEnvsFiltering = filtering
	return os.Setenv(configs.IsSecretEnvsFilteringKey, strconv.FormatBool(filtering))
}

func isSteplibOfflineMode() bool {
	isSteplibOfflineMode := os.Getenv(configs.IsSteplibOfflineModeEnvKey)
	return isSteplibOfflineMode == "true"
}

func registerSteplibOfflineMode(offlineMode bool) {
	configs.IsSteplibOfflineMode = offlineMode
	// Disable analytics if running in Offline mode
	os.Setenv(analytics.DisabledEnvKey, strconv.FormatBool(offlineMode))
}

func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func GetBitriseConfigFromBase64Data(configBase64Str string) (models.BitriseDataModel, []string, error) {
	configBase64Bytes, err := base64.StdEncoding.DecodeString(configBase64Str)
	if err != nil {
		return models.BitriseDataModel{}, []string{}, fmt.Errorf("Failed to decode base 64 string, error: %s", err)
	}

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes(configBase64Bytes)
	if err != nil {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("Failed to parse bitrise config, error: %s", err)
	}

	return config, warnings, nil
}

func GetBitriseConfigFilePath(bitriseConfigPath string) (string, error) {
	if bitriseConfigPath == "" {
		bitriseConfigPath = filepath.Join(configs.CurrentDir, DefaultBitriseConfigFileName)

		if exist, err := pathutil.IsPathExists(bitriseConfigPath); err != nil {
			return "", err
		} else if !exist {
			return "", fmt.Errorf("bitrise.yml path not defined and not found on it's default path: %s", bitriseConfigPath)
		}
	}

	return bitriseConfigPath, nil
}

func CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath string) (models.BitriseDataModel, []string, error) {
	bitriseConfig := models.BitriseDataModel{}
	warnings := []string{}

	if bitriseConfigBase64Data != "" {
		config, warns, err := GetBitriseConfigFromBase64Data(bitriseConfigBase64Data)
		warnings = warns
		if err != nil {
			return models.BitriseDataModel{}, warnings, fmt.Errorf("Failed to get config (bitrise.yml) from base 64 data, err: %s", err)
		}
		bitriseConfig = config
	} else {
		bitriseConfigPath, err := GetBitriseConfigFilePath(bitriseConfigPath)
		if err != nil {
			return models.BitriseDataModel{}, []string{}, fmt.Errorf("Failed to get config (bitrise.yml) path: %s", err)
		}
		if bitriseConfigPath == "" {
			return models.BitriseDataModel{}, []string{}, errors.New("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
		}

		config, warns, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
		warnings = warns
		if err != nil {
			return models.BitriseDataModel{}, warnings, fmt.Errorf("Config (path:%s) is not valid: %s", bitriseConfigPath, err)
		}
		bitriseConfig = config
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.FormatVersion, bitriseConfig.FormatVersion)
	if err != nil {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("Failed to compare bitrise CLI supported format version (%s) with the bitrise.yml format version (%s): %s", models.FormatVersion, bitriseConfig.FormatVersion, err)
	}
	if !isConfigVersionOK {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("The bitrise.yml has a higher format version (%s) than the bitrise CLI supported format version (%s), please upgrade your bitrise CLI to use this bitrise.yml", bitriseConfig.FormatVersion, models.FormatVersion)
	}

	return bitriseConfig, warnings, nil
}

func GetInventoryFromBase64Data(inventoryBase64Str string) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryBase64Bytes, err := base64.StdEncoding.DecodeString(inventoryBase64Str)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to decode base 64 string, error: %s", err)
	}

	inventory, err := bitrise.InventoryModelFromYAMLBytes(inventoryBase64Bytes)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	return inventory.Envs, nil
}

func GetInventoryFilePath(inventoryPath string) (string, error) {
	if inventoryPath == "" {
		log.Debug("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = filepath.Join(configs.CurrentDir, DefaultSecretsFileName)

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			return "", err
		} else if !exist {
			inventoryPath = ""
		}
	}

	return inventoryPath, nil
}

func CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath string) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryEnvironments := []envmanModels.EnvironmentItemModel{}

	if inventoryBase64Data != "" {
		inventory, err := GetInventoryFromBase64Data(inventoryBase64Data)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to get inventory from base 64 data, err: %s", err)
		}
		inventoryEnvironments = inventory
	} else {
		inventoryPath, err := GetInventoryFilePath(inventoryPath)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to get inventory path: %s", err)
		}

		if inventoryPath != "" {
			bytes, err := fileutil.ReadBytesFromFile(inventoryPath)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, err
			}

			if len(bytes) == 0 {
				return []envmanModels.EnvironmentItemModel{}, errors.New("empty config")
			}

			inventory, err := bitrise.CollectEnvironmentsFromFile(inventoryPath)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Invalid inventory format: %s", err)
			}
			inventoryEnvironments = inventory
		}
	}

	return inventoryEnvironments, nil
}

func getCurrentBitriseSourceDir(envlist []envmanModels.EnvironmentItemModel) (string, error) {
	bitriseSourceDir := os.Getenv(configs.BitriseSourceDirEnvKey)
	for i := len(envlist) - 1; i >= 0; i-- {
		env := envlist[i]

		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return bitriseSourceDir, err
		}

		if key == configs.BitriseSourceDirEnvKey && value != "" {
			return value, nil
		}
	}
	return bitriseSourceDir, nil
}

func checkAndInstallStepDependencies(step stepmanModels.StepModel) error {
	if len(step.Dependencies) > 0 {
		log.Warnf("step.dependencies is deprecated... Use step.deps instead.")
	}

	if step.Deps != nil && (len(step.Deps.Brew) > 0 || len(step.Deps.AptGet) > 0) {
		//
		// New dependency handling
		switch runtime.GOOS {
		case "darwin":
			for _, brewDep := range step.Deps.Brew {
				if err := bitrise.InstallWithBrewIfNeeded(brewDep, configs.IsCIMode); err != nil {
					log.Infof("Failed to install (%s) with brew", brewDep.Name)
					return err
				}
				log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", brewDep.GetBinaryName())
			}
		case "linux":
			for _, aptGetDep := range step.Deps.AptGet {
				log.Infof("Start installing (%s) with apt-get", aptGetDep.Name)
				if err := bitrise.InstallWithAptGetIfNeeded(aptGetDep, configs.IsCIMode); err != nil {
					log.Infof("Failed to install (%s) with apt-get", aptGetDep.Name)
					return err
				}
				log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", aptGetDep.GetBinaryName())
			}
		default:
			return errors.New("unsupported os")
		}
	} else if len(step.Dependencies) > 0 {
		log.Info("Deprecated dependencies found")
		//
		// Deprecated dependency handling
		for _, dep := range step.Dependencies {
			isSkippedBecauseOfPlatform := false
			switch dep.Manager {
			case depManagerBrew:
				if runtime.GOOS == "darwin" {
					err := bitrise.InstallWithBrewIfNeeded(stepmanModels.BrewDepModel{Name: dep.Name}, configs.IsCIMode)
					if err != nil {
						return err
					}
				} else {
					isSkippedBecauseOfPlatform = true
				}
				break
			default:
				return errors.New("Not supported dependency (" + dep.Manager + ") (" + dep.Name + ")")
			}

			if isSkippedBecauseOfPlatform {
				log.Debugf(" * Dependency (%s) skipped, manager (%s) not supported on this platform (%s)", dep.Name, dep.Manager, runtime.GOOS)
			} else {
				log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", dep.Name)
			}
		}
	}

	return nil
}

func logStepStarted(stepInfo stepmanModels.StepInfoModel, step stepmanModels.StepModel, idx int, stepExcutionID string, stepStartTime time.Time) {
	title := ""
	if stepInfo.Step.Title != nil && *stepInfo.Step.Title != "" {
		title = *stepInfo.Step.Title
	}

	params := log.StepStartedParams{
		ExecutionID: stepExcutionID,
		Position:    idx,
		Title:       title,
		ID:          stepInfo.ID,
		Version:     stepInfo.Version,
		Collection:  stepInfo.Library,
		Toolkit:     toolkits.ToolkitForStep(step).ToolkitName(),
		StartTime:   stepStartTime.Format(time.RFC3339),
	}
	log.PrintStepStartedEvent(params)
}

func prepareAnalyticsStepInfo(step stepmanModels.StepModel, stepInfoPtr stepmanModels.StepInfoModel) analytics.StepInfo {
	return analytics.StepInfo{
		StepID:      stepInfoPtr.ID,
		StepTitle:   pointers.StringWithDefault(step.Title, ""),
		StepVersion: stepInfoPtr.Version,
		StepSource:  pointers.StringWithDefault(step.SourceCodeURL, ""),
		Skippable:   pointers.BoolWithDefault(step.IsSkippable, false),
	}
}

func collectToolVersions(tracker analytics.Tracker) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("user home dir not found: %w", err)
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	reporter := toolversions.NewASDFVersionReporter(envV2.NewCommandLocator(), commandV2.NewFactory(envV2.NewRepository()), logger, userHomeDir)

	if !reporter.IsAvailable() {
		log.Debugf("ASDF is not available, skipping tool version reporting")
		return
	}

	toolVersions, err := reporter.CurrentToolVersions()
	if err != nil {
		log.Warnf("Tool version reporting: %s", err)
		return
	}
	toolVersionsBytes, err := json.Marshal(toolVersions)
	if err != nil {
		logger.Warnf("Tool version reporting: JSON marshal: %s", err)
		return
	}

	tracker.SendToolVersionSnapshot(string(toolVersionsBytes), analytics.ToolSnapshotEndOfWorkflowValue)
}

func addTestMetadata(testDirPath string, testResultStepInfo models.TestResultStepInfo) error {
	// check if the test dir is empty
	if empty, err := isDirEmpty(testDirPath); err != nil {
		return fmt.Errorf("failed to check if dir empty: %s, error: %s", testDirPath, err)
	} else if empty {
		// if the test dir is empty then we need to remove the dir from the temp location to not to spam the system with empty dirs
		if err := os.Remove(testDirPath); err != nil {
			return fmt.Errorf("failed to remove dir: %s, error: %s", testDirPath, err)
		}
	} else {
		// if the step put files into the test dir(so it is used) then we won't need to remove the test dir, moreover we need to add extra info from the step parameters
		stepInfoFilePath := filepath.Join(testDirPath, "step-info.json")
		stepResultInfoFile, err := os.Create(stepInfoFilePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %s, error: %s", stepInfoFilePath, err)
		}
		if err := json.NewEncoder(stepResultInfoFile).Encode(testResultStepInfo); err != nil {
			return fmt.Errorf("failed to encode to JSON, error: %s", err)
		}
	}
	return nil
}

func secretEnvKeysEnvironment(keys []string) envmanModels.EnvironmentItemModel {
	value := secretkeys.NewManager().Format(keys)
	return envmanModels.EnvironmentItemModel{secretkeys.EnvKey: value}
}