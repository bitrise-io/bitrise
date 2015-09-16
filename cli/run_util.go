package cli

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
	"github.com/ryanuber/go-glob"
)

// GetWorkflowIDByPattern ...
func GetWorkflowIDByPattern(config models.BitriseDataModel, pattern string) (string, error) {
	// Check for workflow ID in trigger map
	for _, item := range config.TriggerMap {
		if glob.Glob(item.Pattern, pattern) {
			if !item.IsPullRequestAllowed && IsPullRequestMode {
				return "", fmt.Errorf("Trigger pattern (%s) match found, but pull request is not enabled", pattern)
			}
			return item.WorkflowID, nil
		}
	}

	// Check for direct workflow selection
	_, exist := config.Workflows[pattern]
	if !exist {
		return "", fmt.Errorf("Specified Workflow (%s) does not exist!", pattern)
	} else if IsPullRequestMode {
		return "", fmt.Errorf("Run triggered by pull request (pattern: %s), but no matching pattern found", pattern)
	}

	return pattern, nil
}

// GetBitriseConfigFromBase64Data ...
func GetBitriseConfigFromBase64Data(configBase64Str string) (models.BitriseDataModel, error) {
	configBase64Bytes, err := base64.StdEncoding.DecodeString(configBase64Str)
	if err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("Failed to decode base 64 string, error: %s", err)
	}

	config, err := bitrise.ConfigModelFromYAMLBytes(configBase64Bytes)
	if err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("Failed to parse bitrise config, error: %s", err)
	}

	return config, nil
}

// GetBitriseConfigFilePath ...
func GetBitriseConfigFilePath(c *cli.Context) (string, error) {
	bitriseConfigPath := c.String(ConfigKey)

	if bitriseConfigPath == "" {
		bitriseConfigPath = c.String(PathKey)
		if bitriseConfigPath != "" {
			log.Warn("'path' key is deprecated, use 'config' instead!")
		}
	}

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

// CreateBitriseConfigFromCLIParams ...
func CreateBitriseConfigFromCLIParams(c *cli.Context) (models.BitriseDataModel, error) {
	bitriseConfig := models.BitriseDataModel{}

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	if bitriseConfigBase64Data != "" {
		config, err := GetBitriseConfigFromBase64Data(bitriseConfigBase64Data)
		if err != nil {
			return models.BitriseDataModel{}, fmt.Errorf("Failed to get config (bitrise.yml) from base 64 data, err: %s", err)
		}
		bitriseConfig = config
	} else {
		bitriseConfigPath, err := GetBitriseConfigFilePath(c)
		if err != nil {
			return models.BitriseDataModel{}, fmt.Errorf("Failed to get config (bitrise.yml) path: %s", err)
		}
		if bitriseConfigPath == "" {
			return models.BitriseDataModel{}, errors.New("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
		}

		config, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
		if err != nil {
			return models.BitriseDataModel{}, fmt.Errorf("Failed to validate config: %s", err)
		}
		bitriseConfig = config
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI model version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		return models.BitriseDataModel{}, fmt.Errorf("Failed to compare bitrise CLI models's version with the bitrise.yml FormatVersion: %s", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		return models.BitriseDataModel{}, errors.New("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	return bitriseConfig, nil
}

// GetInventoryFromBase64Data ...
func GetInventoryFromBase64Data(inventoryBase64Str string) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryBase64Bytes, err := base64.StdEncoding.DecodeString(inventoryBase64Str)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to decode base 64 string, error: %s", err)
	}

	var envstore envmanModels.EnvsYMLModel
	if err := yaml.Unmarshal(inventoryBase64Bytes, &envstore); err != nil {
		return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to unmasrhal bitrise inventory, error: %s", err)
	}

	for _, env := range envstore.Envs {
		if err := env.Normalize(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to normalize bitrise inventory, error: %s", err)
		}
		if err := env.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to fill bitrise inventory, error: %s", err)
		}
		if err := env.Validate(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to validate bitrise inventory, error: %s", err)
		}
	}

	return envstore.Envs, nil
}

// GetInventoryFilePath ...
func GetInventoryFilePath(c *cli.Context) (string, error) {
	inventoryPath := c.String(InventoryKey)

	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = path.Join(bitrise.CurrentDir, DefaultSecretsFileName)

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			return "", err
		} else if !exist {
			inventoryPath = ""
		}
	}

	return inventoryPath, nil
}

