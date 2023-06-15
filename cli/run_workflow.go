package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/stepoutput"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/envman/env"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	coreanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

func (r WorkflowRunner) runWorkflow(
	plan models.WorkflowExecutionPlan,
	workflowID string,
	workflow models.WorkflowModel,
	steplibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool, tracker analytics.Tracker, buildIDProperties coreanalytics.Properties) models.BuildRunResultsModel {

	workflowIDProperties := coreanalytics.Properties{analytics.WorkflowExecutionID: plan.UUID}
	bitrise.PrintRunningWorkflow(workflow.Title)
	tracker.SendWorkflowStarted(buildIDProperties.Merge(workflowIDProperties), workflowID, workflow.Title)
	*environments = append(*environments, workflow.Environments...)
	results := r.activateAndRunSteps(plan, workflow, steplibSource, buildRunResults, environments, secrets, isLastWorkflow, tracker, workflowIDProperties)
	tracker.SendWorkflowFinished(workflowIDProperties, results.IsBuildFailed())
	return results
}

func (r WorkflowRunner) activateAndRunSteps(
	plan models.WorkflowExecutionPlan,
	workflow models.WorkflowModel,
	defaultStepLibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel,
	secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool,
	tracker analytics.Tracker,
	workflowIDProperties coreanalytics.Properties) models.BuildRunResultsModel {
	log.Debug("[BITRISE_CLI] - Activating and running steps")

	if len(workflow.Steps) == 0 {
		log.Warnf("%s workflow has no steps to run, moving on to the next workflow...", workflow.Title)
		return buildRunResults
	}

	// ------------------------------------------
	// In function global variables - These are global for easy use in local register step run result methods.
	var stepStartTime time.Time
	runResultCollector := newBuildRunResultCollector(tracker)

	// ------------------------------------------
	// Main - Preparing & running the steps
	for idx, stepListItm := range workflow.Steps {
		stepPlan := plan.Steps[idx]
		stepExecutionID := stepPlan.UUID
		stepIDProperties := coreanalytics.Properties{analytics.StepExecutionID: stepExecutionID}
		stepStartedProperties := workflowIDProperties.Merge(stepIDProperties)
		// Per step variables
		stepStartTime = time.Now()
		isLastStep := isLastWorkflow && (idx == len(workflow.Steps)-1)
		// TODO: stepInfoPtr.Step is not a real step, only stores presentation properties (printed in the step boxes)
		stepInfoPtr := stepmanModels.StepInfoModel{}
		stepIdxPtr := idx

		// Per step cleanup
		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
		}

		//
		// Preparing the step
		if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
		}

		if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, *environments); err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
		}

		// Get step id & version data
		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
		}
		stepInfoPtr.ID = compositeStepIDStr
		if workflowStep.Title != nil && *workflowStep.Title != "" {
			stepInfoPtr.Step.Title = pointers.NewStringPtr(*workflowStep.Title)
		} else {
			stepInfoPtr.Step.Title = pointers.NewStringPtr(compositeStepIDStr)
		}

		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
		}
		stepInfoPtr.ID = stepIDData.IDorURI
		if stepInfoPtr.Step.Title == nil || *stepInfoPtr.Step.Title == "" {
			stepInfoPtr.Step.Title = pointers.NewStringPtr(stepIDData.IDorURI)
		}
		stepInfoPtr.Version = stepIDData.Version
		stepInfoPtr.Library = stepIDData.SteplibSource

		//
		// Activating the step
		stepDir := configs.BitriseWorkStepsDirPath

		activator := newStepActivator()
		stepYMLPth, origStepYMLPth, err := activator.activateStep(stepIDData, &buildRunResults, stepDir, configs.BitriseWorkDirPath, &workflowStep, &stepInfoPtr)
		if err != nil {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
			continue
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
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodePreparationFailed, 1, fmt.Errorf("failed to parse step definition (%s): %s", ymlPth, err),
					isLastStep, true, map[string]string{}, stepStartedProperties)
				continue
			}

			mergedStep, err = models.MergeStepWith(specStep, workflowStep)
			if err != nil {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, true, map[string]string{}, stepStartedProperties)
				continue
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

		//
		// Run step
		logStepStarted(stepInfoPtr, mergedStep, idx, stepExecutionID, stepStartTime)

		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			envList, err := tools.EnvmanReadEnvList(configs.InputEnvstorePath)
			if err != nil {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodePreparationFailed, 1, fmt.Errorf("EnvmanReadEnvList failed, err: %s", err),
					isLastStep, false, map[string]string{}, stepStartedProperties)
				continue
			}

			isRun, err := bitrise.EvaluateTemplateToBool(*mergedStep.RunIf, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, envList)
			if err != nil {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodePreparationFailed, 1, err, isLastStep, false, map[string]string{}, stepStartedProperties)
				continue
			}
			if !isRun {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeSkippedWithRunIf, 0, err, isLastStep, false, map[string]string{}, stepStartedProperties)
				continue
			}
		}

		isAlwaysRun := stepmanModels.DefaultIsAlwaysRun
		if mergedStep.IsAlwaysRun != nil {
			isAlwaysRun = *mergedStep.IsAlwaysRun
		} else {
			log.Warnf("Step (%s) mergedStep.IsAlwaysRun is nil, should not!", stepIDData.IDorURI)
		}

		if buildRunResults.IsBuildFailed() && !isAlwaysRun {
			runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
				*mergedStep.RunIf, models.StepRunStatusCodeSkipped, 0, err, isLastStep, false, map[string]string{}, stepStartedProperties)
		} else {
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

			// ensure a new testDirPath and if created successfuly then attach it to the step process by and env
			testDirPath, err := ioutil.TempDir(os.Getenv(configs.BitriseTestDeployDirEnvKey), "test_result")
			if err != nil {
				log.Errorf("Failed to create test result dir, error: %s", err)
			}

			if testDirPath != "" {
				// managed to create the test dir, set the env for it for the next step run
				additionalEnvironments = append(additionalEnvironments, envmanModels.EnvironmentItemModel{
					configs.BitrisePerStepTestResultDirEnvKey: testDirPath,
				})
			}

			environmentItemModels := append(*environments, additionalEnvironments...)
			envSource := &env.DefaultEnvironmentSource{}
			stepDeclaredEnvironments, expandedStepEnvironment, redactedInputsWithType, err := prepareStepEnvironment(prepareStepInputParams{
				environment:       environmentItemModels,
				inputs:            mergedStep.Inputs,
				buildRunResults:   buildRunResults,
				isCIMode:          configs.IsCIMode,
				isPullRequestMode: configs.IsPullRequestMode,
			}, envSource)
			if err != nil {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodePreparationFailed, 1,
					fmt.Errorf("failed to prepare step environment variables: %s", err),
					isLastStep, false, map[string]string{}, stepStartedProperties)
				continue
			}

			stepSecrets := tools.GetSecretValues(secrets)
			if configs.IsSecretEnvsFiltering {
				sensitiveEnvs, err := getSensitiveEnvs(stepDeclaredEnvironments, expandedStepEnvironment)
				if err != nil {
					runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
						*mergedStep.RunIf, models.StepRunStatusCodePreparationFailed, 1,
						fmt.Errorf("failed to get sensitive inputs: %s", err),
						isLastStep, false, map[string]string{}, stepStartedProperties)
					continue
				}

				stepSecrets = append(stepSecrets, tools.GetSecretValues(sensitiveEnvs)...)
			}

			redactedStepInputs, redactedOriginalInputs, err := redactStepInputs(expandedStepEnvironment, mergedStep.Inputs, stepSecrets)
			if err != nil {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodePreparationFailed, 1,
					fmt.Errorf("failed to redact step inputs: %s", err),
					isLastStep, false, map[string]string{}, stepStartedProperties)
				continue
			}

			for key, value := range redactedStepInputs {
				if _, ok := redactedInputsWithType[key]; !ok {
					redactedInputsWithType[key] = value
				}
			}

			tracker.SendStepStartedEvent(stepStartedProperties, prepareAnalyticsStepInfo(mergedStep, stepInfoPtr), redactedInputsWithType, redactedOriginalInputs)

			exit, outEnvironments, err := r.runStep(stepExecutionID, mergedStep, stepIDData, stepDir, stepDeclaredEnvironments, stepSecrets)

			if testDirPath != "" {
				if err := addTestMetadata(testDirPath, models.TestResultStepInfo{Number: idx, Title: *mergedStep.Title, ID: stepIDData.IDorURI, Version: stepIDData.Version}); err != nil {
					log.Errorf("Failed to normalize test result dir, error: %s", err)
				}
			}

			if err := tools.EnvmanClear(configs.OutputEnvstorePath); err != nil {
				log.Errorf("Failed to clear output envstore, error: %s", err)
			}

			*environments = append(*environments, outEnvironments...)
			if err != nil {
				if *mergedStep.IsSkippable {
					runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
						*mergedStep.RunIf, models.StepRunStatusCodeFailedSkippable, exit, err, isLastStep, false, redactedStepInputs, stepIDProperties)
				} else {
					runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
						*mergedStep.RunIf, models.StepRunStatusCodeFailed, exit, err, isLastStep, false, redactedStepInputs, stepIDProperties)
				}
			} else {
				runResultCollector.registerStepRunResults(&buildRunResults, stepExecutionID, stepStartTime, mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeSuccess, 0, nil, isLastStep, false, redactedStepInputs, stepIDProperties)
			}
		}
	}

	return buildRunResults
}

