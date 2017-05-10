package bitrise

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// InventoryModelFromYAMLBytes ...
func InventoryModelFromYAMLBytes(inventoryBytes []byte) (inventory envmanModels.EnvsSerializeModel, err error) {
	if err = yaml.Unmarshal(inventoryBytes, &inventory); err != nil {
		return
	}

	for _, env := range inventory.Envs {
		if err := env.Normalize(); err != nil {
			return envmanModels.EnvsSerializeModel{}, fmt.Errorf("Failed to normalize bitrise inventory, error: %s", err)
		}
		if err := env.FillMissingDefaults(); err != nil {
			return envmanModels.EnvsSerializeModel{}, fmt.Errorf("Failed to fill bitrise inventory, error: %s", err)
		}
		if err := env.Validate(); err != nil {
			return envmanModels.EnvsSerializeModel{}, fmt.Errorf("Failed to validate bitrise inventory, error: %s", err)
		}
	}

	return
}

func searchEnvInSlice(searchForEnvKey string, searchIn []envmanModels.EnvironmentItemModel) (envmanModels.EnvironmentItemModel, int, error) {
	for idx, env := range searchIn {
		key, _, err := env.GetKeyValuePair()
		if err != nil {
			return envmanModels.EnvironmentItemModel{}, -1, err
		}

		if key == searchForEnvKey {
			return env, idx, nil
		}
	}
	return envmanModels.EnvironmentItemModel{}, -1, nil
}

// ApplyOutputAliases ...
func ApplyOutputAliases(onEnvs, basedOnEnvs []envmanModels.EnvironmentItemModel) ([]envmanModels.EnvironmentItemModel, error) {
	for _, basedOnEnv := range basedOnEnvs {
		envKey, envKeyAlias, err := basedOnEnv.GetKeyValuePair()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}

		envToAlias, idx, err := searchEnvInSlice(envKey, onEnvs)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}

		if idx > -1 && envKeyAlias != "" {
			_, origValue, err := envToAlias.GetKeyValuePair()
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, err
			}

			origOptions, err := envToAlias.GetOptions()
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, err
			}

			onEnvs[idx] = envmanModels.EnvironmentItemModel{
				envKeyAlias:             origValue,
				envmanModels.OptionsKey: origOptions,
			}
		}
	}
	return onEnvs, nil
}

// CollectEnvironmentsFromFile ...
func CollectEnvironmentsFromFile(pth string) ([]envmanModels.EnvironmentItemModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	return CollectEnvironmentsFromFileContent(bytes)
}

// CollectEnvironmentsFromFileContent ...
func CollectEnvironmentsFromFileContent(bytes []byte) ([]envmanModels.EnvironmentItemModel, error) {
	var envstore envmanModels.EnvsSerializeModel
	if err := yaml.Unmarshal(bytes, &envstore); err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	for _, env := range envstore.Envs {
		if err := env.Normalize(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
		if err := env.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
		if err := env.Validate(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
	}

	return envstore.Envs, nil
}

// ExportEnvironmentsList ...
func ExportEnvironmentsList(envsList []envmanModels.EnvironmentItemModel) error {
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

		isExpand := envmanModels.DefaultIsExpand
		if opts.IsExpand != nil {
			isExpand = *opts.IsExpand
		}

		skipIfEmpty := envmanModels.DefaultSkipIfEmpty
		if opts.SkipIfEmpty != nil {
			skipIfEmpty = *opts.SkipIfEmpty
		}

		if err := tools.EnvmanAdd(configs.InputEnvstorePath, key, value, isExpand, skipIfEmpty); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run envman add")
			return err
		}
	}
	return nil
}

// CleanupStepWorkDir ...
func CleanupStepWorkDir() error {
	stepYMLPth := filepath.Join(configs.BitriseWorkDirPath, "current_step.yml")
	if err := command.RemoveFile(stepYMLPth); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step yml: ", err))
	}

	stepDir := configs.BitriseWorkStepsDirPath
	if err := command.RemoveDir(stepDir); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step work dir: ", err))
	}
	return nil
}

// GetBuildFailedEnvironments ...
func GetBuildFailedEnvironments(failed bool) []string {
	statusStr := "0"
	if failed {
		statusStr = "1"
	}

	environments := []string{}
	steplibBuildStatusEnv := "STEPLIB_BUILD_STATUS" + "=" + statusStr
	environments = append(environments, steplibBuildStatusEnv)

	bitriseBuildStatusEnv := "BITRISE_BUILD_STATUS" + "=" + statusStr
	environments = append(environments, bitriseBuildStatusEnv)
	return environments
}

