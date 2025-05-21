package envfile

import (
	"os"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	"gopkg.in/yaml.v3"
)

type EnvFile struct {
	Envs map[string]string `yaml:"envs"`
}

func MergeEnvfileWithRuntimeEnvs(runtimeEnvs envmanModels.EnvsJSONListModel, envFilePath string) (map[string]string, error) {
	originalBuildTriggerEnvs, err := load(envFilePath)
	if err != nil {
		return nil, err
	}

	for k, v := range runtimeEnvs {
		if v == "" {
			// Env var value was possibly cleared because of its length, we should restore it from
			// the original env file
			if originalValue, ok := originalBuildTriggerEnvs[k]; ok {
				runtimeEnvs[k] = originalValue
			}
		}
	}

	return runtimeEnvs, nil
}

func LogLargeEnvWarning(envfilePath string) {
	
}

func load(filepath string) (map[string]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var envFile EnvFile
	err = yaml.Unmarshal(data, &envFile)
	if err != nil {
		return nil, err
	}

	if envFile.Envs == nil {
		return make(map[string]string), nil
	}

	return envFile.Envs, nil
}

