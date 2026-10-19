package toolprovider

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v3/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
)

// ListToolVersions resolves aliases, validates the tool name against
// SupportedTools, and returns the provider's released versions newest first.
func ListToolVersions(toolName string, versionPrefix string, tp provider.ToolProvider) ([]string, error) {
	canonicalName := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	if !IsSupported(toolName) {
		return nil, fmt.Errorf("%q is not a supported tool. Supported tools: %v", toolName, canonicalToolNames())
	}

	versions, err := tp.ListReleasedVersions(provider.ToolID(canonicalName), versionPrefix)
	if err != nil {
		return nil, fmt.Errorf("list versions for %s: %w", toolName, err)
	}

	return versions, nil
}
