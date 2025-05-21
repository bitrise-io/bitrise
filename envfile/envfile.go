package envfile

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"gopkg.in/yaml.v3"
)

// DefaultEnvfilePathEnv is the env var that points to the (platform-dependent) envfile location.
const DefaultEnvfilePathEnv = "BITRISEIO_ENVFILE_PATH"

type EnvFile struct {
	Envs           map[string]string `yaml:"envs"`
	// TODO: naming
	ClearedEnvKeys []string          `yaml:"cleared_env_keys"`
}

// GetEnv returns the true value of an env var, even if it was not exposed to the CLI process because of its size.
// Typical large env vars are git-related build trigger env vars, like BITRISE_GIT_COMMIT_MESSAGES or BITRISE_GIT_CHANGED_FILES.
// If these were exposed as env vars to the CLI process, the execve() syscall would fail because it has a limit on
// the size of all env vars and arguments. Instead, the agent launching the Bitrise CLI process clears these env vars and
// stores their original values in a file on disk.
// Why is this whole thing not implemented in envman? Because (currently) Bitrise CLI interacts with envman through
// its CLI interface, so that subprocess exec would also fail with the same error when passing large env vars.
// Note: envfilePath must point to an existing file, you should not call this unconditionally.
func GetEnv(key string, runtimeEnvs envmanModels.EnvsJSONListModel, envfilePath string) (string, error) {
	originalBuildTriggerEnvs, err := load(envfilePath)
	if err != nil {
		return "", err
	}

	runtimeEnvValue, ok := runtimeEnvs[key]
	if !ok {
		// Same behavior as os.Getenv(key)
		return "", nil
	}

	if runtimeEnvValue == "" {
		// Env var value was possibly cleared because of its length, we should restore it from
		// the env file
		if originalValue, ok := originalBuildTriggerEnvs.Envs[key]; ok {
			return originalValue, nil
		}
		// Note: !ok means no original value found in envfile, but this can be a valid case
		// if somehow one empty env var (value) ends up in runtime envs.
	}

	// If the value is not empty, it means that it didn't hit the size limit, we can just return it
	return runtimeEnvValue, nil
}

func LogLargeEnvWarning(envfilePath string) {
	originalBuildTriggerEnvs, err := load(envfilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// We are on the critical path and should not fail here.
		log.Warnf("Failed to load envfile at $%s: %s", envfilePath, err)
		return
	}

	if len(originalBuildTriggerEnvs.ClearedEnvKeys) == 0 {
		return
	}

	clearedEnvList := ""
	for _, key := range originalBuildTriggerEnvs.ClearedEnvKeys {
		clearedEnvList += fmt.Sprintf("- %s", key)
		clearedEnvList += "\n"
	}

	message := fmt.Sprintf(`Some env vars were removed because their size would exceed system limits.
If you rely on these env vars in steps, you should read the original values from a file on disk. This file is available at $%s.
The following env vars were removed and are not available in the runtime environment:
%s`, DefaultEnvfilePathEnv, clearedEnvList)
	log.Warnf(message)
}

func load(filepath string) (EnvFile, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return EnvFile{}, err
	}

	var envFile EnvFile
	err = yaml.Unmarshal(data, &envFile)
	if err != nil {
		return EnvFile{}, err
	}

	return envFile, nil
}
