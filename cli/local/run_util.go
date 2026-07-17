package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/docker"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/log/logwriter"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/stepruncmd"
	"github.com/bitrise-io/bitrise/v2/tools"
	envman "github.com/bitrise-io/envman/v2/cli"
	envmanEnv "github.com/bitrise-io/envman/v2/env"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-steputils/v2/secretkeys"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	coreanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/toolkits"
)

func (r WorkflowRunner) runWorkflow(
	plan models.WorkflowExecutionPlan,
	steplibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool, buildIDProperties coreanalytics.Properties,
) models.BuildRunResultsModel {
	workflowIDProperties := coreanalytics.Properties{analytics.WorkflowExecutionID: plan.UUID}
	r.tracker.SendWorkflowStarted(buildIDProperties.Merge(workflowIDProperties), plan.WorkflowID, plan.WorkflowTitle)

	results := r.activateAndRunSteps(plan, steplibSource, buildRunResults, environments, secrets, isLastWorkflow, workflowIDProperties)

	r.tracker.SendWorkflowFinished(workflowIDProperties, results.IsBuildFailed())

	return results
}

func (r WorkflowRunner) activateAndRunSteps(
	plan models.WorkflowExecutionPlan,
	defaultStepLibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool,
	workflowIDProperties coreanalytics.Properties,
) models.BuildRunResultsModel {
	log.Debug("[BITRISE_CLI] - Activating and running steps")

	if len(plan.Steps) == 0 {
		log.Warnf("%s workflow has no steps to run, moving on to the next workflow...", plan.WorkflowTitle)
		return buildRunResults
	}

	runResultCollector := newBuildRunResultCollector(r.logger, r.tracker)

	// Global variables for restricting Step Bundle's environment variables for the given Step Bundle
	currentStepBundleUUID := ""
	var currentStepBundleEnvVars []envmanModels.EnvironmentItemModel

	// Each Step Bundle's run_if is evaluated once, when the Bundle is entered, and the decision is
	// cached here keyed by the Bundle's UUID. This keeps the Bundle's run_if encapsulated: a Step in
	// the Bundle cannot change the run_if outcome for its sibling Steps by setting an output env var.
	stepBundleRunIfResults := map[string]bool{}

	// ------------------------------------------
	// Main - Preparing & running the steps
	for idx, stepPlan := range plan.Steps {
		r.containerManager.UpdateWithStepStarted(stepPlan, *environments)

		workflowEnvironments := append([]envmanModels.EnvironmentItemModel{}, *environments...)

		if stepPlan.StepBundleUUID != currentStepBundleUUID {
			if stepPlan.StepBundleUUID != "" {
				currentStepBundleEnvVars = append(workflowEnvironments, stepPlan.StepBundleEnvs...)
			}

			currentStepBundleUUID = stepPlan.StepBundleUUID
		}

		var envsForStepRun []envmanModels.EnvironmentItemModel
		if currentStepBundleUUID != "" {
			envsForStepRun = currentStepBundleEnvVars
		} else {
			envsForStepRun = workflowEnvironments
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
			envsForStepRun,
			secrets,
			buildRunResults,
			plan.IsSteplibOfflineMode,
			stepStartTime,
			stepStartedProperties,
			stepPlan.StepBundleRunIfs,
			stepBundleRunIfResults,
		)

		*environments = append(*environments, result.OutputEnvironments...)
		if currentStepBundleUUID != "" {
			currentStepBundleEnvVars = append(currentStepBundleEnvVars, result.OutputEnvironments...)
		}

		isLastStepInWorkflow := idx == len(plan.Steps)-1
		isLastStep := isLastWorkflow && isLastStepInWorkflow

		previousBuildRunResult := buildRunResults

		runResultCollector.registerStepRunResults(&buildRunResults, stepPlan.UUID, stepStartTime, stepmanModels.StepModel{}, result.StepInfoPtr, idx,
			result.StepRunStatus, result.StepRunExitCode, result.StepRunErr, isLastStep, result.PrintStepHeader, result.RedactedStepInputs, stepStartedProperties)

		r.containerManager.UpdateWithStepFinished(idx, plan, stepPlan)

		currentBuildRunResult := buildRunResults
		if !previousBuildRunResult.IsBuildFailed() && currentBuildRunResult.IsBuildFailed() {
			if len(currentBuildRunResult.FailedSteps) == 1 {
				failedStepRunResult := currentBuildRunResult.FailedSteps[0]
				failedStepEnvs := bitrise.FailedStepEnvs(failedStepRunResult)
				*environments = append(*environments, failedStepEnvs...)
				if currentStepBundleUUID != "" {
					currentStepBundleEnvVars = append(currentStepBundleEnvVars, failedStepEnvs...)
				}
			}

			buildStatusEnvs := bitrise.BuildStatusEnvs(true)
			*environments = append(*environments, buildStatusEnvs...)
			if currentStepBundleUUID != "" {
				currentStepBundleEnvVars = append(currentStepBundleEnvVars, buildStatusEnvs...)
			}
		}

	}

	return buildRunResults
}

