package cli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/bitrise/tools"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/versions"
	stepmanCLI "github.com/bitrise-io/stepman/cli"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

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

	if isPRMode {
		log.Info(colorstring.Yellow("bitrise runs in PR mode"))
	}
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

	if isCIMode {
		log.Info(colorstring.Yellow("bitrise runs in CI mode"))
	}
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

	if filtering {
		log.Info(colorstring.Yellow("bitrise runs in Secret Filtering mode"))
	}
	return os.Setenv(configs.IsSecretFilteringKey, strconv.FormatBool(filtering))
}

func isDirEmpty(path string) (bool, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// GetBitriseConfigFromBase64Data ...
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

// GetBitriseConfigFilePath ...
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

// CreateBitriseConfigFromCLIParams ...
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

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI model version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		return models.BitriseDataModel{}, warnings, fmt.Errorf("Failed to compare bitrise CLI models's version with the bitrise.yml FormatVersion: %s", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		return models.BitriseDataModel{}, warnings, errors.New("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml")
	}

	return bitriseConfig, warnings, nil
}

// GetInventoryFromBase64Data ...
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

// GetInventoryFilePath ...
func GetInventoryFilePath(inventoryPath string) (string, error) {
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = filepath.Join(configs.CurrentDir, DefaultSecretsFileName)

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			return "", err
		} else if !exist {
			inventoryPath = ""
		}
	}

	return inventoryPath, nil
}

// CreateInventoryFromCLIParams ...
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

	if step.Deps != nil && (len(step.Deps.Brew) > 0 || len(step.Deps.AptGet) > 0 || len(step.Deps.CheckOnly) > 0) {
		//
		// New dependency handling
		for _, checkOnlyDep := range step.Deps.CheckOnly {
			if err := bitrise.DependencyTryCheckTool(checkOnlyDep.Name); err != nil {
				return err
			}
			log.Infof(" * "+colorstring.Green("[OK]")+" Step dependency (%s) installed, available.", checkOnlyDep.Name)
		}

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
			return errors.New("Unsupported os")
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
			case depManagerTryCheck:
				err := bitrise.DependencyTryCheckTool(dep.Name)
				if err != nil {
					return err
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

func executeStep(
	step stepmanModels.StepModel, sIDData models.StepIDData,
	stepAbsDirPath, bitriseSourceDir string,
	secrets []envmanModels.EnvironmentItemModel) (int, error) {
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

	return tools.EnvmanRun(configs.InputEnvstorePath, bitriseSourceDir, cmd, timeout, secrets)
}

func runStep(
	step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string,
	environments []envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	buildRunResults models.BuildRunResultsModel) (int, []envmanModels.EnvironmentItemModel, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check & Install Step Dependencies
	// [!] Make sure this happens BEFORE the Toolkit Bootstrap,
	// so that if a Toolkit requires/allows the use of additional dependencies
	// required for the step (e.g. a brew installed OpenSSH) it can be done
	// with a Toolkit+Deps
	if err := retry.Times(2).Try(func(attempt uint) error {
		if attempt > 0 {
			fmt.Println()
			log.Warn("Installing Step dependency failed, retrying ...")
		}

		return checkAndInstallStepDependencies(step)
	}); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to install Step dependency, error: %s", err)
	}

	// Collect step inputs
	if err := tools.EnvmanInitAtPath(configs.InputEnvstorePath); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to init envman for the Step, error: %s", err)
	}

	if err := tools.ExportEnvironmentsList(configs.InputEnvstorePath, environments); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, fmt.Errorf("Failed to export environment list for the Step, error: %s", err)
	}

	evaluatedInputs := []envmanModels.EnvironmentItemModel{}
	for _, input := range step.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, err
		}

		options, err := input.GetOptions()
		if err != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, err
		}

		if options.IsTemplate != nil && *options.IsTemplate {
			outStr, err := tools.EnvmanJSONPrint(configs.InputEnvstorePath)
			if err != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, fmt.Errorf("EnvmanJSONPrint failed, err: %s", err)
			}

			envList, err := envmanModels.NewEnvJSONList(outStr)
			if err != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, fmt.Errorf("CreateFromJSON failed, err: %s", err)
			}

			evaluatedValue, err := bitrise.EvaluateTemplateToString(value, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, envList)
			if err != nil {
				return 1, []envmanModels.EnvironmentItemModel{}, err
			}

			input[key] = evaluatedValue
		}

		evaluatedInputs = append(evaluatedInputs, input)
	}
	environments = append(environments, evaluatedInputs...)

	if err := tools.EnvmanInitAtPath(configs.InputEnvstorePath); err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	if err := tools.ExportEnvironmentsList(configs.InputEnvstorePath, environments); err != nil {
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

	if exit, err := executeStep(step, stepIDData, stepDir, bitriseSourceDir, secrets); err != nil {
		stepOutputs, envErr := bitrise.CollectEnvironmentsFromFile(configs.OutputEnvstorePath)
		if envErr != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, envErr
		}

		updatedStepOutputs, updateErr := bitrise.ApplyOutputAliases(stepOutputs, step.Outputs)
		if updateErr != nil {
			return 1, []envmanModels.EnvironmentItemModel{}, updateErr
		}

		return exit, updatedStepOutputs, err
	}

	stepOutputs, err := bitrise.CollectEnvironmentsFromFile(configs.OutputEnvstorePath)
	if err != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, err
	}

	updatedStepOutputs, updateErr := bitrise.ApplyOutputAliases(stepOutputs, step.Outputs)
	if updateErr != nil {
		return 1, []envmanModels.EnvironmentItemModel{}, updateErr
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	return 0, updatedStepOutputs, nil
}