func (r WorkflowRunner) runStep(
	stepUUID string,
	step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string,
	environments []envmanModels.EnvironmentItemModel, secrets []string) (int, []envmanModels.EnvironmentItemModel, error) {
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

	if exit, err := r.executeStep(stepUUID, step, stepIDData, stepDir, bitriseSourceDir, secrets); err != nil {
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
	step stepmanModels.StepModel, sIDData models.StepIDData,
	stepAbsDirPath, bitriseSourceDir string,
	secrets []string) (int, error) {
	toolkitForStep := toolkits.ToolkitForStep(step)
	toolkitName := toolkitForStep.ToolkitName()

	if err := toolkitForStep.PrepareForStepRun(step, sIDData, stepAbsDirPath); err != nil {
		return 1, fmt.Errorf("Failed to prepare the step for execution through the required toolkit (%s), error: %s",
			toolkitName, err)
	}

	cmd, err := toolkitForStep.StepRunCommandArguments(step, sIDData, stepAbsDirPath)
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
	writer := stepoutput.NewWriter(stepSecrets, opts)

	return tools.EnvmanRun(
		configs.InputEnvstorePath,
		bitriseSourceDir,
		cmd,
		timeout,
		noOutputTimeout,
		nil,
		writer)
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

func logStepStarted(stepInfo stepmanModels.StepInfoModel, step stepmanModels.StepModel, idx int, stepExcutionId string, stepStartTime time.Time) {
	title := ""
	if stepInfo.Step.Title != nil && *stepInfo.Step.Title != "" {
		title = *stepInfo.Step.Title
	}

	params := log.StepStartedParams{
		ExecutionId: stepExcutionId,
		Position:    idx,
		Title:       title,
		Id:          stepInfo.ID,
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

func isDirEmpty(path string) (bool, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
