package toolprovider

import (
	"errors"
	"fmt"
	"time"

	"github.com/bitrise-io/colorstring"
	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func Run(config models.BitriseDataModel, tracker analytics.Tracker, isCI bool, workflowID string) ([]envmanModels.EnvironmentItemModel, error) {
	startTime := time.Now()
	toolRequests, err := getToolRequests(config, workflowID)
	if err != nil {
		return nil, fmt.Errorf("tools: %w", err)
	}

	if len(toolRequests) == 0 {
		return nil, nil
	}

	toolConfig := defaultToolConfig()
	if config.ToolConfig != nil {
		if config.ToolConfig.Provider != "" {
			toolConfig.Provider = config.ToolConfig.Provider
		}
		toolConfig.ExperimentalFastInstall = config.ToolConfig.ExperimentalFastInstall
		toolConfig.ExperimentalFastInstallForce = config.ToolConfig.ExperimentalFastInstallForce
	}
	providerID := toolConfig.Provider

	var toolProvider provider.ToolProvider
	switch providerID {
	case "asdf":
		toolProvider = &asdf.AsdfToolProvider{
			ExecEnv: execenv.ExecEnv{
				// At this time, the asdf tool provider relies on the system-wide asdf install and config provided by the stack.
				EnvVars:            map[string]string{},
				ShellInit:          "",
				ClearInheritedEnvs: false,
			},
		}
	case "mise":
		miseInstallDir, miseDataDir := mise.Dirs(mise.GetMiseVersion())
		toolProvider, err = mise.NewToolProvider(miseInstallDir, miseDataDir, toolConfig)
		if err != nil {
			return nil, fmt.Errorf("create mise tool provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported tool provider: %s", providerID)
	}

	err = toolProvider.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("bootstrap %s: %w", providerID, err)
	}

	printToolRequests(toolRequests)

	var toolInstalls []provider.ToolInstallResult
	for _, toolRequest := range toolRequests {
		toolStartTime := time.Now()
		canonicalToolID := getCanonicalToolID(toolRequest.ToolName)
		toolRequest.ToolName = canonicalToolID

		printInstallStart(toolRequest)
		result, err := toolProvider.InstallTool(toolRequest)
		if err != nil {
			var toolErr provider.ToolInstallError
			if errors.As(err, &toolErr) {
				printInstallError(toolErr)
				return nil, fmt.Errorf("see error details above")
			}

			tracker.SendToolSetupEvent(providerID, toolRequest, result, false, time.Since(toolStartTime))
			return nil, fmt.Errorf("install %s %s: %w", toolRequest.ToolName, toolRequest.UnparsedVersion, err)
		}
		toolInstalls = append(toolInstalls, result)
		duration := time.Since(toolStartTime)
		printInstallResult(toolRequest, result, duration)
		tracker.SendToolSetupEvent(providerID, toolRequest, result, true, duration)
	}

	var activations []provider.EnvironmentActivation
	for _, install := range toolInstalls {
		activation, err := toolProvider.ActivateEnv(install)
		if err != nil {
			return nil, fmt.Errorf("activate %s: %w", install.ToolName, err)
		}
		activations = append(activations, activation)
	}

	duration := time.Since(startTime).Round(time.Millisecond)
	log.Printf("%s (took %s)", colorstring.Green("✓ Tool setup complete"), duration)
	log.Printf("")

	return convertToEnvmanEnvs(activations), nil
}
