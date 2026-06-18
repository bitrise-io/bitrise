package toolprovider

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/versionsort"
)

// ListToolVersions resolves aliases, validates the tool name against
// SupportedTools, and returns all released versions sorted newest-first.
// If versionPrefix is non-empty, only versions starting with that prefix are returned.
func ListToolVersions(toolName string, versionPrefix string, tp provider.ToolProvider) ([]string, error) {
	canonicalName := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	if !slices.Contains(SupportedTools(), canonicalName) {
		return nil, fmt.Errorf("%q is not a supported tool. Supported tools: %v", toolName, SupportedTools())
	}

	versions, err := tp.ListReleasedVersions(provider.ToolID(canonicalName))
	if err != nil {
		return nil, fmt.Errorf("list versions for %s: %w", toolName, err)
	}

	versions = versionsort.SortSemverDescending(versions)

	if versionPrefix != "" {
		versionPrefix = strings.TrimRight(versionPrefix, ".")
		var filtered []string
		for _, v := range versions {
			if v == versionPrefix || strings.HasPrefix(v, versionPrefix+".") {
				filtered = append(filtered, v)
			}
		}
		versions = filtered
	}

	return versions, nil
}
