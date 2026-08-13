package configs

import (
	"context"
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

	return saveConfig(existed, config, func(c *internalconfig.Config) error {
		c.LastCLIUpdateCheck = config.LastCLIUpdateCheck
		return nil
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

	return saveConfig(existed, config, func(c *internalconfig.Config) error {
		if c.LastPluginUpdateChecks == nil {
			c.LastPluginUpdateChecks = map[string]time.Time{}
		}
		c.LastPluginUpdateChecks[plugin] = config.LastPluginUpdateChecks[plugin]
		return nil
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

	return saveConfig(existed, config, func(c *internalconfig.Config) error {
		c.SetupVersion = config.SetupVersion
		return nil
	})
}

// SetAppID persists appID as the default app for cloud commands (e.g. `bitrise
// app create`), so later commands can resolve it without --app/BITRISE_APP_ID.
// Unlike SetupVersion/LastCLIUpdateCheck/LastPluginUpdateChecks, app_id has no
// ~/.bitrise/config.json counterpart, so it never touches the legacy file.
//
// Routed through Config.Set rather than assigning c.AppID directly, so this
// writer and `bitrise config set app_id` can't drift if validation is ever
// added to that key.
func SetAppID(appID string) error {
	return saveGlobalConfig(func(c *internalconfig.Config) error {
		return c.Set(internalconfig.KeyAppID, appID)
	})
}

// SetAppID persists appID as the default app for cloud commands (e.g. `bitrise
// app create`), so later commands can resolve it without --app/BITRISE_APP_ID.
// Unlike SetupVersion/LastCLIUpdateCheck/LastPluginUpdateChecks, app_id has no
// ~/.bitrise/config.json counterpart, so it never touches the legacy file.
//
// Routed through Config.Set rather than assigning c.AppID directly, so this
// writer and `bitrise config set app_id` can't drift if validation is ever
// added to that key.
func SetAppID(appID string) error {
	return saveGlobalConfig(func(c *internalconfig.Config) error {
		return c.Set(internalconfig.KeyAppID, appID)
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
// else, so callers can't drift from it. A load failure in any layer (legacy,
// per-dir, or global) is logged and treated as that layer being absent, so
// the other layers and the eventual default (e.g. APIBaseURL) still resolve
// normally. A legacy load failure is still returned as an error — Check*
// above treat it as fatal and ignore the resolved value — but the resolved
// value itself is no longer discarded, so other callers (e.g. the CLI's root
// context) don't end up with an all-zero config just because the legacy file
// failed to load.
func ResolveConfig() (internalconfig.Resolved, error) {
	legacy, _, legacyErr := LoadLegacyConfig()
	if legacyErr != nil {
		legacy = ConfigModel{}
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
	return internalconfig.Resolve(legacy.ToConfig(), dirCfg, globalCfg), legacyErr
}

// saveGlobalConfig loads the current global config.yml
// (~/.config/bitrise/cli/config.yml), applies mutate, and saves it back,
// returning any load/save error. A failing mutate aborts before the save,
// leaving the file untouched.
//
// The load-mutate-save sequence holds the same cross-process lock
// `bitrise config set` takes, so a concurrent writer (another CLI invocation
// setting a different key) can't have its write dropped by this one. None of
// this function's callers have a cobra command to draw a request-scoped
// context from, so the wait is bounded by internalconfig.LockStaleAfter
// instead of blocking on context.Background() indefinitely.
func saveGlobalConfig(mutate func(*internalconfig.Config) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), internalconfig.LockStaleAfter)
	defer cancel()
	unlock, err := internalconfig.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	globalCfg, err := internalconfig.Load()
	if err != nil {
		return err
	}
	if err := mutate(&globalCfg); err != nil {
		return err
	}
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
//
// Only the config.yml half (saveGlobalConfig) is protected by
// internalconfig.Lock. saveLegacyConfig's read-modify-write of the legacy
// ~/.bitrise/config.json above is not — a second concurrent call here can
// still drop this call's update to that file.
func saveConfig(existed bool, legacy ConfigModel, mutate func(*internalconfig.Config) error) error {
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