type activateAndRunStepResult struct {
	Step               stepmanModels.StepModel
	StepInfoPtr        stepmanModels.StepInfoModel
	StepRunStatus      models.StepRunStatus
	StepRunExitCode    int
	StepRunErr         error
	PrintStepHeader    bool
	RedactedStepInputs map[string]string
	OutputEnvironments []envmanModels.EnvironmentItemModel
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
	environments []envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	buildRunResults models.BuildRunResultsModel,
	isStepLibOfflineMode bool,
	stepStartTime time.Time,
	stepStartedProperties coreanalytics.Properties,
	stepBundleRunIfs []models.StepBundleRunIf,
	stepBundleRunIfResults map[string]bool,
) activateAndRunStepResult {
	stepInfoPtr, stepIDData, err := newStepInfoPtr(stepID, defaultStepLibSource, step)
	if err != nil {
		return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, true, map[string]string{}, nil)
	}

	if len(stepBundleRunIfs) > 0 {
		// To run the Step each of the including Step Bundles run_if statements must evaluate to true, from the top most to the bottom most.
		// Each Bundle's run_if is evaluated only once, when the Bundle is entered (its first Step), and the cached decision is reused
		// for the Bundle's remaining Steps. This keeps the run_if encapsulated: a Step in the Bundle cannot change the outcome for its
		// sibling Steps by modifying an env var.
		for _, stepBundleRunIf := range stepBundleRunIfs {
			isRun, evaluated := stepBundleRunIfResults[stepBundleRunIf.BundleUUID]
			if !evaluated {
				runIfEnvList, err := envman.ConvertToEnvsJSONModel(environments, true, false, &envmanEnv.DefaultEnvironmentSource{})
				if err != nil {
					err = fmt.Errorf("EnvmanReadEnvList failed, err: %s", err)
					return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, true, map[string]string{}, nil)
				}

				isRun, err = bitrise.EvaluateTemplateToBool(stepBundleRunIf.RunIf, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, runIfEnvList)
				if err != nil {
					return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, true, map[string]string{}, nil)
				}
				stepBundleRunIfResults[stepBundleRunIf.BundleUUID] = isRun
			}

			if !isRun {
				// In the workflow run logs stepInfoPtr.Step.RunIf is used as a reason for skipping the step.
				stepInfoPtr.Step.RunIf = pointers.NewStringPtr(stepBundleRunIf.RunIf)
				return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodeSkippedWithRunIf, 0, nil, true, map[string]string{}, nil)
			}
		}
	}

	// Evaluate run_if before activation when it's explicitly set in bitrise.yml.
	// If not set, the step.yml default may apply and can only be checked post-activation.
	if step.RunIf != nil && *step.RunIf != "" {
		runIfEnvList, err := envman.ConvertToEnvsJSONModel(environments, true, false, &envmanEnv.DefaultEnvironmentSource{})
		if err != nil {
			err = fmt.Errorf("EnvmanReadEnvList failed, err: %s", err)
			return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, true, map[string]string{}, nil)
		}

		isRun, err := bitrise.EvaluateTemplateToBool(*step.RunIf, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, runIfEnvList)
		if err != nil {
			return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, err, true, map[string]string{}, nil)
		}
		if !isRun {
			stepInfoPtr.Step.RunIf = pointers.NewStringPtr(*step.RunIf)
			return newActivateAndRunStepResult(step, stepInfoPtr, models.StepRunStatusCodeSkippedWithRunIf, 0, nil, true, map[string]string{}, nil)
		}
	}

	//
	// Activate step
	activateStartTime := time.Now()
	activateResult := r.activateStep(step, stepInfoPtr, stepIDData, buildRunResults, isStepLibOfflineMode)
	activateDuration := time.Since(activateStartTime)
	if activateResult.Err != nil {
		return newActivateAndRunStepResult(activateResult.Step, activateResult.StepInfoPtr, models.StepRunStatusCodePreparationFailed, 1, activateResult.Err, true, map[string]string{}, nil)
	}

	stepInfoPtr = activateResult.StepInfoPtr
	mergedStep := activateResult.Step
	stepDir := activateResult.StepDir

	//
	// Run step
	logStepStarted(r.logger, stepInfoPtr, mergedStep, stepIDx, stepExecutionID, stepStartTime)

	// Evaluate run_if from step.yml default (only reached when bitrise.yml didn't set run_if).
	if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
		runIfEnvList, err := envman.ConvertToEnvsJSONModel(environments, true, false, &envmanEnv.DefaultEnvironmentSource{})
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
	r.tracker.SendStepStartedEvent(stepStartedProperties, prepareAnalyticsStepInfo(mergedStep, stepInfoPtr), activateDuration, redactedInputsWithType, redactedOriginalInputs)

	exit, outEnvironments, stepRunErr := r.runStep(stepExecutionID, mergedStep, stepIDData, stepDir, activateResult.ExecutablePath, stepDeclaredEnvironments, stepSecretValues)

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
		}
		return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeFailed, exit, stepRunErr, false, redactedStepInputs, outEnvironments)
	}

	return newActivateAndRunStepResult(mergedStep, stepInfoPtr, models.StepRunStatusCodeSuccess, 0, nil, false, redactedStepInputs, outEnvironments)
}

