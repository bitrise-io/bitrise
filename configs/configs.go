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

	// NoOutputTimeout is the timeout after Steps are aborted, when no output is received
	NoOutputTimeout time.Duration
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
	// LogLevelEnvKey ...
	LogLevelEnvKey = "LOGLEVEL"
	// IsSecretFilteringKey ...
	IsSecretFilteringKey = "BITRISE_SECRET_FILTERING"
	// IsSecretEnvsFilteringKey ...
	IsSecretEnvsFilteringKey = "BITRISE_SECRET_ENVS_FILTERING"
	// NoOutputTimeoutEnvKey ...
	NoOutputTimeoutEnvKey = "BITRISE_NO_OUTPUT_TIMEOUT"

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

// DeleteBitriseConfigDir ...
func DeleteBitriseConfigDir() error {
	confDirPth := GetBitriseHomeDirPath()
	return os.RemoveAll(confDirPth)
}

// EnsureBitriseConfigDirExists ...
func EnsureBitriseConfigDirExists() error {
	confDirPth := GetBitriseHomeDirPath()
	return pathutil.EnsureDirExist(confDirPth)
}

// CheckIsCLIUpdateCheckRequired ...
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

// SaveCLIUpdateCheck ...
func SaveCLIUpdateCheck() error {
	config, err := loadBitriseConfig()
	if err != nil {
		return err
	}

	config.LastCLIUpdateCheck = time.Now()

	return saveBitriseConfig(config)
}

// CheckIsPluginUpdateCheckRequired ...
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

// SavePluginUpdateCheck ...
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

// CheckIsSetupWasDoneForVersion ...
func CheckIsSetupWasDoneForVersion(ver string) bool {
	config, err := loadBitriseConfig()
	if err != nil {
		return false
	}
	return (config.SetupVersion == ver)
}

// SaveSetupSuccessForVersion ...
func SaveSetupSuccessForVersion(ver string) error {
	config, err := loadBitriseConfig()
	if err != nil {
		return err
	}

	config.SetupVersion = ver

	return saveBitriseConfig(config)
}
