package mise

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

var errNoMatchingVersion = errors.New("no matching version found")

func (m *MiseToolProvider) resolveToConcreteVersionAfterInstall(tool provider.ToolRequest) (string, error) {
	// Mise doesn't tell us what version it resolved to when installing the user-provided (and potentially fuzzy) version.
	// But we can use `mise latest` to find out the concrete version.
	switch tool.ResolutionStrategy {
	case provider.ResolutionStrategyLatestInstalled:
		return m.resolveToLatestInstalled(tool.ToolName, tool.UnparsedVersion)
	case provider.ResolutionStrategyLatestReleased, provider.ResolutionStrategyStrict:
		// Mise works with fuzzy versions by default, so it happily installs both node@20 and node@20.19.3.
		// Therefore, when the Bitrise config contains simply 20 (and not 20:latest), it actually behaves
		// as "latest released".
		return m.resolveToLatestReleased(tool.ToolName, tool.UnparsedVersion)
	default:
		return "", fmt.Errorf("unknown resolution strategy: %v", tool.ResolutionStrategy)
	}
}

func (m *MiseToolProvider) resolveToLatestReleased(toolName provider.ToolID, version string) (string, error) {
	return resolveToLatestReleased(m.ExecEnv, toolName, version)
}

func resolveToLatestReleased(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (string, error) {
	// Even if version is empty string "sometool@" will not cause an error.
	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "latest", fmt.Sprintf("%s@%s", toolName, version))
	if err != nil {
		return "", fmt.Errorf("mise latest %s@%s: %w", toolName, version, err)
	}

	v := strings.TrimSpace(string(output))
	if v == "" {
		return "", errNoMatchingVersion
	}

	return v, nil
}

func (m *MiseToolProvider) resolveToLatestInstalled(toolName provider.ToolID, version string) (string, error) {
	return resolveToLatestInstalled(m.ExecEnv, toolName, version)
}

func resolveToLatestInstalled(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (string, error) {
	// Even if version is empty string "sometool@" will not cause an error.
	var toolString = string(toolName)
	if version != "" && version != "installed" {
		// tool@installed is not valid, so only append version when it's not "installed"
		toolString = fmt.Sprintf("%s@%s", toolName, version)
	}

	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "latest", "--installed", "--quiet", toolString)
	if err != nil {
		return "", fmt.Errorf("mise latest --installed %s: %w", toolString, err)
	}

	v := strings.TrimSpace(string(output))
	if v == "" {
		return "", errNoMatchingVersion
	}

	return v, nil
}

// versionExists checks whether the given version of the tool is available.
// Note: this checks both local and remote availability, does not check strategy.
func (m *MiseToolProvider) versionExists(toolName provider.ToolID, version string) (bool, error) {
	return versionExists(m.ExecEnv, toolName, version)
}

// versionExistsLocal checks whether the given version of the tool is installed locally.
// Or if version is empty, checks whether at least one version is installed.
func versionExistsLocal(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (bool, error) {
	// List all installed versions to see if there is at least one version available (or a specific one).
	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls", "--installed", "--json", "--quiet", string(toolName))
	if err != nil {
		return false, fmt.Errorf("mise ls --installed %s: %w", toolName, err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		// Parse JSON array returned by mise and check if the provided version is installed (when version is not empty).
		installedExists, err := parseInstalledVersionsJSON(trimmed, version)
		if err != nil {
			return false, fmt.Errorf("parsing mise ls --installed %s output: %w", toolName, err)
		}

		if installedExists {
			return true, nil
		}
	}

	return false, nil
}

// versionExists checks whether the given version of the tool is available.
// Note: this checks both local and remote availability, does not check strategy.
func versionExists(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (bool, error) {
	// This and 'latest' keywords are not equal to strategy.
	if version == "installed" {
		existsLocally, err := versionExistsLocal(execEnv, toolName, "")
		if err != nil {
			return false, err
		}

		if existsLocally {
			return true, nil
		}

		// Fallback: no installed versions found, fall through to remote (ls-remote) existence check.
	}

	versionString := string(toolName)
	if version != "" && version != "latest" && version != "installed" {
		versionString = fmt.Sprintf("%s@%s", toolName, version)
	}

	// Notes:
	// - ls-remote accepts both fuzzy and concrete versions
	// - it can return multiple versions (one per line) when a fuzzy version is provided
	// - in case of no matching version, the exit code is still 0, just there is no output
	// - in case of a non-existing tool, the exit code is 1, but a non-existing tool ID fails earlier than this check
	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls-remote", "--quiet", versionString)
	if err != nil {
		return false, fmt.Errorf("mise ls-remote %s: %w", versionString, err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// parseInstalledVersionsJSON returns true if at least one installed entry is present.
func parseInstalledVersionsJSON(raw, requiredVersion string) (bool, error) {
	var entries []struct {
		Version   string `json:"version"`
		Installed bool   `json:"installed"`
	}
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.Installed {
			if requiredVersion == "" {
				return true, nil
			}
			if e.Version == requiredVersion {
				return true, nil
			}
		}
	}
	return false, nil
}
