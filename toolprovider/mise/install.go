package mise

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

// shouldSetPythonPrecompiledFlavor determines if MISE_PYTHON_PRECOMPILED_FLAVOR should be set.
// This is needed for Python 3.14+ with mise versions before 2026 to avoid missing lib directory errors.
// The fix was implemented in mise v2026.3.10.
func shouldSetPythonPrecompiledFlavor(toolName provider.ToolID, concreteVersion string, miseVersion string) bool {
	toolNameStr := string(toolName)
	if !strings.HasSuffix(toolNameStr, "python") {
		return false
	}

	if strings.HasPrefix(miseVersion, "v2026") {
		return false
	}

	parts := strings.Split(concreteVersion, ".")
	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	// Extract minor version (may have letters like "14a1")
	minorStr := parts[1]
	var minor int
	for i, c := range minorStr {
		if c < '0' || c > '9' {
			minorStr = minorStr[:i]
			break
		}
	}
	minor, err = strconv.Atoi(minorStr)
	if err != nil {
		return false
	}

	if major > 3 || (major == 3 && minor >= 14) {
		return true
	}

	return false
}

func (m *MiseToolProvider) installToolVersion(toolName provider.ToolID, concreteVersion string) error {
	versionString := miseVersionString(toolName, concreteVersion)

	// Conditionally set MISE_PYTHON_PRECOMPILED_FLAVOR for Python 3.14+ with mise versions before 2026
	// to avoid missing lib directory errors. The fix was implemented in mise v2026.3.10.
	// https://mise.jdx.dev/lang/python.html#python.precompiled_flavor
	// https://github.com/jdx/mise/releases/tag/v2026.3.10
	extraEnvs := make(map[string]string)
	if shouldSetPythonPrecompiledFlavor(toolName, concreteVersion, GetMiseVersion()) {
		extraEnvs["MISE_PYTHON_PRECOMPILED_FLAVOR"] = "install_only_stripped"
		if !m.Silent {
			log.Debugf("[TOOLPROVIDER] Setting MISE_PYTHON_PRECOMPILED_FLAVOR for Python %s", concreteVersion)
		}
	}

	output, err := m.ExecEnv.RunMiseWithTimeoutAndEnvs(execenv.InstallTimeout, extraEnvs, "install", "--yes", versionString)
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