// SetBuildFailedEnv ...
func SetBuildFailedEnv(failed bool) error {
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

// FormattedSecondsToMax8Chars ...
func FormattedSecondsToMax8Chars(t time.Duration) (string, error) {
	sec := t.Seconds()
	min := t.Minutes()
	hour := t.Hours()

	if sec < 1.0 {
		// 0.999999 sec -> 0.99 sec
		return fmt.Sprintf("%.2f sec", sec), nil // 8
	} else if sec < 10.0 {
		// 9.99999 sec -> 9.99 sec
		return fmt.Sprintf("%.2f sec", sec), nil // 8
	} else if sec < 600 {
		// 599,999 sec -> 599 sec
		return fmt.Sprintf("%.f sec", sec), nil // 7
	} else if min < 60 {
		// 59,999 min -> 59.9 min
		return fmt.Sprintf("%.1f min", min), nil // 8
	} else if hour < 10 {
		// 9.999 hour -> 9.9 hour
		return fmt.Sprintf("%.1f hour", hour), nil // 8
	} else if hour < 1000 {
		// 999,999 hour -> 999 hour
		return fmt.Sprintf("%.f hour", hour), nil // 8
	}

	return "", fmt.Errorf("time (%f hour) greater than max allowed (999 hour)", hour)
}

// SaveConfigToFile ...
func SaveConfigToFile(pth string, bitriseConf models.BitriseDataModel) error {
	contBytes, err := generateYAML(bitriseConf)
	if err != nil {
		return err
	}
	if err := fileutil.WriteBytesToFile(pth, contBytes); err != nil {
		return err
	}
	return nil
}

func generateYAML(v interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

func normalizeValidateFillMissingDefaults(bitriseData *models.BitriseDataModel) ([]string, error) {
	if err := bitriseData.Normalize(); err != nil {
		return []string{}, err
	}
	warnings, err := bitriseData.Validate()
	if err != nil {
		return warnings, err
	}
	if err := bitriseData.FillMissingDefaults(); err != nil {
		return warnings, err
	}
	return warnings, nil
}

// ConfigModelFromYAMLBytes ...
func ConfigModelFromYAMLBytes(configBytes []byte) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	if err = yaml.Unmarshal(configBytes, &bitriseData); err != nil {
		return
	}

	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData)
	if err != nil {
		return
	}

	return
}

// ConfigModelFromJSONBytes ...
func ConfigModelFromJSONBytes(configBytes []byte) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	if err = json.Unmarshal(configBytes, &bitriseData); err != nil {
		return
	}
	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData)
	if err != nil {
		return
	}

	return
}

// ReadBitriseConfig ...
func ReadBitriseConfig(pth string) (models.BitriseDataModel, []string, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.BitriseDataModel{}, []string{}, err
	} else if !isExists {
		return models.BitriseDataModel{}, []string{}, fmt.Errorf("No file found at path: %s", pth)
	}

	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return models.BitriseDataModel{}, []string{}, err
	}

	if len(bytes) == 0 {
		return models.BitriseDataModel{}, []string{}, errors.New("empty config")
	}

	if strings.HasSuffix(pth, ".json") {
		log.Debugln("=> Using JSON parser for: ", pth)
		return ConfigModelFromJSONBytes(bytes)
	}

	log.Debugln("=> Using YAML parser for: ", pth)
	return ConfigModelFromYAMLBytes(bytes)
}

// ReadSpecStep ...
func ReadSpecStep(pth string) (stepmanModels.StepModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return stepmanModels.StepModel{}, err
	} else if !isExists {
		return stepmanModels.StepModel{}, fmt.Errorf("No file found at path: %s", pth)
	}

	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return stepmanModels.StepModel{}, err
	}

	var stepModel stepmanModels.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.Normalize(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.ValidateInputAndOutputEnvs(false); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	return stepModel, nil
}

func getInputByKey(inputs []envmanModels.EnvironmentItemModel, key string) (envmanModels.EnvironmentItemModel, error) {
	for _, input := range inputs {
		aKey, _, err := input.GetKeyValuePair()
		if err != nil {
			return envmanModels.EnvironmentItemModel{}, err
		}
		if aKey == key {
			return input, nil
		}
	}
	return envmanModels.EnvironmentItemModel{}, fmt.Errorf("No Environmnet found for key (%s)", key)
}

func isStringSliceWithSameElements(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	m := make(map[string]bool, len(s1))
	for _, s := range s1 {
		m[s] = true
	}

	for _, s := range s2 {
		v, found := m[s]
		if !found || !v {
			return false
		}
		delete(m, s)
	}
	return len(m) == 0
}

func isDependecyEqual(d1, d2 stepmanModels.DependencyModel) bool {
	return (d1.Manager == d2.Manager && d1.Name == d2.Name)
}