type activateStepResult struct {
	Step           stepmanModels.StepModel
	StepInfoPtr    stepmanModels.StepInfoModel
	StepDir        string
	ExecutablePath string
	Err            error
}

func newActivateStepResult(step stepmanModels.StepModel, stepInfoPtr stepmanModels.StepInfoModel, stepDir, executablePath string, err error) activateStepResult {
	return activateStepResult{Step: step, StepInfoPtr: stepInfoPtr, StepDir: stepDir, ExecutablePath: executablePath, Err: err}
}

func newStepInfoPtr(stepID, defaultStepLibSource string, step stepmanModels.StepModel) (stepmanModels.StepInfoModel, stepid.CanonicalID, error) {
	// TODO: stepInfoPtr.Step is not a real step, only stores presentation properties (printed in the step boxes)
	stepInfoPtr := stepmanModels.StepInfoModel{}

	compositeStepIDStr := stepID

	if step.Title != nil && *step.Title != "" {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(*step.Title)
	} else {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(compositeStepIDStr)
	}

	stepIDData, err := stepid.CreateCanonicalIDFromString(compositeStepIDStr, defaultStepLibSource)
	if err != nil {
		return stepInfoPtr, stepIDData, err
	}

	stepInfoPtr.ID = stepIDData.IDorURI
	if stepInfoPtr.Step.Title == nil || *stepInfoPtr.Step.Title == "" {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(stepIDData.IDorURI)
	}
	stepInfoPtr.Version = stepIDData.Version
	stepInfoPtr.Library = stepIDData.SteplibSource

	return stepInfoPtr, stepIDData, nil
}

