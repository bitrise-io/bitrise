package local

import (
	"strconv"
	"time"

	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/bitrise-io/bitrise/v3/log"
	envmanModels "github.com/bitrise-io/bitrise/v3/envman/models"
)

// readNoOutputTimoutConfiguration resolves BITRISE_NO_OUTPUT_TIMEOUT (seconds)
// from the process env, secrets and app envs, with the precedence of
// resolveBuildEnvOverrides. Zero or a negative value disables hang detection.
func readNoOutputTimoutConfiguration(environments []envmanModels.EnvironmentItemModel) time.Duration {
	const defaultTimeout = -1
	keys := []string{configs.NoOutputTimeoutEnvKey}
	resolved := resolveBuildEnvOverrides(keys, environments, func(key, value string) bool {
		if _, err := strconv.ParseInt(value, 10, 0); err != nil {
			log.Warnf("Ignoring invalid %s value %q: must be an integer", key, value)
			return false
		}
		return true
	})

	envVal, ok := resolved[configs.NoOutputTimeoutEnvKey]
	if !ok {
		return defaultTimeout
	}

	timeout, _ := strconv.ParseInt(envVal, 10, 0)
	if timeout <= 0 {
		return defaultTimeout
	}

	return time.Duration(timeout) * time.Second
}
