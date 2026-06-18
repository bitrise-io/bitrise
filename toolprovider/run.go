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
	"github.com/bitrise-io/bitrise/v2/toolprovider/versionresolver"
)

type toolSetupResult struct {
	request   provider.ToolRequest
	result    provider.ToolInstallResult
	startTime time.Time
}

// knownGitHubTokenEnvVars lists the env var names that users possibly set as a valid GitHub API token.
var knownGitHubTokenEnvVars = []string{
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITHUB_API_TOKEN",
	"MISE_GITHUB_TOKEN",
	"MISE_GITHUB_ENTERPRISE_TOKEN",
}

// findGitHubTokenEnv returns the first known GitHub token env var found in envs,
// checked in priority order defined by knownGitHubTokenEnvVars.
func findGitHubTokenEnv(envs map[string]string) (name, value string, found bool) {
	for _, n := range knownGitHubTokenEnvVars {
		if v, ok := envs[n]; ok {
			return n, v, true
		}
	}
	return "", "", false
}

func RunDeclarativeSetup(config models.BitriseDataModel, tracker analytics.Tracker, isCI bool, workflowID string, silent bool, providerOverride *string, fastInstallOverride *bool, envs map[string]string) ([]provider.EnvironmentActivation, error) {
	toolRequests, err := getToolRequests(config, workflowID)
	if err != nil {
		return nil, fmt.Errorf("tools: %w", err)
	}

	if len(toolRequests) == 0 {
		return nil, nil
	}

	var provider string
	if providerOverride != nil {
		provider = *providerOverride
	} else {
		provider = selectProvider(config)
	}

	useFastInstall := DefaultFastInstall()
	if config.ToolConfig != nil && config.ToolConfig.FastInstall != nil {
		useFastInstall = *config.ToolConfig.FastInstall
	}
	if fastInstallOverride != nil {
		useFastInstall = *fastInstallOverride
	}

	var extraEnvs map[string]string
	if tokenName, tokenValue, ok := findGitHubTokenEnv(envs); ok {
		if !silent {
			log.Printf("Using %s for GitHub API authentication during tool setup", tokenName)
		}
		// Mise recognizes [a variety of env vars](// See https://mise.jdx.dev/dev-tools/github-tokens.html), but let's use
		// a generic one because this layer is provider-agnostic.
		extraEnvs = map[string]string{"GITHUB_TOKEN": tokenValue}
	}

	return installTools(toolRequests, provider, useFastInstall, tracker, silent, extraEnvs)
}

func installTools(toolRequests []provider.ToolRequest, providerID string, useFastInstall bool, tracker analytics.Tracker, silent bool, extraEnvs map[string]string) ([]provider.EnvironmentActivation, error) {
	startTime := time.Now()

	if !silent {
		log.Debugf("[TOOLPROVIDER] Install tools using provider: %s, fast install: %v", providerID, useFastInstall)
	}
	toolProvider, err := CreateProvider(providerID, useFastInstall, silent, extraEnvs)
	if err != nil {
		return nil, err
	}

	for i, req := range toolRequests {
		if req.ResolutionStrategy != provider.ResolutionStrategyConstraint {
			continue
		}

		canonicalToolID := alias.GetCanonicalToolID(req.ToolName)
		versions, err := toolProvider.ListReleasedVersions(canonicalToolID)
		if err != nil {
			return nil, fmt.Errorf("list versions for %s: %w", canonicalToolID, err)
		}

		resolved, err := versionresolver.ResolveConstraint(req.UnparsedVersion, versions)
		if err != nil {
			return nil, fmt.Errorf("resolve constraint %q for %s: %w", req.UnparsedVersion, canonicalToolID, err)
		}

		if !silent {
			log.Debugf("[TOOLPROVIDER] Resolved %s constraint %q to version %s", canonicalToolID, req.UnparsedVersion, resolved)
		}

		toolRequests[i] = provider.ToolRequest{
			ToolName:           req.ToolName,
			UnparsedVersion:    resolved,
			ResolutionStrategy: provider.ResolutionStrategyStrict,
			PluginURL:          req.PluginURL,
		}
	}

	if !silent {
		printToolRequests(toolRequests)
	}

	return installResolvedTools(toolRequests, providerID, toolProvider, tracker, silent, startTime)
}