// CreateInventoryFromCLIParams ...
func CreateInventoryFromCLIParams(c *cli.Context) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryEnvironments := []envmanModels.EnvironmentItemModel{}

	inventoryBase64Data := c.String(InventoryBase64Key)
	if inventoryBase64Data != "" {
		inventory, err := GetInventoryFromBase64Data(inventoryBase64Data)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to get inventory from base 64 data, err: %s", err)
		}
		inventoryEnvironments = inventory
	} else {
		inventoryPath, err := GetInventoryFilePath(c)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to get inventory path: %s", err)
		}

		if inventoryPath != "" {
			var err error
			inventory, err := bitrise.CollectEnvironmentsFromFile(inventoryPath)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Invalid invetory format: %s", err)
			}
			inventoryEnvironments = inventory
		}
	}

	return inventoryEnvironments, nil
}

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string, environments []envmanModels.EnvironmentItemModel) (int, []envmanModels.EnvironmentItemModel, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check dependencies
	for _, dep := range step.Dependencies {
		isSkippedBecauseOfPlatform := false
		switch dep.Manager {
		case depManagerBrew:
			if runtime.GOOS == "darwin" {
				err := bitrise.InstallWithBrewIfNeeded(dep.Name, IsCIMode)
				if err != nil {
					return 1, []envmanModels.EnvironmentItemModel{}, err
				}
			} else {
				isSkippedBecauseOfPlatform = true
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

		if isSkippedBecauseOfPlatform {
			log.Debugf(" * Dependency (%s) skipped, manager (%s) not supported on this platform (%s)", dep.Name, dep.Manager, runtime.GOOS)
		} else {
			log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", dep.Name)
		}
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

func getStepID(stepListItem models.StepListItemModel) (string, error) {
	stepIDDataStr, step, err := models.GetStepIDStepDataPair(stepListItem)
	if err != nil {
		return "", err
	}

	ID := stepIDDataStr
	if step.Title != nil && *step.Title != "" {
		ID = *step.Title
	}
	return ID, nil
}

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel, isLastWorkflow bool) models.BuildRunResultsModel {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	var stepStartTime time.Time

	// hold pointer to current step info, for easy usage in local register step run result methods
	// the value is filled with the current running step info
	var stepInfoPtr models.StepInfoModel

	registerStepRunResults := func(step stepmanModels.StepModel, resultCode, exitCode int, err error, isLastStep bool) {
		stepInfoCopy := models.StepInfoModel{
			ID:      stepInfoPtr.ID,
			Version: stepInfoPtr.Version,
			Latest:  stepInfoPtr.Latest,
		}

		stepResults := models.StepRunResultsModel{
			StepInfo: stepInfoCopy,
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
			log.Errorf("Step (%s) failed, error: (%v)", stepInfoCopy.ID, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case models.StepRunStatusCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", stepInfoCopy.ID, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case models.StepRunStatusCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", stepInfoCopy.ID)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case models.StepRunStatusCodeSkippedWithRunIf:
			log.Warn("The step's (" + stepInfoCopy.ID + ") Run-If expression evaluated to false - skipping")
			log.Info("The Run-If expression was: ", colorstring.Blue(*step.RunIf))
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(stepResults, isLastStep)
	}

	registerStepListItemRunResults := func(stepListItem models.StepListItemModel, resultCode, exitCode int, err error, isLastStep bool) {
		stepInfoCopy := models.StepInfoModel{
			ID:      stepInfoPtr.ID,
			Version: stepInfoPtr.Version,
			Latest:  stepInfoPtr.Latest,
		}

		stepResults := models.StepRunResultsModel{
			StepInfo: stepInfoCopy,
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
			log.Errorf("Step (%s) failed, error: (%v)", stepInfoCopy.ID, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case models.StepRunStatusCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", stepInfoCopy.ID, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case models.StepRunStatusCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", stepInfoCopy.ID)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case models.StepRunStatusCodeSkippedWithRunIf:
			log.Warn("The step's (" + stepInfoCopy.ID + ") Run-If expression evaluated to false - skipping")
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(stepResults, isLastStep)
	}

	for idx, stepListItm := range workflow.Steps {
		stepStartTime = time.Now()
		isLastStep := isLastWorkflow && (idx == len(workflow.Steps)-1)
		stepInfoPtr = models.StepInfoModel{}

		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}

		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		stepInfoPtr.ID = compositeStepIDStr
		if err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		stepInfoPtr.ID = stepIDData.IDorURI
		stepInfoPtr.Version = stepIDData.Version
		if err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
			continue
		}
		stepVersionForInfoPrint := stepIDData.Version

		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := path.Join(bitrise.BitriseWorkDirPath, "current_step.yml")

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)

			if err := cmdex.CopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepAbsLocalPth, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			if len(stepIDData.IDorURI) > 20 {
				stepVersionForInfoPrint = fmt.Sprintf("path:...%s", stringutil.MaxLastChars(stepIDData.IDorURI, 17))
			} else {
				stepVersionForInfoPrint = fmt.Sprintf("path:%s", stepIDData.IDorURI)
			}
			stepInfoPtr.Version = stepVersionForInfoPrint
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			if err := cmdex.CopyFile(path.Join(stepDir, "step.yml"), stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			stepVersionForInfoPrint = fmt.Sprintf("git:...%s", stringutil.MaxLastChars(stepIDData.IDorURI, 10))
			if stepIDData.Version != "" {
				stepVersionForInfoPrint = stepVersionForInfoPrint + "@" + stepIDData.Version
			}
			stepInfoPtr.Version = stepVersionForInfoPrint
		} else if stepIDData.SteplibSource == "_" {
			log.Debugf("[BITRISE_CLI] - Steplib independent step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

			// Steplib independent steps are completly defined in workflow
			stepYMLPth = ""
			if err := workflowStep.FillMissingDefaults(); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			stepVersionForInfoPrint = "_"
			if stepIDData.Version != "" {
				if len(stepIDData.Version) > 20 {
					stepVersionForInfoPrint = stepVersionForInfoPrint + "@..." + stringutil.MaxLastChars(stepIDData.Version, 17)
				} else {
					stepVersionForInfoPrint = stepVersionForInfoPrint + "@" + stepIDData.Version
				}
			}
			stepInfoPtr.Version = stepVersionForInfoPrint
		} else if stepIDData.SteplibSource != "" {
			log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.StepmanSetup(stepIDData.SteplibSource); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			stepInfo, err := bitrise.StepmanStepInfo(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err != nil {
				if buildRunResults.IsStepLibUpdated(stepIDData.SteplibSource) {
					registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
					continue
				}
				// May StepLib should be updated
				log.Info("Step info not found in StepLib (%s) -- Updating ...")
				if err := bitrise.StepmanUpdate(stepIDData.SteplibSource); err != nil {
					registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
					continue
				}
				buildRunResults.StepmanUpdates[stepIDData.SteplibSource]++
				stepInfo, err = bitrise.StepmanStepInfo(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
				if err != nil {
					registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
					continue
				}
			}

			stepVersionForInfoPrint = stepInfo.Version
			stepInfoPtr.Version = stepVersionForInfoPrint

			if err := bitrise.StepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			}
		} else {
			registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData), isLastStep)
			continue
		}

		mergedStep := workflowStep
		if stepYMLPth != "" {
			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}

			mergedStep, err = models.MergeStepWith(specStep, workflowStep)
			if err != nil {
				registerStepListItemRunResults(stepListItm, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}
		}

		// Run step
		bitrise.PrintRunningStep(stepInfoPtr, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, buildRunResults)
			if err != nil {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeFailed, 1, err, isLastStep)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeSkippedWithRunIf, 0, err, isLastStep)
				continue
			}
		}
		outEnvironments := []envmanModels.EnvironmentItemModel{}

		isAlwaysRun := stepmanModels.DefaultIsAlwaysRun
		if mergedStep.IsAlwaysRun != nil {
			isAlwaysRun = *mergedStep.IsAlwaysRun
		} else {
			log.Warn("Step (%s) mergedStep.IsAlwaysRun is nil, should not!", stepIDData.IDorURI)
		}

		if buildRunResults.IsBuildFailed() && !isAlwaysRun {
			registerStepRunResults(mergedStep, models.StepRunStatusCodeSkipped, 0, err, isLastStep)
		} else {
			exit, out, err := runStep(mergedStep, stepIDData, stepDir, *environments)
			outEnvironments = out
			if err != nil {
				if *mergedStep.IsSkippable {
					registerStepRunResults(mergedStep, models.StepRunStatusCodeFailedSkippable, exit, err, isLastStep)
				} else {
					registerStepRunResults(mergedStep, models.StepRunStatusCodeFailed, exit, err, isLastStep)
				}
			} else {
				registerStepRunResults(mergedStep, models.StepRunStatusCodeSuccess, 0, nil, isLastStep)
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

func activateAndRunWorkflow(workflowID string, workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel, buildRunResults models.BuildRunResultsModel, environments *[]envmanModels.EnvironmentItemModel, lastWorkflowID string) (models.BuildRunResultsModel, error) {
	var err error
	// Run these workflows before running the target workflow
	for _, beforeWorkflowID := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowID]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist!", beforeWorkflowID)
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowID
		}
		buildRunResults, err = activateAndRunWorkflow(beforeWorkflowID, beforeWorkflow, bitriseConfig, buildRunResults, environments, lastWorkflowID)
		if err != nil {
			return buildRunResults, err
		}
	}

	// Run the target workflow
	isLastWorkflow := (workflowID == lastWorkflowID)
	buildRunResults = runWorkflow(workflow, bitriseConfig.DefaultStepLibSource, buildRunResults, environments, isLastWorkflow)

	// Run these workflows after running the target workflow
	for _, afterWorkflowID := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowID]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist!", afterWorkflowID)
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowID
		}
		buildRunResults, err = activateAndRunWorkflow(afterWorkflowID, afterWorkflow, bitriseConfig, buildRunResults, environments, lastWorkflowID)
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

// RunWorkflowWithConfiguration ...
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
		StartTime:      startTime,
		StepmanUpdates: map[string]int{},
	}

	environments = append(environments, workflowToRun.Environments...)

	lastWorkflowID, err := lastWorkflowIDInConfig(workflowToRunID, bitriseConfig)
	if err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to get last workflow id: %s", err)
	}

	buildRunResults, err = activateAndRunWorkflow(workflowToRunID, workflowToRun, bitriseConfig, buildRunResults, &environments, lastWorkflowID)
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
