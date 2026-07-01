package cmdutil

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configmerge"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	ver "github.com/hashicorp/go-version"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"
	// OutputFormatKey ...
	OutputFormatKey = "output-format"
)

// CreateDefaultMerger ...
func CreateDefaultMerger() (*configmerge.Merger, error) {
	opts := log.GetGlobalLoggerOpts()
	logger := log.NewLogger(opts)
	configReader, err := configmerge.NewConfigReader(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create config module reader: %w", err)
	}
	merger := configmerge.NewMerger(configReader, logger)
	return &merger, nil
}

// GetBitriseConfigFromBase64Data ...
func GetBitriseConfigFromBase64Data(configBase64Str string, validation bitrise.ValidationType) (models.BitriseDataModel, []string, error) {
	configBase64Bytes, err := base64.StdEncoding.DecodeString(configBase64Str)
	if err != nil {
		return models.BitriseDataModel{}, []string{}, fmt.Errorf("failed to decode base 64 string, error: %s", err)
	}

	config, warnings, err := bitrise.ConfigModelFromYAMLBytesWithValidation(configBase64Bytes, validation)
	if err != nil {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("failed to parse bitrise config, error: %s", err)
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
func CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath string, validation bitrise.ValidationType) (models.BitriseDataModel, []string, error) {
	var bitriseConfig *models.BitriseDataModel
	warnings := []string{}

	if bitriseConfigBase64Data != "" {
		config, warns, err := GetBitriseConfigFromBase64Data(bitriseConfigBase64Data, validation)
		warnings = warns
		if err != nil {
			return models.BitriseDataModel{}, warnings, fmt.Errorf("failed to get Bitrise config (bitrise.yml) from base 64 data: %w", err)
		}
		bitriseConfig = &config
	} else {
		bitriseConfigPath, err := GetBitriseConfigFilePath(bitriseConfigPath)
		if err != nil {
			return models.BitriseDataModel{}, []string{}, fmt.Errorf("failed to get Bitrise config (bitrise.yml) path: %w", err)
		}
		if bitriseConfigPath == "" {
			return models.BitriseDataModel{}, []string{}, errors.New("empty Bitrise config (bitrise.yml) path")
		}

		isModularConfig, err := configmerge.IsModularConfig(bitriseConfigPath)
		if err != nil {
			return models.BitriseDataModel{}, []string{}, fmt.Errorf("failed to check if the config is modular: %s", err)
		}

		if isModularConfig {
			merger, err := CreateDefaultMerger()
			if err != nil {
				return models.BitriseDataModel{}, warnings, fmt.Errorf("failed to create config module merger: %w", err)
			}
			mergedConfigContent, _, err := merger.MergeConfig(bitriseConfigPath)
			if err != nil {
				return models.BitriseDataModel{}, []string{}, fmt.Errorf("failed to merge Bitrise config (%s): %w", bitriseConfigPath, err)
			}

			isJSON := filepath.Ext(bitriseConfigPath) == ".json"
			config, warns, err := bitrise.ConfigModelFromFileContent([]byte(mergedConfigContent), isJSON, validation)
			warnings = warns
			if err != nil {
				return models.BitriseDataModel{}, warnings, fmt.Errorf("config (%s) is not valid: %w", bitriseConfigPath, err)
			}
			bitriseConfig = &config
		} else {
			config, warns, err := bitrise.ReadBitriseConfig(bitriseConfigPath, validation)
			warnings = warns
			if err != nil {
				return models.BitriseDataModel{}, warnings, fmt.Errorf("config (%s) is not valid: %w", bitriseConfigPath, err)
			}
			bitriseConfig = &config
		}
	}

	supportedVersion, err := ver.NewVersion(models.FormatVersion)
	if err != nil {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("failed to parse bitrise CLI supported format version (%s): %s", models.FormatVersion, err)
	}

	configVersion, err := ver.NewVersion(bitriseConfig.FormatVersion)
	if err != nil {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("failed to parse bitrise.yml format version (%s): %s", bitriseConfig.FormatVersion, err)
	}

	if configVersion.GreaterThan(supportedVersion) {
		return models.BitriseDataModel{}, warnings, fmt.Errorf("the bitrise.yml has a higher format version (%s) than the bitrise CLI supported format version (%s), please upgrade your bitrise CLI to use this bitrise.yml", bitriseConfig.FormatVersion, models.FormatVersion)
	}

	return *bitriseConfig, warnings, nil
}

// GetInventoryFromBase64Data ...
func GetInventoryFromBase64Data(inventoryBase64Str string) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryBase64Bytes, err := base64.StdEncoding.DecodeString(inventoryBase64Str)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to decode base 64 string, error: %s", err)
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

// CreateInventoryFromCLIParams ...
func CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath string) ([]envmanModels.EnvironmentItemModel, error) {
	inventoryEnvironments := []envmanModels.EnvironmentItemModel{}

	if inventoryBase64Data != "" {
		inventory, err := GetInventoryFromBase64Data(inventoryBase64Data)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to get inventory from base 64 data, err: %s", err)
		}
		inventoryEnvironments = inventory
	} else {
		inventoryPath, err := GetInventoryFilePath(inventoryPath)
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to get inventory path: %s", err)
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
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("invalid inventory format: %s", err)
			}
			inventoryEnvironments = inventory
		}
	}

	return inventoryEnvironments, nil
}