func installResolvedTools(
	toolRequests []provider.ToolRequest,
	providerID string,
	toolProvider provider.ToolProvider,
	tracker analytics.Tracker,
	silent bool,
	startTime time.Time,
) ([]provider.EnvironmentActivation, error) {
	var toolSetups []toolSetupResult
	for _, toolRequest := range toolRequests {
		toolStartTime := time.Now()
		canonicalToolID := alias.GetCanonicalToolID(toolRequest.ToolName)
		toolRequest.ToolName = canonicalToolID

		if !silent {
			printInstallStart(toolRequest)
		}

		result, err := toolProvider.InstallTool(toolRequest)
		if err != nil {
			tracker.SendToolSetupEvent(providerID, toolRequest, result, false, time.Since(toolStartTime))

			var toolErr provider.ToolInstallError
			if errors.As(err, &toolErr) {
				printInstallError(toolErr)
				return nil, fmt.Errorf("see error details above")
			}

			return nil, fmt.Errorf("install %s %s: %w", toolRequest.ToolName, toolRequest.UnparsedVersion, err)
		}

		toolSetups = append(toolSetups, toolSetupResult{
			request:   toolRequest,
			result:    result,
			startTime: toolStartTime,
		})

		duration := time.Since(toolStartTime)
		if !silent {
			printInstallResult(toolRequest, result, duration)
		}
	}

	var activations []provider.EnvironmentActivation
	for _, setup := range toolSetups {
		activation, err := toolProvider.ActivateEnv(setup.result)
		if err != nil {
			tracker.SendToolSetupEvent(providerID, setup.request, setup.result, false, time.Since(setup.startTime))
			return nil, fmt.Errorf("activate %s: %w", setup.result.ToolName, err)
		}
		activations = append(activations, activation)
		tracker.SendToolSetupEvent(providerID, setup.request, setup.result, true, time.Since(setup.startTime))
	}

	if !silent {
		duration := time.Since(startTime).Round(time.Millisecond)
		log.Printf("%s (took %s)", colorstring.Green("✓ Tool setup complete"), duration)
		log.Printf("")
	}

	return activations, nil
}

// InstallSingleTool installs a single tool with the specified version using the given provider.
func InstallSingleTool(toolRequest provider.ToolRequest, providerID string, useFastInstall bool, tracker analytics.Tracker, silent bool) ([]provider.EnvironmentActivation, error) {
	// extraEnvs=nil: this runs as a CLI subcommand from a user's shell, so secrets are already in the process env.
	return installTools([]provider.ToolRequest{toolRequest}, providerID, useFastInstall, tracker, silent, nil)
}

// GetLatestVersion queries the latest version of a tool without installing it (installed or released).
// Supports both mise and asdf providers.
func GetLatestVersion(toolRequest provider.ToolRequest, providerID string, useFastInstall bool, silent bool) (string, error) {
	canonicalToolID := alias.GetCanonicalToolID(toolRequest.ToolName)
	toolRequest.ToolName = canonicalToolID

	switch providerID {
	case "asdf":
		asdfProvider := &asdf.AsdfToolProvider{
			ExecEnv: execenv.ExecEnv{
				EnvVars:            map[string]string{},
				ShellInit:          "",
				ClearInheritedEnvs: false,
			},
			Silent: silent,
		}
		return asdfProvider.ResolveLatestVersion(toolRequest)
	case "mise":
		miseInstallDir, miseDataDir := mise.Dirs(mise.GetMiseVersion())
		// extraEnvs=nil: this runs as a CLI subcommand from a user's shell, so secrets are already in the process env.
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, useFastInstall, silent, nil)
		if err != nil {
			return "", fmt.Errorf("create mise tool provider: %w", err)
		}

		err = miseProvider.Bootstrap()
		if err != nil {
			return "", fmt.Errorf("bootstrap mise: %w", err)
		}

		return miseProvider.ResolveLatestVersion(toolRequest)
	default:
		return "", fmt.Errorf("unsupported tool provider: %s", providerID)
	}
}
