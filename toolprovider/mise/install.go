package mise

import (
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
type nixChecker func(tool provider.ToolRequest, silent bool) bool

func canBeInstalledWithNix(tool provider.ToolRequest, execEnv execenv.ExecEnv, useFastInstall bool, nixChecker nixChecker, silent bool) bool {
	// Force switch for integration testing. No fallback to regular install when this is active. This makes failures explicit.
	forceNix := os.Getenv("BITRISE_TOOLSETUP_FAST_INSTALL_FORCE") == "true"
	useNix := nixChecker(tool, silent)

	canProceed := (useFastInstall && useNix) || forceNix
	if !canProceed {
		return false
	}

	// Enable experimental settings for custom backend
	if _, err := execEnv.RunMise("settings", "experimental=true"); err != nil {
		if !silent {
			log.Warnf("Error while enabling experimental settings: %v.", err)
		}
		return forceNix
	}

	// If the plugin is already installed, Mise will not throw an error.
	_, err := execEnv.RunMisePlugin("install", nixpkgs.PluginName, nixpkgs.PluginGitURL)
	if err != nil {
		if !silent {
			log.Warnf("Error while installing nixpkgs plugin (%s): %v.", nixpkgs.PluginGitURL, err)
		}
		return forceNix
	}

	// Note: even we just installed the plugin above, it might get preinstalled to the environment one day. To be safe, we update it here
	// because the index might be outdated.
	_, err = execEnv.RunMisePlugin("update", nixpkgs.PluginName)
	if err != nil {
		if !silent {
			log.Warnf("Error while updating nixpkgs plugin (%s): %v. Possibly using outdated plugin version.", nixpkgs.PluginGitURL, err)
		}
	}

	if forceNix {
		// In force mode, we do not care about version existence, as failure is expected if the version is not in nixpkgs.
		// But we still need to make sure the plugin above is installed.
		return true
	}

	nameWithBackend := provider.ToolID(fmt.Sprintf("nixpkgs:%s", tool.ToolName))
	available, err := versionExistsRemote(execEnv, nameWithBackend, tool.UnparsedVersion)
	if err != nil {
		if !silent {
			log.Warnf("Error while checking nixpkgs index for %s@%s: %v. Falling back to core plugin installation.", tool.ToolName, tool.UnparsedVersion, err)
		}
		return false
	}
	if !available {
		if !silent {
			log.Warnf("%s@%s not found in nixpkgs index, doing a source build. This may take some time...", tool.ToolName, tool.UnparsedVersion)
		}
		return false
	}

	return true
}

func (m *MiseToolProvider) installToolVersion(toolName provider.ToolID, concreteVersion string) error {
	versionString := miseVersionString(toolName, concreteVersion)

	output, err := m.ExecEnv.RunMiseWithTimeout(execenv.InstallTimeout, "install", "--yes", versionString)
	if err != nil {
		return provider.ToolInstallError{
			ToolName:         toolName,
			RequestedVersion: concreteVersion,
			Cause:            fmt.Sprintf("mise install %s: %s", versionString, err),
			RawOutput:        string(output),
		}
	}
	return nil
}

func (m *MiseToolProvider) isAlreadyInstalled(
	toolName provider.ToolID,
	concreteVersion string,
) (bool, error) {
	return versionExistsLocal(m.ExecEnv, toolName, concreteVersion)
}

func miseVersionString(toolName provider.ToolID, concreteVersion string) string {
	return fmt.Sprintf("%s@%s", toolName, concreteVersion)
}
