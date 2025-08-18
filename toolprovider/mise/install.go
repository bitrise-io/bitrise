package mise

import (
	"errors"
	"fmt"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func (m *MiseToolProvider) installToolVersion(tool provider.ToolRequest) error {
	versionString, err := miseVersionString(tool, m.resolveToLatestInstalled)
	if err != nil {
		return err
	}

	output, err := m.ExecEnv.RunMise("install", "--yes", versionString)
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
// Inputs: tool ID, tool version
// Returns: latest installed version of the tool, or an error if no matching version is installed
type latestInstalledResolver func(provider.ToolID, string) (string, error)

func isAlreadyInstalled(tool provider.ToolRequest, latestInstalledResolver latestInstalledResolver) (bool, error) {
	_, err := latestInstalledResolver(tool.ToolName, tool.UnparsedVersion)
	var isAlreadyInstalled bool
	if err != nil {
		if errors.Is(err, errNoMatchingVersion) {
			isAlreadyInstalled = false
		} else {
			return false, err
		}
	} else {
		isAlreadyInstalled = true
	}
	return isAlreadyInstalled, nil
}

func miseVersionString(tool provider.ToolRequest, latestInstalledResolver latestInstalledResolver) (string, error) {
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
				// No local version satisfies the request -> fallback to latest released
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
