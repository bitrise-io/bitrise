package bitrise

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

type ValidationType int

const (
	ValidationTypeFull ValidationType = iota
	ValidationTypeMinimal
)

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

func ApplySensitiveOutputs(onEnvs, basedOnEnvs []envmanModels.EnvironmentItemModel) ([]envmanModels.EnvironmentItemModel, error) {
	for _, basedOnEnv := range basedOnEnvs {
		envKey, _, err := basedOnEnv.GetKeyValuePair()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}

		opts, err := basedOnEnv.GetOptions()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}

		if opts.IsSensitive == nil || !*opts.IsSensitive {
			continue
		}

		envToAlias, idx, err := searchEnvInSlice(envKey, onEnvs)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}

		if idx > -1 {
			origKey, origValue, err := envToAlias.GetKeyValuePair()
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, err
			}

			origOptions, err := envToAlias.GetOptions()
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, err
			}

			if origKey == envKey {
				origOptions.IsSensitive = pointers.NewBoolPtr(true)

				onEnvs[idx] = envmanModels.EnvironmentItemModel{
					origKey:                 origValue,
					envmanModels.OptionsKey: origOptions,
				}
			}
		}
	}
	return onEnvs, nil
}

func CollectEnvironmentsFromFile(pth string) ([]envmanModels.EnvironmentItemModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	return CollectEnvironmentsFromFileContent(bytes)
}

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

func BuildStatusEnvs(isBuildFailed bool) []envmanModels.EnvironmentItemModel {
	statusStr := "0"
	if isBuildFailed {
		statusStr = "1"
	}

	return []envmanModels.EnvironmentItemModel{
		{"STEPLIB_BUILD_STATUS": statusStr},
		{"BITRISE_BUILD_STATUS": statusStr},
	}
}

func FailingStepEnvs(failingStepRunResult models.StepRunResultsModel) []envmanModels.EnvironmentItemModel {
	failingStep := failingStepRunResult.StepInfo.ID
	if failingStepRunResult.StepInfo.Step.Title != nil && len(*failingStepRunResult.StepInfo.Step.Title) > 0 {
		failingStep = *failingStepRunResult.StepInfo.Step.Title
	}

	failureReason := failingStepRunResult.ErrorStr

	return []envmanModels.EnvironmentItemModel{
		{"BITRISE_FAILED_STEP_TITLE": failingStep},
		{"BITRISE_FAILED_STEP_FAILURE_REASON": failureReason},
	}
}

func normalizeValidateFillMissingDefaults(bitriseData *models.BitriseDataModel, validation ValidationType) ([]string, error) {
	if err := bitriseData.Normalize(); err != nil {
		return []string{}, err
	}

	var validationFunc func() ([]string, error)
	if validation == ValidationTypeFull {
		validationFunc = bitriseData.Validate
	} else {
		validationFunc = bitriseData.MinimalValidation
	}

	warnings, err := validationFunc()
	if err != nil {
		return warnings, err
	}

	if err := bitriseData.FillMissingDefaults(); err != nil {
		return warnings, err
	}

	return warnings, nil
}

func ConfigModelFromFileContent(configBytes []byte, isJSON bool, validation ValidationType) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	if isJSON {
		bitriseData, err = jsonBytesToConfig(configBytes)
	} else {
		bitriseData, err = yamlBytesToConfig(configBytes)
	}

	if err != nil {
		return
	}

	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData, validation)
	return
}

func ConfigModelFromYAMLBytes(configBytes []byte) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	bitriseData, err = yamlBytesToConfig(configBytes)
	if err != nil {
		return
	}

	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData, ValidationTypeFull)
	return
}

func ConfigModelFromYAMLBytesWithValidation(configBytes []byte, validation ValidationType) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	bitriseData, err = yamlBytesToConfig(configBytes)
	if err != nil {
		return
	}

	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData, validation)
	return
}

func ConfigModelFromJSONBytes(configBytes []byte) (bitriseData models.BitriseDataModel, warnings []string, err error) {
	bitriseData, err = jsonBytesToConfig(configBytes)
	if err != nil {
		return
	}

	warnings, err = normalizeValidateFillMissingDefaults(&bitriseData, ValidationTypeFull)
	return
}

func jsonBytesToConfig(bytes []byte) (models.BitriseDataModel, error) {
	var config models.BitriseDataModel
	if err := json.Unmarshal(bytes, &config); err != nil {
		return models.BitriseDataModel{}, err
	}

	return config, nil
}

func yamlBytesToConfig(bytes []byte) (models.BitriseDataModel, error) {
	var config models.BitriseDataModel
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return models.BitriseDataModel{}, err
	}

	return config, nil
}

func ReadBitriseConfig(pth string, validation ValidationType) (models.BitriseDataModel, []string, error) {
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

	return ConfigModelFromFileContent(bytes, strings.HasSuffix(pth, ".json"), validation)
}

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