func (r WorkflowRunner) activateStep(
	step stepmanModels.StepModel,
	stepInfoPtr stepmanModels.StepInfoModel,
	stepIDData stepid.CanonicalID,
	buildRunResults models.BuildRunResultsModel,
	isStepLibOfflineMode bool,
) activateStepResult {
	//
	// Activating the step
	if err := bitrise.CleanupStepWorkDir(); err != nil {
		return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, "", "", err)
	}

	stepDir := configs.BitriseWorkStepsDirPath

	isStepLibUpdated := false
	if stepIDData.SteplibSource != "" {
		isStepLibUpdated = buildRunResults.IsStepLibUpdated(stepIDData.SteplibSource)
	}

	activationStartedAt := time.Now()
	activator := newStepActivator()
	activatedStep, err := activator.activateStep(stepIDData, isStepLibUpdated, stepDir, configs.BitriseWorkDirPath, isStepLibOfflineMode)
	r.tracker.SendStepActivationEvent(
		activatedStep.ActivationType,
		stepIDData.IDorURI,
		err == nil,
		time.Since(activationStartedAt),
		activatedStep.DidStepLibUpdate,
	)
	if activatedStep.DidStepLibUpdate {
		buildRunResults.StepmanUpdates[stepIDData.SteplibSource]++
	}
	if err != nil {
		return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepDir, "", err)
	}

	// Fill the presentation step info (shown in the step header boxes) from stepman's result.
	// Since stepman v0.21.3 this is populated for every activation type, so no guard is needed.
	// ID and Title are left as newStepInfoPtr seeded them (the bitrise.yml reference, e.g. "./"
	// for a path step) so we keep the relative ref rather than stepman's absolute path.
	stepInfoPtr.Version = activatedStep.StepInfo.Version
	stepInfoPtr.LatestVersion = activatedStep.StepInfo.LatestVersion
	stepInfoPtr.OriginalVersion = activatedStep.StepInfo.OriginalVersion
	stepInfoPtr.GroupInfo = activatedStep.StepInfo.GroupInfo

	// Fill step info with default step info, if exist
	mergedStep := step
	if activatedStep.StepYMLPath != "" {
		specStep, err := bitrise.ReadSpecStep(activatedStep.StepYMLPath)
		log.Debugf("Spec read from YML: %#v", specStep)
		if err != nil {
			err = fmt.Errorf("parse step.yml of '%s': %s", stepIDData.IDorURI, err)
			return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepDir, activatedStep.ExecutablePath, err)
		}

		// Merge step fields coming from bitrise.yml with the original step fields defined in step.yml
		// For example, a `run_if` can be overridden in a specific workflow.
		mergedStep, err = models.MergeStepWith(specStep, step)
		if err != nil {
			return newActivateStepResult(stepmanModels.StepModel{}, stepInfoPtr, stepDir, activatedStep.ExecutablePath, err)
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

	return newActivateStepResult(mergedStep, stepInfoPtr, stepDir, activatedStep.ExecutablePath, nil)
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
	envSource := &envmanEnv.DefaultEnvironmentSource{}
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
	stepExecutablePath string,
	environments []envmanModels.EnvironmentItemModel,
	secrets []string,
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
			fmt.Errorf("failed to install Step dependency, error: %s", err)
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

	if exit, err := r.executeStep(stepUUID, step, stepIDData, stepDir, stepExecutablePath, bitriseSourceDir, secrets); err != nil {
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
	stepAbsDirPath, stepExecutablePath, bitriseSourceDir string,
	secrets []string,
) (int, error) {
	var cmdArgs []string

	if stepExecutablePath != "" {
		cmdArgs = []string{stepExecutablePath}
	} else {
		toolkitForStep := toolkits.ToolkitForStep(step, r.logger)
		toolkitName := toolkitForStep.ToolkitName()

		prepareResult, prepareErr := toolkitForStep.PrepareForStepRun(step, sIDData, stepAbsDirPath)
		trackToolkitPrepare(r.tracker, stepUUID, toolkitName, sIDData, prepareResult, prepareErr)
		if prepareErr != nil {
			return 1, fmt.Errorf("failed to prepare the step for execution through the required toolkit (%s), error: %s",
				toolkitName, prepareErr)
		}

		cmdFromToolkit, err := toolkitForStep.StepRunCommandArguments(step, sIDData, stepAbsDirPath)
		if err != nil {
			return 1, fmt.Errorf("toolkit (%s) rejected the step, error: %s",
				toolkitName, err)
		}
		cmdArgs = cmdFromToolkit
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

	containerDef, runningContainer := r.containerManager.GetExecutionContainerForStep(stepUUID)
	if containerDef != nil {
		if runningContainer == nil {
			return 1, fmt.Errorf("docker container does not exist")
		}

		envs, err := envman.ReadAndEvaluateEnvs(configs.InputEnvstorePath, &docker.EnvironmentSource{
			Logger: logger,
		})
		if err != nil {
			return 1, fmt.Errorf("failed to read command environment: %w", err)
		}

		name = "docker"
		args = runningContainer.ExecuteCommandArgs(envs)
		args = append(args, cmdArgs...)

		cmd := stepruncmd.New(name, args, bitriseSourceDir, envs, stepSecrets, timeout, noOutputTimeout, stdout, logV2.NewLogger())

		logger.Infof("Step is running in container: %s", containerDef.Image)
		return cmd.Run()
	}

	envs, err := envman.ReadAndEvaluateEnvs(configs.InputEnvstorePath, &envmanEnv.DefaultEnvironmentSource{})
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

func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
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

func logStepStarted(logger log.Logger, stepInfo stepmanModels.StepInfoModel, step stepmanModels.StepModel, idx int, stepExcutionID string, stepStartTime time.Time) {
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
		Toolkit:     toolkits.ToolkitForStep(step, logger).ToolkitName(),
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

// trackToolkitPrepare sends a toolkit prepare telemetry event, but skips no-op toolkits
// (e.g. bash, swift without a precompiled binary) where duration is always 0 and cache_hit
// is always false, which would produce misleading data.
func trackToolkitPrepare(tracker analytics.Tracker, stepUUID, toolkitName string, sIDData stepid.CanonicalID, result toolkits.PrepareForStepRunResult, err error) {
	if err != nil || result.PrepareDuration > 0 {
		tracker.SendToolkitPrepareEvent(stepUUID, toolkitName, sIDData.IDorURI, sIDData.Version, result, err)
	}
}

func secretEnvKeysEnvironment(keys []string) envmanModels.EnvironmentItemModel {
	value := secretkeys.NewManager().Format(keys)
	return envmanModels.EnvironmentItemModel{secretkeys.EnvKey: value}
}
