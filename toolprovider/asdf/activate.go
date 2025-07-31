package asdf

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func (a AsdfToolProvider) ActivateEnv(result provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	envKey := fmt.Sprint("ASDF_", strings.ToUpper(string(result.ToolName)), "_VERSION")
	return provider.EnvironmentActivation{
		ContributedEnvVars: map[string]string{
			envKey: result.ConcreteVersion,
		},
		ContributedPaths: []string{}, // TODO: shims dir?
	}, nil
}
