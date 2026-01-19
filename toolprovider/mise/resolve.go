package mise

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

var errNoMatchingVersion = errors.New("no matching version found")

// extractLastLine extracts the last non-empty line from multi-line output.
// This is needed because mise may output plugin installation messages before the actual version.
func extractLastLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

func resolveToLatestReleased(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (string, error) {
	// Note: Even if version is an empty string, "sometool@" will not cause an error.
	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "latest", "--quiet", fmt.Sprintf("%s@%s", toolName, version))
	if err != nil {
		return "", fmt.Errorf("mise latest %s@%s: %w", toolName, version, err)
	}

	// Extract the last non-empty line, as mise may output plugin installation messages before the version
	v := extractLastLine(string(output))
	if v == "" {
		return "", errNoMatchingVersion
	}

	return v, nil
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

	// Extract the last non-empty line, as mise may output plugin installation messages before the version
	v := extractLastLine(string(output))
	if v == "" {
		return "", errNoMatchingVersion
	}

	return v, nil
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

// versionExistsRemote checks if a version exists in the remote registry.
// version can be fuzzy (e.g., "20") or concrete (e.g., "20.18.1")
func versionExistsRemote(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (bool, error) {
	versionString := string(toolName)
	if version != "" && version != "latest" {
		versionString = fmt.Sprintf("%s@%s", toolName, version)
	}

	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls-remote", "--quiet", versionString)
	if err != nil {
		return false, fmt.Errorf("mise ls-remote %s: %w", versionString, err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

func normalizeRequest(
	execEnv execenv.ExecEnv,
	request provider.ToolRequest,
) (provider.ToolRequest, error) {
	normalizedRequest := request
	// Handle "installed" and "latest" special keywords
	if request.UnparsedVersion == "installed" {
		normalizedRequest.UnparsedVersion = ""
		normalizedRequest.ResolutionStrategy = provider.ResolutionStrategyLatestInstalled
	}
	if request.UnparsedVersion == "latest" {
		normalizedRequest.UnparsedVersion = ""
		normalizedRequest.ResolutionStrategy = provider.ResolutionStrategyLatestReleased
	}

	// Latest installed: check if any installed versions exist, otherwise fallback to latest released
	if normalizedRequest.ResolutionStrategy == provider.ResolutionStrategyLatestInstalled {
		_, err := resolveToLatestInstalled(execEnv, normalizedRequest.ToolName, normalizedRequest.UnparsedVersion)
		if err == nil {
			// Installed version found, return as is
			return normalizedRequest, nil
		}

		if errors.Is(err, errNoMatchingVersion) {
			log.Infof("No installed versions found, fallback to latest released")
			return provider.ToolRequest{
				ToolName:           normalizedRequest.ToolName,
				UnparsedVersion:    normalizedRequest.UnparsedVersion,
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			}, nil
		}

		return normalizedRequest, err
	}
	return normalizedRequest, nil
}

// resolveToConcreteVersion resolves any version input to a concrete version string.
// Assumes strategy has already been normalized via normalizeRequest.
func resolveToConcreteVersion(
	execEnv execenv.ExecEnv,
	toolName provider.ToolID,
	version string,
	strategy provider.ResolutionStrategy,
) (string, error) {
	switch strategy {
	case provider.ResolutionStrategyStrict, provider.ResolutionStrategyLatestReleased:
		// Both strategies resolve fuzzy versions the same way.
		// The difference is in the fallback behavior (handled by normalizeRequest),
		// not in how versions are resolved.
		return resolveToLatestReleased(execEnv, toolName, version)
	case provider.ResolutionStrategyLatestInstalled:
		return resolveToLatestInstalled(execEnv, toolName, version)
	default:
		return "", fmt.Errorf("unknown resolution strategy: %v", strategy)
	}
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
