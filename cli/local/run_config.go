package local

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/bitrise-io/bitrise/v3/log"
	envmanModels "github.com/bitrise-io/bitrise/v3/envman/models"
)

const noOutputTimeoutDisabled = -1

// readNoOutputTimoutConfiguration resolves BITRISE_NO_OUTPUT_TIMEOUT (seconds)
// from the process env, secrets and app envs, with the precedence of
// resolveBuildEnvOverrides. Zero or a negative value disables hang detection.
func readNoOutputTimoutConfiguration(environments []envmanModels.EnvironmentItemModel) time.Duration {
	keys := []string{configs.NoOutputTimeoutEnvKey}
	resolved := resolveBuildEnvOverrides(keys, environments, func(key, value string) bool {
		if _, err := parseNoOutputTimeout(value); err != nil {
			log.Warnf("Ignoring invalid %s value %q: %s", key, value, err)
			return false
		}
		return true
	})

	envVal, ok := resolved[configs.NoOutputTimeoutEnvKey]
	if !ok {
		return noOutputTimeoutDisabled
	}

	// Validated by resolveBuildEnvOverrides.
	timeout, _ := parseNoOutputTimeout(envVal)
	return timeout
}

func parseNoOutputTimeout(value string) (time.Duration, error) {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("must be an integer")
	}
	if seconds > math.MaxInt64/int64(time.Second) {
		return 0, fmt.Errorf("must be at most %d", math.MaxInt64/int64(time.Second))
	}
	if seconds <= 0 {
		return noOutputTimeoutDisabled, nil
	}

	return time.Duration(seconds) * time.Second, nil
}

// resolveBuildEnvOverrides resolves CLI configuration that a build sets through
// an env var before the workflow runs, when only the raw declarations exist:
// values are not expanded, so `$OTHER_VAR` is taken literally and fails
// validation. Workflow-level envs are not considered, as the workflow to run
// can still change.
//
// environments lists the secrets first, then the app envs, so precedence is
// process env < secret < app env, and a later duplicate wins as it does in
// envman; the highest-precedence valid value of each key wins. Empty values are
// ignored and invalid ones are skipped as they are considered, so a malformed
// higher-precedence value never discards a valid lower-precedence one. Keys
// with no valid value are absent from the result.
func resolveBuildEnvOverrides(keys []string, environments []envmanModels.EnvironmentItemModel, isValid func(key, value string) bool) map[string]string {
	resolved := map[string]string{}

	consider := func(key, value string) {
		if !slices.Contains(keys, key) || value == "" || !isValid(key, value) {
			return
		}
		resolved[key] = value
	}

	for _, key := range keys {
		consider(key, os.Getenv(key))
	}
	for _, env := range environments {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			continue
		}
		consider(key, value)
	}

	return resolved
}
