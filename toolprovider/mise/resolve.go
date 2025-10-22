package mise

import (
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
	// Even if version is empty string "sometool@" will not cause an error.
	output, err := m.ExecEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "latest", fmt.Sprintf("%s@%s", toolName, version))
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
	// Even if version is empty string "sometool@" will not cause an error.
	output, err := m.ExecEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "latest", "--installed", "--quiet", fmt.Sprintf("%s@%s", toolName, version))
	if err != nil {
		return "", fmt.Errorf("mise latest --installed %s@%s: %w", toolName, version, err)
	}

	v := strings.TrimSpace(string(output))
	if v == "" {
		return "", errNoMatchingVersion
	}

	return v, nil
}

func (m *MiseToolProvider) versionExists(toolName provider.ToolID, version string) (bool, error) {
	// Notes:
	// - ls-remote accepts both fuzzy and concrete versions
	// - it can return multiple versions (one per line) when a fuzzy version is provided
	// - in case of no matching version, the exit code is still 0, just there is no output
	// - in case of a non-existing tool, the exit code is 1, but a non-existing tool ID fails earlier than this check

	if version == "installed" {
		// List all installed versions to see if there is at least one version available.
		output, err := m.ExecEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls", "--installed", "--quiet", string(toolName))
		if err != nil {
			return false, fmt.Errorf("mise ls --installed %s: %w", toolName, err)
		}

		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			return false, nil
		}

		// Mise outputs installed versions line by line, first is header (in some cases).
		lines := strings.Split(trimmed, "\n")
		return len(lines) > 1 || (len(lines) == 1 && !strings.HasPrefix(lines[0], "Tool")), nil
	}

	search := string(toolName)
	if version != "" && version != "latest" {
		search = fmt.Sprintf("%s@%s", toolName, version)
	}

	output, err := m.ExecEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls-remote", "--quiet", search)
	if err != nil {
		return false, fmt.Errorf("mise ls-remote %s: %w", search, err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}
