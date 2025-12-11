package toolprovider

import (
	"os"
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
	if pathEnv == "" && len(pathsToAdd) == 0 {
		return ""
	}
	if pathEnv == "" {
		return strings.Join(pathsToAdd, ":")
	}
	if len(pathsToAdd) == 0 {
		return pathEnv
	}

	existingPaths := strings.Split(pathEnv, ":")

	// Create a map of existing paths for O(1) lookup
	existingPathsMap := make(map[string]bool)
	for _, p := range existingPaths {
		if p != "" {
			existingPathsMap[p] = true
		}
	}

	// Remove paths that we're about to add from the existing list to avoid duplicates
	dedupedExisting := make([]string, 0, len(existingPaths))
	for _, p := range existingPaths {
		if p == "" {
			continue
		}
		// Check if this path is in our pathsToAdd list
		shouldRemove := false
		for _, newPath := range pathsToAdd {
			if p == newPath {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			dedupedExisting = append(dedupedExisting, p)
		}
	}

	// Prepend the new paths
	allPaths := append(pathsToAdd, dedupedExisting...)
	return strings.Join(allPaths, ":")
}