func activateStepLibStep(stepIDData models.StepIDData, destination, stepYMLCopyPth string, isStepLibUpdated bool) (stepmanModels.StepInfoModel, bool, error) {
	didStepLibUpdate := false

	log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
	if err := stepman.SetupLibrary(stepIDData.SteplibSource); err != nil {
		return stepmanModels.StepInfoModel{}, false, err
	}

	versionConstraint, err := stepmanModels.ParseRequiredVersion(stepIDData.Version)
	if err != nil {
		return stepmanModels.StepInfoModel{}, false,
			fmt.Errorf("activating step (%s) from source (%s) failed, invalid version specified: %s", stepIDData.IDorURI, stepIDData.SteplibSource, err)
	}
	if versionConstraint.VersionLockType == stepmanModels.InvalidVersionConstraint {
		return stepmanModels.StepInfoModel{}, false,
			fmt.Errorf("activating step (%s) from source (%s) failed, version constraint is invalid", stepIDData.IDorURI, stepIDData.SteplibSource)
	}

	isStepLibUpdateNeeded := (versionConstraint.VersionLockType == stepmanModels.Latest) ||
		(versionConstraint.VersionLockType == stepmanModels.MinorLocked) ||
		(versionConstraint.VersionLockType == stepmanModels.MajorLocked)
	if !isStepLibUpdated && isStepLibUpdateNeeded {
		log.Infof("Step uses latest version -- Updating StepLib ...")
		if _, err := stepman.UpdateLibrary(stepIDData.SteplibSource); err != nil {
			log.Warnf("Step version constraint is latest or version locked, but failed to update StepLib, err: %s", err)
		} else {
			didStepLibUpdate = true
		}
	}

	info, err := stepmanCLI.QueryStepInfoFromLibrary(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
	if err != nil {
		if isStepLibUpdated {
			return stepmanModels.StepInfoModel{}, didStepLibUpdate, fmt.Errorf("stepman JSON steplib step info failed: %s", err)
		}

		// May StepLib should be updated
		log.Infof("Step info not found in StepLib (%s) -- Updating ...", stepIDData.SteplibSource)
		if _, err := stepman.UpdateLibrary(stepIDData.SteplibSource); err != nil {
			return stepmanModels.StepInfoModel{}, didStepLibUpdate, err
		}

		didStepLibUpdate = true

		info, err = stepmanCLI.QueryStepInfoFromLibrary(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
		if err != nil {
			return stepmanModels.StepInfoModel{}, didStepLibUpdate, fmt.Errorf("stepman JSON steplib step info failed: %s", err)
		}
	}

	if info.Step.Title == nil || *info.Step.Title == "" {
		info.Step.Title = pointers.NewStringPtr(info.ID)
	}
	info.OriginalVersion = stepIDData.Version

	if err := stepmanCLI.Activate(stepIDData.SteplibSource, stepIDData.IDorURI, info.Version, destination, stepYMLCopyPth, false); err != nil {
		return stepmanModels.StepInfoModel{}, didStepLibUpdate, err
	}
	log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)

	return info, didStepLibUpdate, nil
}

func activateAndRunSteps(
	workflow models.WorkflowModel,
	defaultStepLibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool) models.BuildRunResultsModel {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	// ------------------------------------------
	// In function global variables - These are global for easy use in local register step run result methods.
	var stepStartTime time.Time

	// ------------------------------------------
	// In function method - Registration methods, for register step run results.
	registerStepRunResults := func(step stepmanModels.StepModel, stepInfoPtr stepmanModels.StepInfoModel,
		stepIdxPtr int, runIf string, resultCode, exitCode int, err error, isLastStep, printStepHeader bool) {

		if printStepHeader {
			bitrise.PrintRunningStepHeader(stepInfoPtr, step, stepIdxPtr)
		}

		stepInfoCopy := stepmanModels.StepInfoModel{
			Library:         stepInfoPtr.Library,
			ID:              stepInfoPtr.ID,
			Version:         stepInfoPtr.Version,
			OriginalVersion: stepInfoPtr.OriginalVersion,
			LatestVersion:   stepInfoPtr.LatestVersion,
			GroupInfo:       stepInfoPtr.GroupInfo,
			Step:            stepInfoPtr.Step,
			DefinitionPth:   stepInfoPtr.DefinitionPth,
		}

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		stepResults := models.StepRunResultsModel{
			StepInfo:  stepInfoCopy,
			Status:    resultCode,
			Idx:       buildRunResults.ResultsCount(),
			RunTime:   time.Now().Sub(stepStartTime),
			ErrorStr:  errStr,
			ExitCode:  exitCode,
			StartTime: stepStartTime,
		}

		isExitStatusError := true
		if err != nil {
			isExitStatusError = errorutil.IsExitStatusError(err)
		}

		switch resultCode {
		case models.StepRunStatusCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case models.StepRunStatusCodeFailed:
			if !isExitStatusError {
				log.Errorf("Step (%s) failed: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
			}

			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case models.StepRunStatusCodeFailedSkippable:
			if !isExitStatusError {
				log.Warnf("Step (%s) failed, but was marked as skippable: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
			} else {
				log.Warnf("Step (%s) failed, but was marked as skippable", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"))
			}

			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case models.StepRunStatusCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"))

			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case models.StepRunStatusCodeSkippedWithRunIf:
			log.Warn("The step's (" + pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title") + ") Run-If expression evaluated to false - skipping")
			if runIf != "" {
				log.Info("The Run-If expression was: ", colorstring.Blue(runIf))
			}

			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unknown result code")
			return
		}

		bitrise.PrintRunningStepFooter(stepResults, isLastStep)
	}

	// ------------------------------------------
	// Main - Preparing & running the steps
	for idx, stepListItm := range workflow.Steps {
		// Per step variables
		stepStartTime = time.Now()
		isLastStep := isLastWorkflow && (idx == len(workflow.Steps)-1)
		stepInfoPtr := stepmanModels.StepInfoModel{}
		stepIdxPtr := idx

		// Per step cleanup
		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
			continue
		}

		//
		// Preparing the step
		if err := tools.EnvmanInitAtPath(configs.InputEnvstorePath); err != nil {
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
			continue
		}

		if err := tools.ExportEnvironmentsList(configs.InputEnvstorePath, *environments); err != nil {
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
			continue
		}

		// Get step id & version data
		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
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
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
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
		stepYMLPth := filepath.Join(configs.BitriseWorkDirPath, "current_step.yml")
		var origStepYMLPth string

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)

			origStepYMLPth = filepath.Join(stepAbsLocalPth, "step.yml")
			if err := command.CopyFile(origStepYMLPth, stepYMLPth); err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}

			if err := command.CopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			repo, err := git.New(stepDir)
			if err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
			}
			if err := repo.CloneTagOrBranch(stepIDData.IDorURI, stepIDData.Version).Run(); err != nil {
				if strings.HasPrefix(stepIDData.IDorURI, "git@") {
					fmt.Println(colorstring.Yellow(`Note: if the step's repository is an open source one,`))
					fmt.Println(colorstring.Yellow(`you should probably use a "https://..." git clone URL,`))
					fmt.Println(colorstring.Yellow(`instead of the "git@..." git clone URL which usually requires authentication`))
					fmt.Println(colorstring.Yellow(`even if the repository is open source!`))
				}
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}

			if err := command.CopyFile(filepath.Join(stepDir, "step.yml"), stepYMLPth); err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}
		} else if stepIDData.SteplibSource == "_" {
			log.Debugf("[BITRISE_CLI] - Steplib independent step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

			// Steplib independent steps are completly defined in workflow
			stepYMLPth = ""
			if err := workflowStep.FillMissingDefaults(); err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}

			repo, err := git.New(stepDir)
			if err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
			}
			if err := repo.CloneTagOrBranch(stepIDData.IDorURI, stepIDData.Version).Run(); err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			isUpdated := buildRunResults.IsStepLibUpdated(stepIDData.SteplibSource)
			stepInfo, didUpdate, err := activateStepLibStep(stepIDData, stepDir, stepYMLPth, isUpdated)
			if didUpdate {
				buildRunResults.StepmanUpdates[stepIDData.SteplibSource]++
			}

			stepInfoPtr.ID = stepInfo.ID
			if stepInfoPtr.Step.Title == nil || *stepInfoPtr.Step.Title == "" {
				stepInfoPtr.Step.Title = pointers.NewStringPtr(stepInfo.ID)
			}
			stepInfoPtr.Version = stepInfo.Version
			stepInfoPtr.LatestVersion = stepInfo.LatestVersion
			stepInfoPtr.OriginalVersion = stepInfo.OriginalVersion
			stepInfoPtr.GroupInfo = stepInfo.GroupInfo

			if err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}
		} else {
			registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
				"", models.StepRunStatusCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData), isLastStep, true)
			continue
		}

		// Fill step info with default step info, if exist
		mergedStep := workflowStep
		if stepYMLPth != "" {
			specStep, err := bitrise.ReadSpecStep(stepYMLPth)
			log.Debugf("Spec read from YML: %#v\n", specStep)
			if err != nil {
				ymlPth := stepYMLPth
				if origStepYMLPth != "" {
					// in case of local step (path:./) we use the original step definition path,
					// instead of the activated step's one.
					ymlPth = origStepYMLPth
				}
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, fmt.Errorf("failed to parse step definition (%s): %s", ymlPth, err), isLastStep, true)
				continue
			}

			mergedStep, err = models.MergeStepWith(specStep, workflowStep)
			if err != nil {
				registerStepRunResults(stepmanModels.StepModel{}, stepInfoPtr, stepIdxPtr,
					"", models.StepRunStatusCodeFailed, 1, err, isLastStep, true)
				continue
			}
		}

		if mergedStep.SupportURL != nil {
			stepInfoPtr.Step.SupportURL = pointers.NewStringPtr(*mergedStep.SupportURL)
		}
		if mergedStep.SourceCodeURL != nil {
			stepInfoPtr.Step.SourceCodeURL = pointers.NewStringPtr(*mergedStep.SourceCodeURL)
		}

		//
		// Run step
		bitrise.PrintRunningStepHeader(stepInfoPtr, mergedStep, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			outStr, err := tools.EnvmanJSONPrint(configs.InputEnvstorePath)
			if err != nil {
				registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeFailed, 1, fmt.Errorf("EnvmanJSONPrint failed, err: %s", err), isLastStep, false)
				continue
			}

			envList, err := envmanModels.NewEnvJSONList(outStr)
			if err != nil {
				registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeFailed, 1, fmt.Errorf("CreateFromJSON failed, err: %s", err), isLastStep, false)
				continue
			}

			isRun, err := bitrise.EvaluateTemplateToBool(*mergedStep.RunIf, configs.IsCIMode, configs.IsPullRequestMode, buildRunResults, envList)
			if err != nil {
				registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeFailed, 1, err, isLastStep, false)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeSkippedWithRunIf, 0, err, isLastStep, false)
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
			registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
				*mergedStep.RunIf, models.StepRunStatusCodeSkipped, 0, err, isLastStep, false)
		} else {
			// beside of the envs coming from the current parent process these will be added as an extra
			var additionalEnvironments []envmanModels.EnvironmentItemModel

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

			exit, outEnvironments, err := runStep(
				mergedStep, stepIDData, stepDir,
				append(*environments, additionalEnvironments...), secrets,
				buildRunResults,
			)

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
					registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
						*mergedStep.RunIf, models.StepRunStatusCodeFailedSkippable, exit, err, isLastStep, false)
				} else {
					registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
						*mergedStep.RunIf, models.StepRunStatusCodeFailed, exit, err, isLastStep, false)
				}
			} else {
				registerStepRunResults(mergedStep, stepInfoPtr, stepIdxPtr,
					*mergedStep.RunIf, models.StepRunStatusCodeSuccess, 0, nil, isLastStep, false)
			}
		}
	}

	return buildRunResults
}

