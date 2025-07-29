package toolprovider

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func Run(config models.BitriseDataModel) error {
	toolRequests, err := getToolRequests(config)
	if err != nil {
		return fmt.Errorf("tools: %w", err)
	}
	providerID := defaultToolConfig().Provider
	if config.ToolConfig != nil {
		if config.ToolConfig.Provider != "" {
			providerID = config.ToolConfig.Provider
		}
	}

	var toolProvider provider.ToolProvider
	switch providerID {
	case "asdf":
		toolProvider = asdf.AsdfToolProvider{
			ExecEnv: execenv.ExecEnv{
				// At this time, the asdf tool provider relies on the system-wide asdf install and config provided by the stack.
				EnvVars:            map[string]string{},
				ShellInit:          "",
				ClearInheritedEnvs: false,
			},
		}
	default:
		return fmt.Errorf("unsupported tool provider: %s", providerID)
	}

	err = toolProvider.Bootstrap()
	if err != nil {
		return fmt.Errorf("bootstrap %s: %w", providerID, err)
	}

	if len(toolRequests) == 0 {
		log.Info("No tools to set up.")
		return nil
	}

	// TODO: continue

	return nil
}
