package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/log"
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
	// IsSteplibOfflineModeEnvKey when set to true:
	// - StepLib update will be disabled when using non-exact step version (latest minor or major).
	// - When a step or step version is not found in the cache, will not be downloaded. Instead will log
	//  a error message (including what other Step versions are available).
	// - Analytics will be disabled.
	IsSteplibOfflineModeEnvKey = "BITRISE_OFFLINE_MODE"
	// SetupNoUpdateEnvKey when set to "true", skips updating core tools (stepman/envman) and plugins
	// during setup if they are already installed, even if their version is below the minimum required.
	// Tools that are missing entirely are still installed. Intended for CI environments where GitHub
	// fetches during setup are prone to rate-limiting.
	SetupNoUpdateEnvKey = "BITRISE_SETUP_NO_UPDATE"

	// --- Debug Options

	// DebugUseSystemTools ...
	DebugUseSystemTools = "BITRISE_DEBUG_USE_SYSTEM_TOOLS"
)

const (
	bitriseConfigFileName = "config.json"
)

const (
	selfUpdateInterval   = 24 * time.Hour
	PluginUpdateInterval = 24 * time.Hour
)

// IsDebugUseSystemTools ...
func IsDebugUseSystemTools() bool {
	return os.Getenv(DebugUseSystemTools) == "true"
}

// LoadLegacyConfig exposes the on-disk ~/.bitrise/config.json contents. The
// returned bool reports whether the file exists — needed because a missing
// file and one that decodes to an all-zero-value ConfigModel are otherwise
// indistinguishable, and Save* below must not create the file for a
// brand-new user (see saveConfig).
func LoadLegacyConfig() (ConfigModel, bool, error) {
	if err := EnsureBitriseConfigDirExists(); err != nil {
		return ConfigModel{}, false, err
	}

	configPth := getLegacyConfigFilePath()
	exist, err := pathutil.IsPathExists(configPth)
	if err != nil {
		return ConfigModel{}, false, err
	}
	if !exist {
		return ConfigModel{}, false, nil
	}

	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return ConfigModel{}, true, err
	}

	if len(bytes) == 0 {
		return ConfigModel{}, true, errors.New("empty config file")
	}

	config := ConfigModel{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return ConfigModel{}, true, fmt.Errorf("failed to unmarshal config (%s), error: %s", string(bytes), err)
	}

	return config, true, nil
}

func saveLegacyConfig(config ConfigModel) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configPth := getLegacyConfigFilePath()
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
	config, err := ResolveConfig()
	if err != nil {
		return false
	}

	duration := time.Since(config.LastCLIUpdateCheck)
	return duration >= selfUpdateInterval
}

func SaveCLIUpdateCheck() error {
	config, existed, err := LoadLegacyConfig()
	if err != nil {
		return err
	}
	config.LastCLIUpdateCheck = time.Now()

	return saveConfig(existed, config, func(c *internalconfig.Config) {
		c.LastCLIUpdateCheck = config.LastCLIUpdateCheck
	})
}

func CheckIsPluginUpdateCheckRequired(plugin string) bool {
	config, err := ResolveConfig()
	if err != nil {
		return false
	}

	duration := time.Since(config.LastPluginUpdateChecks[plugin])
	return duration >= PluginUpdateInterval
}

func SavePluginUpdateCheck(plugin string) error {
	config, existed, err := LoadLegacyConfig()
	if err != nil {
		return err
	}
	if config.LastPluginUpdateChecks == nil {
		config.LastPluginUpdateChecks = map[string]time.Time{}
	}
	config.LastPluginUpdateChecks[plugin] = time.Now()

	return saveConfig(existed, config, func(c *internalconfig.Config) {
		if c.LastPluginUpdateChecks == nil {
			c.LastPluginUpdateChecks = map[string]time.Time{}
		}
		c.LastPluginUpdateChecks[plugin] = config.LastPluginUpdateChecks[plugin]
	})
}

func CheckIsSetupWasDoneForVersion(ver string) (bool, string) {
	config, err := ResolveConfig()
	if err != nil {
		return false, ""
	}
	return config.SetupVersion == ver, config.SetupVersion
}

func SaveSetupSuccessForVersion(ver string) error {
	config, existed, err := LoadLegacyConfig()
	if err != nil {
		return err
	}
	config.SetupVersion = ver

	return saveConfig(existed, config, func(c *internalconfig.Config) {
		c.SetupVersion = config.SetupVersion
	})
}

func (m ConfigModel) ToConfig() internalconfig.Config {
	return internalconfig.Config{
		SetupVersion:           m.SetupVersion,
		LastCLIUpdateCheck:     m.LastCLIUpdateCheck,
		LastPluginUpdateChecks: m.LastPluginUpdateChecks,
	}
}

// ResolveConfig merges the legacy config with the current per-dir and global
// layers, via internal/config.Resolve — the same precedence used everywhere
// else, so callers can't drift from it. A load failure in the per-dir or
// global layer is logged and treated as that layer being absent; a legacy
// load failure is returned, since Check* above treat it as fatal (fail
// closed).
func ResolveConfig() (internalconfig.Resolved, error) {
	legacy, _, err := LoadLegacyConfig()
	if err != nil {
		return internalconfig.Resolved{}, err
	}
	dirCfg, _, err := internalconfig.LoadDir()
	if err != nil {
		log.Warnf("Failed to load .bitrise-cli.yml, ignoring: %s", err)
		dirCfg = internalconfig.Config{}
	}
	globalCfg, err := internalconfig.Load()
	if err != nil {
		log.Warnf("Failed to load config.yml, ignoring: %s", err)
		globalCfg = internalconfig.Config{}
	}
	return internalconfig.Resolve(legacy.ToConfig(), dirCfg, globalCfg), nil
}

// saveGlobalConfig loads the current global config.yml
// (~/.config/bitrise/cli/config.yml), applies mutate, and saves it back,
// returning any load/save error.
func saveGlobalConfig(mutate func(*internalconfig.Config)) error {
	globalCfg, err := internalconfig.Load()
	if err != nil {
		return err
	}
	mutate(&globalCfg)
	return internalconfig.Save(globalCfg)
}

// saveConfig writes legacy via saveLegacyConfig only when it existed —
// a brand-new user should never get a legacy file created — and always syncs
// mutate into the new global config.yml.
//
// legacy is deliberately the raw on-disk legacy value (each caller loads it
// via LoadLegacyConfig, not ResolveConfig): both writes below only ever touch
// the one field being changed, so a value that only exists in the per-dir or
// global layer never gets copied into a file that should stay a
// self-contained snapshot of what was actually written to it.
func saveConfig(existed bool, legacy ConfigModel, mutate func(*internalconfig.Config)) error {
	if existed {
		if err := saveLegacyConfig(legacy); err != nil {
			return err
		}
		if err := saveGlobalConfig(mutate); err != nil {
			log.Warnf("Failed to sync config.yml, ignoring: %s", err)
		}
		return nil
	}

	if err := saveGlobalConfig(mutate); err != nil {
		return fmt.Errorf("failed to save config.yml: %w", err)
	}
	return nil
}
