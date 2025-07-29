package asdf

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider"
)

func (a AsdfToolProvider) ActivateEnv(result toolprovider.ToolInstallResult) (toolprovider.EnvironmentActivation, error) {
	envKey := fmt.Sprint("ASDF_", strings.ToUpper(string(result.ToolName)), "_VERSION")
	return toolprovider.EnvironmentActivation{
		ContributedEnvVars: map[string]string{
			envKey: result.ConcreteVersion,
		},
		ContributedPaths: []string{}, // TODO: shims dir?
	}, nil
}
