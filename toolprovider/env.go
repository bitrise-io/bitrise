package toolprovider

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

func ConvertToEnvmanEnvs(activations []provider.EnvironmentActivation, currentPath *string) []envmanModels.EnvironmentItemModel {
	usedPath := ""
	if currentPath == nil {
		path := os.Getenv("PATH")
		currentPath = &path
	} else {
		usedPath = *currentPath
	}

	envs := make([]envmanModels.EnvironmentItemModel, 0)
	for _, activation := range activations {
		for k, v := range activation.ContributedEnvVars {
			envs = append(envs, envmanModels.EnvironmentItemModel{
				k: v,
			})
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
		newPath := prependPath(usedPath, strings.Join(newPathEntries, ":"))
		if newPath != "" {
			envs = append(envs, envmanModels.EnvironmentItemModel{
				"PATH": newPath,
			})
		}
	}

	return envs
}

func prependPath(pathEnv, addition string) string {
	if pathEnv == "" {
		return addition
	}

	pathItems := strings.Split(pathEnv, ":")
	pathItems = slices.DeleteFunc(pathItems, func(p string) bool {
		return p == addition
	})

	if len(pathItems) == 0 {
		return addition
	}

	return fmt.Sprintf("%s:%s", addition, strings.Join(pathItems, ":"))
}
