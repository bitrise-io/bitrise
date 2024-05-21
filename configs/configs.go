package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// ConfigModel ...
type ConfigModel struct {
	SetupVersion           string               `json:"setup_version"`
	LastCLIUpdateCheck     time.Time            `json:"last_cli_update_check"`
	LastPluginUpdateChecks map[string]time.Time `json:"last_plugin_update_checks"`
}

// ---------------------------
// --- Project level vars / configs

var (
	// IsCIMode ...
	IsCIMode = false
	// IsDebugMode ...
	IsDebugMode = false
	// IsPullRequestMode ...
	IsPullRequestMode = false

	// IsSecretFiltering ...
	IsSecretFiltering = false
	// IsSecretEnvsFiltering ...
	IsSecretEnvsFiltering = false

	// IsSteplibOfflineMode should not be used, only for access from setup command
	IsSteplibOfflineMode = false
)

// ---------------------------
// --- Consts

const (
	// CIModeEnvKey ...
	CIModeEnvKey = "CI"
	// PRModeEnvKey ...
	PRModeEnvKey = "PR"
	// PullRequestIDEnvKey ...
	PullRequestIDEnvKey = "PULL_REQUEST_ID"
	// DebugModeEnvKey ...
	DebugModeEnvKey = "DEBUG"
	// IsSecretFilteringKey ...
	IsSecretFilteringKey = "BITRISE_SECRET_FILTERING"
	// IsSecretEnvsFilteringKey ...
	IsSecretEnvsFilteringKey = "BITRISE_SECRET_ENVS_FILTERING"
	// NoOutputTimeoutEnvKey ...
	NoOutputTimeoutEnvKey = "BITRISE_NO_OUTPUT_TIMEOUT"
	// when true:
	// - StepLib update will be disabled when using non-exact step version (latest minor or major)
	// - When a step or step version is not found in the cache, will not be downloaded. Instead will log
	//  a error message (including what other Step versions are available).
	// - Analytics will be disabled
	IsSteplibOfflineModeEnvKey = "BITRISE_OFFLINE_MODE"

	// --- Debug Options

	// DebugUseSystemTools ...
	DebugUseSystemTools = "BITRISE_DEBUG_USE_SYSTEM_TOOLS"
)

const (
	bitriseConfigFileName = "config.json"
)

// IsDebugUseSystemTools ...
func IsDebugUseSystemTools() bool {
	return os.Getenv(DebugUseSystemTools) == "true"
}

func loadBitriseConfig() (ConfigModel, error) {
	if err := EnsureBitriseConfigDirExists(); err != nil {
		return ConfigModel{}, err
	}

	configPth := getBitriseConfigFilePath()
	if exist, err := pathutil.IsPathExists(configPth); err != nil {
		return ConfigModel{}, err
	} else if !exist {
		return ConfigModel{}, nil
	}

	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return ConfigModel{}, err
	}

	if len(bytes) == 0 {
		return ConfigModel{}, errors.New("empty config file")
	}

	config := ConfigModel{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return ConfigModel{}, fmt.Errorf("failed to marshal config (%s), error: %s", string(bytes), err)
	}

	return config, nil
}

func saveBitriseConfig(config ConfigModel) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configPth := getBitriseConfigFilePath()
	return fileutil.WriteBytesToFile(configPth, bytes)
}

func DeleteBitriseConfigDir() error {
	confDirPth := GetBitriseHomeDirPath()
	return os.RemoveAll(confDirPth)
}

func EnsureBitriseConfigDirExists() error {
	confDirPth := GetBitriseHomeDirPath()
	return pathutil.EnsureDirExist(confDirPth)
}

func CheckIsCLIUpdateCheckRequired() bool {
	config, err := loadBitriseConfig()
	if err != nil {
		return false
	}

	duration := time.Now().Sub(config.LastCLIUpdateCheck)
	if duration.Hours() >= 24 {
		return true
	}

	return false
}

func SaveCLIUpdateCheck() error {
	config, err := loadBitriseConfig()
	if err != nil {
		return err
	}

	config.LastCLIUpdateCheck = time.Now()

	return saveBitriseConfig(config)
}

func CheckIsPluginUpdateCheckRequired(plugin string) bool {
	config, err := loadBitriseConfig()
	if err != nil {
		return false
	}

	duration := time.Now().Sub(config.LastPluginUpdateChecks[plugin])
	if duration.Hours() >= 24 {
		return true
	}

	return false
}

func SavePluginUpdateCheck(plugin string) error {
	config, err := loadBitriseConfig()
	if err != nil {
		return err
	}

	if config.LastPluginUpdateChecks == nil {
		config.LastPluginUpdateChecks = map[string]time.Time{}
	}

	config.LastPluginUpdateChecks[plugin] = time.Now()

	return saveBitriseConfig(config)
}

func CheckIsSetupWasDoneForVersion(ver string) bool {
	config, err := loadBitriseConfig()
	if err != nil {
		return false
	}
	return (config.SetupVersion == ver)
}

func SaveSetupSuccessForVersion(ver string) error {
	config, err := loadBitriseConfig()
	if err != nil {
		return err
	}

	config.SetupVersion = ver

	return saveBitriseConfig(config)
}
