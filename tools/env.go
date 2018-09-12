package tools

import (
	"fmt"
	"os"
	"strings"
)

func expandEnvRecursive(source map[string]string, str string) string {
	mape := func(key string) string {
		if v, ok := source[key]; ok {
			return v
		}
		return ""
	}
	ret := os.Expand(str, mape)
	if ret != str {
		return expandEnvRecursive(source, ret)
	}
	return ret
}

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

// ExpandEnv ...
func ExpandEnv(key string, externalEnvs []string) (string, error) {
	externalEnvMap, err := envListToMap(externalEnvs)
	if err != nil {
		return "", err
	}
	return expandEnvRecursive(externalEnvMap, externalEnvMap[key]), nil
}
