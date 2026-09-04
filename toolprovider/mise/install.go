package mise

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/nixpkgs"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/workarounds"
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

	extraEnvs := make(map[string]string)
	if key, value := workarounds.GetPythonPrecompiledFlavorEnv(toolName, concreteVersion, GetMiseVersion(), m.Silent); key != "" {
		extraEnvs[key] = value
	}

	output, err := m.ExecEnv.RunMiseWithTimeoutAndEnvs(execenv.InstallTimeout, extraEnvs, "install", "--yes", versionString)
	if !m.Silent && os.Getenv("MISE_LOG_LEVEL") != "" {
		// Log the output of successful installs as well, useful for debugging and tests.
		log.Printf("[TOOLPROVIDER] mise install output for %s:\n%s", versionString, output)
	}
	if err != nil {
		cause := fmt.Sprintf("mise install %s: %s", versionString, err)
		rawOutput := string(output)
		var recommendation string

		if strings.Contains(strings.ToLower(err.Error()), "rate limit exceeded") {
			// The mise output (which is what makes this recognizable as a rate limit error) is embedded in err,
			// not in output: RunMiseWithTimeoutAndEnvs() only returns command output separately on success.
			cause = fmt.Sprintf("GitHub API rate limit exceeded while installing %s", toolName)
			rawOutput = err.Error()
			recommendation = fmt.Sprintf(
				"GitHub's public API allows only a small number of requests per hour without a token. "+
					"Add one of %s as a secret to raise the limit and avoid this error. "+
					"See the raw output below for details.",
				strings.Join(KnownGitHubTokenEnvVars, ", "),
			)
		}

		return provider.ToolInstallError{
			ToolName:         toolName,
			RequestedVersion: concreteVersion,
			Cause:            cause,
			RawOutput:        rawOutput,
			Recommendation:   recommendation,
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
