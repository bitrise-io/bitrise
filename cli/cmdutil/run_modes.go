package cmdutil

import (
	"os"
	"strconv"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/tools"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

// IsPRMode ...
func IsPRMode(prGlobalFlagPtr *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
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

// RegisterPrMode ...
func RegisterPrMode(isPRMode bool) error {
	configs.IsPullRequestMode = isPRMode
	return os.Setenv(configs.PRModeEnvKey, strconv.FormatBool(isPRMode))
}

// IsCIMode ...
func IsCIMode(ciGlobalFlagPtr *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
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

// RegisterCIMode ...
func RegisterCIMode(isCIMode bool) error {
	configs.IsCIMode = isCIMode
	return os.Setenv(configs.CIModeEnvKey, strconv.FormatBool(isCIMode))
}

// IsSecretFiltering ...
func IsSecretFiltering(filteringFlag *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
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

// RegisterSecretFiltering ...
func RegisterSecretFiltering(filtering bool) error {
	configs.IsSecretFiltering = filtering
	return os.Setenv(configs.IsSecretFilteringKey, strconv.FormatBool(filtering))
}

// IsSecretEnvsFiltering ...
func IsSecretEnvsFiltering(filteringFlag *bool, inventoryEnvironments []envmanModels.EnvironmentItemModel) (bool, error) {
	if filteringFlag != nil {
		return *filteringFlag, nil
	}

	expandedEnvs, err := tools.ExpandEnvItems(inventoryEnvironments, os.Environ())
	if err != nil {
		return false, err
	}

	value, ok := expandedEnvs[configs.IsSecretEnvsFilteringKey]
	if ok {
		if value == "true" {
			return true, nil
		} else if value == "false" {
			return false, nil
		}
	}

	return true, nil
}

// RegisterSecretEnvsFiltering ...
func RegisterSecretEnvsFiltering(filtering bool) error {
	configs.IsSecretEnvsFiltering = filtering
	return os.Setenv(configs.IsSecretEnvsFilteringKey, strconv.FormatBool(filtering))
}

// IsSteplibOfflineMode ...
func IsSteplibOfflineMode() bool {
	isSteplibOfflineMode := os.Getenv(configs.IsSteplibOfflineModeEnvKey)
	return isSteplibOfflineMode == "true"
}

// RegisterSteplibOfflineMode ...
func RegisterSteplibOfflineMode(offlineMode bool) {
	configs.IsSteplibOfflineMode = offlineMode
	// Disable analytics if running in Offline mode
	os.Setenv(analytics.DisabledEnvKey, strconv.FormatBool(offlineMode))
}