func containsDependecy(m map[stepmanModels.DependencyModel]bool, d1 stepmanModels.DependencyModel) bool {
	for d2 := range m {
		if isDependecyEqual(d1, d2) {
			return true
		}
	}
	return false
}

func isDependencySliceWithSameElements(s1, s2 []stepmanModels.DependencyModel) bool {
	if len(s1) != len(s2) {
		return false
	}

	m := make(map[stepmanModels.DependencyModel]bool, len(s1))
	for _, s := range s1 {
		m[s] = true
	}

	for _, d := range s2 {
		if containsDependecy(m, d) == false {
			return false
		}
		delete(m, d)
	}
	return len(m) == 0
}

func removeStepDefaultsAndFillStepOutputs(stepListItem *models.StepListItemModel, defaultStepLibSource string) error {
	// Create stepIDData
	compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(*stepListItem)
	if err != nil {
		return err
	}
	stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
	if err != nil {
		return err
	}

	// Activate step - get step.yml
	tempStepCloneDirPath, err := pathutil.NormalizedOSTempDirPath("step_clone")
	if err != nil {
		return err
	}
	tempStepYMLDirPath, err := pathutil.NormalizedOSTempDirPath("step_yml")
	if err != nil {
		return err
	}
	tempStepYMLFilePath := filepath.Join(tempStepYMLDirPath, "step.yml")

	if stepIDData.SteplibSource == "path" {
		stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
		if err != nil {
			return err
		}
		if err := command.CopyFile(filepath.Join(stepAbsLocalPth, "step.yml"), tempStepYMLFilePath); err != nil {
			return err
		}
	} else if stepIDData.SteplibSource == "git" {
		if err := git.CloneTagOrBranch(stepIDData.IDorURI, tempStepCloneDirPath, stepIDData.Version); err != nil {
			return err
		}
		if err := command.CopyFile(filepath.Join(tempStepCloneDirPath, "step.yml"), tempStepYMLFilePath); err != nil {
			return err
		}
	} else if stepIDData.SteplibSource == "_" {
		// Steplib independent steps are completly defined in workflow
		tempStepYMLFilePath = ""
	} else if stepIDData.SteplibSource != "" {
		if err := tools.StepmanSetup(stepIDData.SteplibSource); err != nil {
			return err
		}
		if err := tools.StepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, tempStepCloneDirPath, tempStepYMLFilePath); err != nil {
			return err
		}
	} else {
		return errors.New("Failed to fill step ouputs: unkown SteplibSource")
	}

	// Fill outputs
	if tempStepYMLFilePath != "" {
		specStep, err := ReadSpecStep(tempStepYMLFilePath)
		if err != nil {
			return err
		}

		if workflowStep.Title != nil && specStep.Title != nil && *workflowStep.Title == *specStep.Title {
			workflowStep.Title = nil
		}
		if workflowStep.Description != nil && specStep.Description != nil && *workflowStep.Description == *specStep.Description {
			workflowStep.Description = nil
		}
		if workflowStep.Summary != nil && specStep.Summary != nil && *workflowStep.Summary == *specStep.Summary {
			workflowStep.Summary = nil
		}
		if workflowStep.Website != nil && specStep.Website != nil && *workflowStep.Website == *specStep.Website {
			workflowStep.Website = nil
		}
		if workflowStep.SourceCodeURL != nil && specStep.SourceCodeURL != nil && *workflowStep.SourceCodeURL == *specStep.SourceCodeURL {
			workflowStep.SourceCodeURL = nil
		}
		if workflowStep.SupportURL != nil && specStep.SupportURL != nil && *workflowStep.SupportURL == *specStep.SupportURL {
			workflowStep.SupportURL = nil
		}
		workflowStep.PublishedAt = nil
		if workflowStep.Source != nil && specStep.Source != nil {
			if workflowStep.Source.Git == specStep.Source.Git {
				workflowStep.Source.Git = ""
			}
			if workflowStep.Source.Commit == specStep.Source.Commit {
				workflowStep.Source.Commit = ""
			}
		}
		if isStringSliceWithSameElements(workflowStep.HostOsTags, specStep.HostOsTags) {
			workflowStep.HostOsTags = []string{}
		}
		if isStringSliceWithSameElements(workflowStep.ProjectTypeTags, specStep.ProjectTypeTags) {
			workflowStep.ProjectTypeTags = []string{}
		}
		if isStringSliceWithSameElements(workflowStep.TypeTags, specStep.TypeTags) {
			workflowStep.TypeTags = []string{}
		}
		if isDependencySliceWithSameElements(workflowStep.Dependencies, specStep.Dependencies) {
			workflowStep.Dependencies = []stepmanModels.DependencyModel{}
		}
		if workflowStep.IsRequiresAdminUser != nil && specStep.IsRequiresAdminUser != nil && *workflowStep.IsRequiresAdminUser == *specStep.IsRequiresAdminUser {
			workflowStep.IsRequiresAdminUser = nil
		}
		if workflowStep.IsAlwaysRun != nil && specStep.IsAlwaysRun != nil && *workflowStep.IsAlwaysRun == *specStep.IsAlwaysRun {
			workflowStep.IsAlwaysRun = nil
		}
		if workflowStep.IsSkippable != nil && specStep.IsSkippable != nil && *workflowStep.IsSkippable == *specStep.IsSkippable {
			workflowStep.IsSkippable = nil
		}
		if workflowStep.RunIf != nil && specStep.RunIf != nil && *workflowStep.RunIf == *specStep.RunIf {
			workflowStep.RunIf = nil
		}

		inputs := []envmanModels.EnvironmentItemModel{}
		for _, input := range workflowStep.Inputs {
			sameValue := false

			wfKey, wfValue, err := input.GetKeyValuePair()
			if err != nil {
				return err
			}

			wfOptions, err := input.GetOptions()
			if err != nil {
				return err
			}

			sInput, err := getInputByKey(specStep.Inputs, wfKey)
			if err != nil {
				return err
			}

			_, sValue, err := sInput.GetKeyValuePair()
			if err != nil {
				return err
			}

			if wfValue == sValue {
				sameValue = true
			}

			sOptions, err := sInput.GetOptions()
			if err != nil {
				return err
			}

			hasOptions := false

			if wfOptions.Title != nil && sOptions.Title != nil && *wfOptions.Title == *sOptions.Title {
				wfOptions.Title = nil
			} else {
				hasOptions = true
			}

			if wfOptions.Description != nil && sOptions.Description != nil && *wfOptions.Description == *sOptions.Description {
				wfOptions.Description = nil
			} else {
				hasOptions = true
			}

			if wfOptions.Summary != nil && sOptions.Summary != nil && *wfOptions.Summary == *sOptions.Summary {
				wfOptions.Summary = nil
			} else {
				hasOptions = true
			}

			if isStringSliceWithSameElements(wfOptions.ValueOptions, sOptions.ValueOptions) {
				wfOptions.ValueOptions = []string{}
			} else {
				hasOptions = true
			}

			if wfOptions.IsRequired != nil && sOptions.IsRequired != nil && *wfOptions.IsRequired == *sOptions.IsRequired {
				wfOptions.IsRequired = nil
			} else {
				hasOptions = true
			}

			if wfOptions.IsExpand != nil && sOptions.IsExpand != nil && *wfOptions.IsExpand == *sOptions.IsExpand {
				wfOptions.IsExpand = nil
			} else {
				hasOptions = true
			}

			if wfOptions.IsDontChangeValue != nil && sOptions.IsDontChangeValue != nil && *wfOptions.IsDontChangeValue == *sOptions.IsDontChangeValue {
				wfOptions.IsDontChangeValue = nil
			} else {
				hasOptions = true
			}

			if !hasOptions && sameValue {
				// default env
			} else {
				if hasOptions {
					input[envmanModels.OptionsKey] = wfOptions
				} else {
					delete(input, envmanModels.OptionsKey)
				}

				inputs = append(inputs, input)
			}
		}

		workflowStep.Inputs = inputs

		// We need only key-value and title from spec outputs
		outputs := []envmanModels.EnvironmentItemModel{}
		for _, output := range specStep.Outputs {
			sKey, sValue, err := output.GetKeyValuePair()
			if err != nil {
				return err
			}

			sOptions, err := output.GetOptions()
			if err != nil {
				return err
			}

			newOutput := envmanModels.EnvironmentItemModel{
				sKey: sValue,
				envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
					Title: sOptions.Title,
				},
			}

			outputs = append(outputs, newOutput)
		}

		workflowStep.Outputs = outputs

		(*stepListItem)[compositeStepIDStr] = workflowStep
	}

	// Cleanup
	if err := command.RemoveDir(tempStepCloneDirPath); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step clone dir: ", err))
	}
	if err := command.RemoveDir(tempStepYMLDirPath); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step clone dir: ", err))
	}

	return nil
}

// RemoveConfigRedundantFieldsAndFillStepOutputs ...
func RemoveConfigRedundantFieldsAndFillStepOutputs(config *models.BitriseDataModel) error {
	for _, workflow := range config.Workflows {
		for _, stepListItem := range workflow.Steps {
			if err := removeStepDefaultsAndFillStepOutputs(&stepListItem, config.DefaultStepLibSource); err != nil {
				return err
			}
		}
	}
	if err := config.RemoveRedundantFields(); err != nil {
		return err
	}

	return nil
}
