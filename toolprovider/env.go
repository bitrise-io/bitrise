package toolprovider

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

func ConvertToEnvMap(activations []provider.EnvironmentActivation) map[string]string {
	pathValue := os.Getenv("PATH")

	envMap := make(map[string]string)
	for _, activation := range activations {
		for k, v := range activation.ContributedEnvVars {
			envMap[k] = v
		}
	}

	var newPathEntries []string
	for _, act := range activations {
		for _, p := range act.ContributedPaths {
			if p != "" {
				newPathEntries = append(newPathEntries, p)
			}
		}
	}

	if len(newPathEntries) > 0 {
		newPath := prependPaths(pathValue, newPathEntries)
		if newPath != "" {
			envMap["PATH"] = newPath
		}
	}

	return envMap
}

func ConvertToEnvmanEnvs(activations []provider.EnvironmentActivation) []envmanModels.EnvironmentItemModel {
	envMap := ConvertToEnvMap(activations)

	envs := make([]envmanModels.EnvironmentItemModel, 0, len(envMap))
	for k, v := range envMap {
		envs = append(envs, envmanModels.EnvironmentItemModel{
			k: v,
		})
	}

	return envs
}

func prependPaths(pathEnv string, pathsToAdd []string) string {
	if pathEnv == "" {
		return strings.Join(pathsToAdd, ":")
	}
	if len(pathsToAdd) == 0 {
		return pathEnv
	}

	pathItems := strings.Split(pathEnv, ":")
	pathItems = slices.DeleteFunc(pathItems, func(p string) bool {
		// Remove any paths that are in pathsToAdd to avoid duplicates
		// We'll prepend them anyway, no point in keeping them in the existing list
		return slices.Contains(pathsToAdd, p) || p == ""
	})

	if len(pathItems) == 0 {
		return strings.Join(pathsToAdd, ":")
	}

	return fmt.Sprintf("%s:%s", strings.Join(pathsToAdd, ":"), strings.Join(pathItems, ":"))
}
