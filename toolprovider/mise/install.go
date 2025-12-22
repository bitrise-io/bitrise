package mise

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/nixpkgs"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func installRequest(toolRequest provider.ToolRequest, useNix bool) provider.ToolRequest {
	if useNix {
		return provider.ToolRequest{
			// Use Mise's backend plugin convention of pluginID:toolID.
			ToolName:           provider.ToolID(fmt.Sprintf("%s:%s", nixpkgs.PluginName, toolRequest.ToolName)),
			UnparsedVersion:    toolRequest.UnparsedVersion,
			ResolutionStrategy: toolRequest.ResolutionStrategy,
			// Only relevant for plugins, that are not handled by the given backend.
			// Nixpkgs handles all tools it supports internally, we should not install anything extra.
			PluginURL: nil,
		}
	} else {
		return toolRequest
	}
}

// nixChecker is a helper for testing.
// The real implementation returns true if Nix (the daemon) is available on the system and various other conditions are met.
type nixChecker func(tool provider.ToolRequest) bool

func canBeInstalledWithNix(tool provider.ToolRequest, execEnv execenv.ExecEnv, useFastInstall bool, nixChecker nixChecker) bool {
	// Force switch for integration testing. No fallback to regular install when this is active. This makes failures explicit.
	forceNix := os.Getenv("BITRISE_TOOLSETUP_FAST_INSTALL_FORCE") == "true"
	useNix := nixChecker(tool)

	canProceed := (useFastInstall && useNix) || forceNix
	if !canProceed {
		return false
	}

	// Enable experimental settings for custom backend
	if _, err := execEnv.RunMise("settings", "experimental=true"); err != nil {
		log.Warnf("Error while enabling experimental settings: %v.", err)
		return forceNix
	}

	// If the plugin is already installed, Mise will not throw an error.
	_, err := execEnv.RunMisePlugin("install", nixpkgs.PluginName, nixpkgs.PluginGitURL)
	if err != nil {
		log.Warnf("Error while installing nixpkgs plugin (%s): %v.", nixpkgs.PluginGitURL, err)
		return forceNix
	}

	// Note: even we just installed the plugin above, it might get preinstalled to the environment one day. To be safe, we update it here
	// because the index might be outdated.
	_, err = execEnv.RunMisePlugin("update", nixpkgs.PluginName)
	if err != nil {
		log.Warnf("Error while updating nixpkgs plugin (%s): %v. Possibly using outdated plugin version.", nixpkgs.PluginGitURL, err)
	}

	if forceNix {
		// In force mode, we do not care about version existence, as failure is expected if the version is not in nixpkgs.
		// But we still need to make sure the plugin above is installed.
		return true
	}

	nameWithBackend := provider.ToolID(fmt.Sprintf("nixpkgs:%s", tool.ToolName))
	available, err := versionExists(execEnv, nameWithBackend, tool.UnparsedVersion)
	if err != nil {
		log.Warnf("Error while checking nixpkgs index for %s@%s: %v. Falling back to core plugin installation.", tool.ToolName, tool.UnparsedVersion, err)
		return false
	}
	if !available {
		log.Warnf("%s@%s not found in nixpkgs index, doing a source build. This may take some time...", tool.ToolName, tool.UnparsedVersion)
		return false
	}

	return true
}

func (m *MiseToolProvider) installToolVersion(tool provider.ToolRequest) error {
	versionString, err := miseVersionString(tool, m.resolveToLatestInstalled)
	if err != nil {
		return err
	}

	output, err := m.ExecEnv.RunMiseWithTimeout(execenv.InstallTimeout, "install", "--yes", versionString)
	if err != nil {
		return provider.ToolInstallError{
			ToolName:         tool.ToolName,
			RequestedVersion: tool.UnparsedVersion,
			Cause:            fmt.Sprintf("mise install %s: %s", versionString, err),
			RawOutput:        string(output),
		}
	}
	return nil
}

// Helper for easier testing.
// Inputs: tool ID, tool version.
// Returns: latest installed version of the tool, or an error if no matching version is installed.
type latestResolver func(provider.ToolID, string) (string, error)

func (m *MiseToolProvider) isAlreadyInstalled(tool provider.ToolRequest) (bool, error) {
	latestInstalledResolver := func(toolName provider.ToolID, versionPrefix string) (string, error) {
		if _, err := m.resolveToLatestInstalled(toolName, versionPrefix); err != nil {
			return "", err
		}
		// This is a secondary check for installed versions as a list too, because 'latest --installed tool@version' command
		// is not reliable.
		exists, err := versionExistsLocal(m.ExecEnv, toolName, versionPrefix)
		if err != nil {
			return "", err
		}
		if !exists {
			return "", errNoMatchingVersion
		}
		return versionPrefix, nil
	}

	return isAlreadyInstalled(tool, latestInstalledResolver, m.resolveToLatestReleased)
}

func isAlreadyInstalled(tool provider.ToolRequest, latestInstalledResolver, latestReleasedResolver latestResolver) (bool, error) {
	toolVersion := tool.UnparsedVersion
	if tool.ResolutionStrategy == provider.ResolutionStrategyLatestReleased {
		// User might gave an incomplete version string, need to resolve to the full version first,
		// so we compare the wanted version to installed versions.
		// e.g. 3.3 -> 3.3.9
		v, err := latestReleasedResolver(tool.ToolName, toolVersion)
		if err != nil {
			return false, err
		}
		toolVersion = v
	}

	_, err := latestInstalledResolver(tool.ToolName, toolVersion)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, errNoMatchingVersion) {
		return false, nil
	}

	return false, err
}

func miseVersionString(tool provider.ToolRequest, latestInstalledResolver latestResolver) (string, error) {
	var miseVersionString string
	resolutionStrategy := tool.ResolutionStrategy
	if tool.UnparsedVersion == "installed" {
		resolutionStrategy = provider.ResolutionStrategyLatestInstalled
	}

	switch resolutionStrategy {
	case provider.ResolutionStrategyStrict:
		miseVersionString = fmt.Sprintf("%s@%s", tool.ToolName, tool.UnparsedVersion)
	case provider.ResolutionStrategyLatestReleased:
		// https://mise.jdx.dev/configuration.html#scopes
		miseVersionString = fmt.Sprintf("%s@prefix:%s", tool.ToolName, tool.UnparsedVersion)
	case provider.ResolutionStrategyLatestInstalled:
		latestInstalledV, err := latestInstalledResolver(tool.ToolName, tool.UnparsedVersion)
		if err == nil {
			miseVersionString = fmt.Sprintf("%s@%s", tool.ToolName, latestInstalledV)
		} else {
			if errors.Is(err, errNoMatchingVersion) {
				// No local version satisfies the request -> fallback to latest released.
				miseVersionString = fmt.Sprintf("%s@prefix:%s", tool.ToolName, tool.UnparsedVersion)
			} else {
				return "", fmt.Errorf("resolve %s %s to latest installed version: %w", tool.ToolName, tool.UnparsedVersion, err)
			}
		}
	default:
		return "", fmt.Errorf("unknown resolution strategy: %v", tool.ResolutionStrategy)
	}
	return miseVersionString, nil

}