func runWorkflow(
	workflow models.WorkflowModel,
	steplibSource string,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	isLastWorkflow bool) models.BuildRunResultsModel {
	bitrise.PrintRunningWorkflow(workflow.Title)

	*environments = append(*environments, workflow.Environments...)
	return activateAndRunSteps(workflow, steplibSource, buildRunResults, environments, secrets, isLastWorkflow)
}

func activateAndRunWorkflow(
	workflowID string, workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	lastWorkflowID string) (models.BuildRunResultsModel, error) {
	var err error
	// Run these workflows before running the target workflow
	for _, beforeWorkflowID := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowID]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist", beforeWorkflowID)
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowID
		}
		buildRunResults, err = activateAndRunWorkflow(
			beforeWorkflowID, beforeWorkflow, bitriseConfig,
			buildRunResults,
			environments, secrets,
			lastWorkflowID)
		if err != nil {
			return buildRunResults, err
		}
	}

	// Run the target workflow
	isLastWorkflow := (workflowID == lastWorkflowID)
	buildRunResults = runWorkflow(
		workflow, bitriseConfig.DefaultStepLibSource,
		buildRunResults,
		environments, secrets,
		isLastWorkflow)

	// Run these workflows after running the target workflow
	for _, afterWorkflowID := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowID]
		if !exist {
			return buildRunResults, fmt.Errorf("Specified Workflow (%s) does not exist", afterWorkflowID)
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowID
		}
		buildRunResults, err = activateAndRunWorkflow(
			afterWorkflowID, afterWorkflow, bitriseConfig,
			buildRunResults,
			environments, secrets,
			lastWorkflowID)
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

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunID]
	if !exist {
		return models.BuildRunResultsModel{}, fmt.Errorf("Specified Workflow (%s) does not exist", workflowToRunID)
	}

	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunID
	}

	// Envman setup
	if err := os.Setenv(configs.EnvstorePathEnvKey, configs.OutputEnvstorePath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := os.Setenv(configs.FormattedOutputPathEnvKey, configs.FormattedOutputPath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := tools.EnvmanInit(); err != nil {
		return models.BuildRunResultsModel{}, errors.New("Failed to run envman init")
	}

	// App level environment
	environments := append(secretEnvironments, bitriseConfig.App.Environments...)

	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_ID", workflowToRunID); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_ID env: %s", err)
	}
	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_TITLE", workflowToRun.Title); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_TITLE env: %s", err)
	}

	environments = append(environments, workflowToRun.Environments...)

	lastWorkflowID, err := lastWorkflowIDInConfig(workflowToRunID, bitriseConfig)
	if err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to get last workflow id: %s", err)
	}

	// Bootstrap Toolkits
	for _, aToolkit := range toolkits.AllSupportedToolkits() {
		toolkitName := aToolkit.ToolkitName()
		if !aToolkit.IsToolAvailableInPATH() {
			// don't bootstrap if any preinstalled version is available,
			// the toolkit's `PrepareForStepRun` can bootstrap for itself later if required
			// or if the system installed version is not sufficient
			if err := aToolkit.Bootstrap(); err != nil {
				return models.BuildRunResultsModel{}, fmt.Errorf("Failed to bootstrap the required toolkit for the step (%s), error: %s",
					toolkitName, err)
			}
		}
	}

	//
	buildRunResults := models.BuildRunResultsModel{
		StartTime:      startTime,
		StepmanUpdates: map[string]int{},
		ProjectType:    bitriseConfig.ProjectType,
	}

	buildRunResults, err = activateAndRunWorkflow(
		workflowToRunID, workflowToRun, bitriseConfig,
		buildRunResults,
		&environments, secretEnvironments,
		lastWorkflowID)
	if err != nil {
		return buildRunResults, errors.New("[BITRISE_CLI] - Failed to activate and run workflow " + workflowToRunID)
	}

	// Build finished
	bitrise.PrintSummary(buildRunResults)

	// Trigger WorkflowRunDidFinish
	if err := plugins.TriggerEvent(plugins.DidFinishRun, buildRunResults); err != nil {
		log.Warnf("Failed to trigger WorkflowRunDidFinish, error: %s", err)
	}

	return buildRunResults, nil
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
