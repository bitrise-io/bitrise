package toolprovider

import (
	"fmt"

	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func Run(config models.BitriseDataModel) ([]envmanModels.EnvironmentItemModel, error) {
	toolRequests, err := getToolRequests(config)
	if err != nil {
		return nil, fmt.Errorf("tools: %w", err)
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
		return nil, fmt.Errorf("unsupported tool provider: %s", providerID)
	}

	err = toolProvider.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("bootstrap %s: %w", providerID, err)
	}

	if len(toolRequests) == 0 {
		return nil, nil
	}

	printToolRequests(toolRequests)

	var toolInstalls []provider.ToolInstallResult
	for _, toolRequest := range toolRequests {
		canonicalToolID := getCanonicalToolID(toolRequest.ToolName)
		toolRequest.ToolName = canonicalToolID

		log.Printf("Installing %s %s...\n", canonicalToolID, toolRequest.UnparsedVersion)
		result, err := toolProvider.InstallTool(toolRequest)
		if err != nil {
			// TODO: better error message
			return nil, fmt.Errorf("install %s %s: %w", toolRequest.ToolName, toolRequest.UnparsedVersion, err)
		}
		toolInstalls = append(toolInstalls, result)

		if result.IsAlreadyInstalled {
			log.Printf("%s %s is already installed.\n", result.ToolName, result.ConcreteVersion)
		} else {
			log.Printf("Successfully installed %s %s.\n", result.ToolName, result.ConcreteVersion)
		}
	}

	var activations []provider.EnvironmentActivation
	for _, install := range toolInstalls {
		activation, err := toolProvider.ActivateEnv(install)
		if err != nil {
			return nil, fmt.Errorf("activate %s: %w", install.ToolName, err)
		}
		activations = append(activations, activation)
	}
	fmt.Print("Environment setup complete!")

	return convertToEnvmanEnvs(activations), nil
}

func printToolRequests(toolRequests []provider.ToolRequest) {
	log.Info("Tools to set up:")

	for _, toolRequest := range toolRequests {
		strategy := "strict" // default
		switch toolRequest.ResolutionStrategy {
		case provider.ResolutionStrategyLatestInstalled:
			strategy = "latest installed"
		case provider.ResolutionStrategyLatestReleased:
			strategy = "latest released"
		}

		log.Printf("- %s %s (resolution: %s)\n",
			toolRequest.ToolName,
			toolRequest.UnparsedVersion,
			strategy)
	}

	if len(toolRequests) > 0 {
		log.Infof("Installing missing tools...")
	}
}
