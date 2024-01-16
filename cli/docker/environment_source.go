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

func (des *DockerEnvironmentSource) GetEnvironment() map[string]string {
	passthroughEnvsList := strings.Split(os.Getenv("BITRISE_DOCKER_PASSTHROUGH_ENVS"), ",")
	passthroughEnvsList = append(passthroughEnvsList, "PATH", "PR", "ENVMAN_ENVSTORE_PATH")
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
	for _, env := range processEnvs {
		key, value := des.splitEnv(env)
		_, allowed := dockerPassthroughEnvsMap[key]
		if !strings.HasPrefix(key, "BITRISE") && (key == "" || !allowed) {
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
