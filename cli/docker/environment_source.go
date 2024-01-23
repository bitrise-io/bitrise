package docker

import (
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/log"
)

// implementing env.EnvironmentSource
type DockerEnvironmentSource struct {
	Logger log.Logger
}

// GetEnvironment ...
// Where envman.ReadAndEvaluateEnvs(configs.InputEnvstorePath, env.EnvironmentSource) is called,
// and we are in the context of using containers, we cannot use the default env.EnvironmentSource
// implementation, as it promotes all the envs from the host to the container, which is not what we want.
// for instance, we may have envs inherited from Bitrise stacks, altering default behavior of certain
// containers (for instance Java).
// Instead, we have our own implementation, filtering for envs that are whitelisted, and that are the envs
// starting with BITRISE_, and additionally the PATH, PR, CI and ENVMAN_ENVSTORE_PATH envs.
func (des *DockerEnvironmentSource) GetEnvironment() map[string]string {
	passthroughEnvsList := strings.Split(os.Getenv("BITRISE_DOCKER_PASSTHROUGH_ENVS"), ",")
	passthroughEnvsList = append(passthroughEnvsList, "PR", "CI", "ENVMAN_ENVSTORE_PATH")
	dockerPassthroughEnvsMap := make(map[string]bool)
	for _, k := range passthroughEnvsList {
		dockerPassthroughEnvsMap[k] = true
	}

	processEnvs := os.Environ()
	envs := make(map[string]string)

	// String names can be duplicated (on Unix), and the Go libraries return the first instance of them:
	// https://github.com/golang/go/blob/98d20fb23551a7ab900fcfe9d25fd9cb6a98a07f/src/syscall/env_unix.go#L45
	// From https://pubs.opengroup.org/onlinepubs/9699919799/:
	// > "There is no meaning associated with the order of strings in the environment.
	// > If more than one string in an environment of a process has the same name, the consequences are undefined."
	des.Logger.Infof("**** all envs: %+v", processEnvs)
	for _, env := range processEnvs {
		key, value := des.splitEnv(env)
		_, allowed := dockerPassthroughEnvsMap[key]
		if !strings.HasPrefix(key, "BITRISE") && (key == "" || !allowed) {
			des.Logger.Infof("**** disallowed env: %s", key)
			continue
		}

		envs[key] = value
	}

	return envs
}

// SplitEnv splits an env returned by os.Environ
func (des *DockerEnvironmentSource) splitEnv(env string) (key string, value string) {
	const sep = "="
	split := strings.SplitAfterN(env, sep, 2)
	if split == nil {
		return "", ""
	}
	key = strings.TrimSuffix(split[0], sep)
	if len(split) > 1 {
		value = split[1]
	}
	return
}
