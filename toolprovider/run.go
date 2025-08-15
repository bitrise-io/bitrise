package toolprovider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/colorstring"
	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func Run(config models.BitriseDataModel) ([]envmanModels.EnvironmentItemModel, error) {
	startTime := time.Now()
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
	case "mise":
		// At this time, we isolate Mise from any system-wide config or other Mise instances.
		// We might want to re-use the data dir of the system-wide Mise instance in the future.
		// (Local execution is not in focus yet)
		rootDir := os.Getenv("XDG_STATE_HOME")
		if rootDir == "" {
			rootDir = filepath.Join(os.Getenv("HOME"), ".local", "state")
		}
		rootDir = filepath.Join(rootDir, "bitrise", "toolprovider")
		installDir := filepath.Join(rootDir, "mise", "install")
		dataDir := filepath.Join(rootDir, "mise", "data")
		toolProvider, err = mise.NewToolProvider(installDir, dataDir)
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

	if len(toolRequests) == 0 {
		return nil, nil
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

			return nil, fmt.Errorf("install %s %s: %w", toolRequest.ToolName, toolRequest.UnparsedVersion, err)
		}
		toolInstalls = append(toolInstalls, result)
		printInstallResult(toolRequest, result, time.Since(toolStartTime))
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
