package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/envman/v2/models"
)

func envListToMap(envs []string) (map[string]string, error) {
	envMap := map[string]string{}
	for _, env := range envs {
		s := strings.Split(env, "=")
		if len(s) < 2 {
			return nil, fmt.Errorf("key should be separated from value by '=' character: %s", env)
		}
		key := s[0]
		value := strings.Join(s[1:], "=")
		envMap[key] = value
	}
	return envMap, nil
}

// ExpandEnvItems ...
func ExpandEnvItems(toExpand []models.EnvironmentItemModel, externalEnvs []string) (map[string]string, error) {
	externalEnvMap, err := envListToMap(externalEnvs)
	if err != nil {
		return nil, err
	}

	mapper := func(key string) string {
		return externalEnvMap[key]
	}

	expanded := map[string]string{}
	for _, env := range toExpand {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return nil, err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return nil, err
		}

		if opts.SkipIfEmpty != nil && *opts.SkipIfEmpty && value == "" {
			continue
		}

		if opts.IsExpand != nil && *opts.IsExpand {
			value = os.Expand(value, mapper)
		}

		externalEnvMap[key] = value
		expanded[key] = value
	}

	return expanded, nil
}

func LimitEnvVarValue(value string, limitInBytes int) (string, bool) {
	if limitInBytes < 5 {
		// limit indicator is '...' and
		// the minimal limit is the length of the indicator + 1 leading and 1 trailing character (1byte each)
		return value, false
	}

	if len(value) <= limitInBytes {
		return value, false
	}

	// Calculate the length of the prefix and suffix
	prefixLength := (limitInBytes - 3) / 2
	suffixLength := limitInBytes - 3 - prefixLength

	// Trim the middle of the value and insert '...'
	trimmedValue := value[:prefixLength] + "..." + value[len(value)-suffixLength:]

	return trimmedValue, true
}
