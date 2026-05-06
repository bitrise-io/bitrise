package toolprovider

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// CreateProvider constructs and bootstraps a ToolProvider for the given provider ID.
// extraEnvs are additional environment variables passed to the provider (e.g. GitHub tokens).
// Pass nil for extraEnvs when running as a CLI subcommand from a user's shell.
func CreateProvider(providerID string, useFastInstall bool, silent bool, extraEnvs map[string]string) (provider.ToolProvider, error) {
	var tp provider.ToolProvider
	var err error

	switch providerID {
	case "asdf":
		tp = &asdf.AsdfToolProvider{
			ExecEnv: execenv.ExecEnv{
				EnvVars:            map[string]string{},
				ShellInit:          "",
				ClearInheritedEnvs: false,
			},
			Silent: silent,
		}
	case "mise":
		miseInstallDir, miseDataDir := mise.Dirs(mise.GetMiseVersion())
		tp, err = mise.NewToolProvider(miseInstallDir, miseDataDir, useFastInstall, silent, extraEnvs)
		if err != nil {
			return nil, fmt.Errorf("create mise tool provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported tool provider: %s", providerID)
	}

	if err := tp.Bootstrap(); err != nil {
		return nil, fmt.Errorf("bootstrap %s: %w", providerID, err)
	}

	return tp, nil
}
