package toolprovider

import (
	"errors"
	"fmt"
	"time"

	"github.com/bitrise-io/colorstring"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func RunDeclarativeSetup(config models.BitriseDataModel, tracker analytics.Tracker, isCI bool, workflowID string, silent bool) ([]provider.EnvironmentActivation, error) {
	toolRequests, err := getToolRequests(config, workflowID)
	if err != nil {
		return nil, fmt.Errorf("tools: %w", err)
	}

	if len(toolRequests) == 0 {
		return nil, nil
	}

	provider := selectProvider(config)
	useFastInstall := selectFastInstall(config)

	return installTools(toolRequests, provider, useFastInstall, tracker, silent)
}

func installTools(toolRequests []provider.ToolRequest, providerID string, useFastInstall bool, tracker analytics.Tracker, silent bool) ([]provider.EnvironmentActivation, error) {
	startTime := time.Now()

	log.Debugf("[TOOLPROVIDER] Install tools using provider: %s, fast install: %v", providerID, useFastInstall)

	var toolProvider provider.ToolProvider
	var err error

	if useFastInstall && !silent {
		log.Printf("")
		log.Warn("Using fast Ruby install because running on edge stack. This behavior is going to be the default on stable stacks soon. If you notice issues, switch to a stable stack temporarily and let us know at https://github.com/bitrise-io/bitrise/issues/new?title=Fast%20tool%20install%20issue:%20")
	}

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
		toolProvider, err = mise.NewToolProvider(miseInstallDir, miseDataDir, useFastInstall)
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

	if !silent {
		printToolRequests(toolRequests)
	}

	var toolInstalls []provider.ToolInstallResult
	for _, toolRequest := range toolRequests {
		toolStartTime := time.Now()
		canonicalToolID := alias.GetCanonicalToolID(toolRequest.ToolName)
		toolRequest.ToolName = canonicalToolID

		if !silent {
			printInstallStart(toolRequest)
		}

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
		if !silent {
			printInstallResult(toolRequest, result, duration)
		}
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

	if !silent {
		duration := time.Since(startTime).Round(time.Millisecond)
		log.Printf("%s (took %s)", colorstring.Green("✓ Tool setup complete"), duration)
		log.Printf("")
	}

	return activations, nil
}

// InstallSingleTool installs a single tool with the specified version using the given provider.
// This is a convenience wrapper around installTools for installing just one tool.
func InstallSingleTool(toolRequest provider.ToolRequest, providerID string, useFastInstall bool, tracker analytics.Tracker, silent bool) ([]provider.EnvironmentActivation, error) {
	return installTools([]provider.ToolRequest{toolRequest}, providerID, useFastInstall, tracker, silent)
}
