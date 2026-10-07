package toolprovider

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v3/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
)

// ListToolVersions resolves aliases, validates the tool name against
// SupportedTools, and returns all released versions newest first.
// If versionPrefix is non-empty, only versions starting with that prefix are returned.
//
// mise lists oldest first and resolves a prefix to the last match in that list, so
// reversing it keeps the first match equal to what mise installs.
// Sorting would not: semver puts 1.20.4 above 1.20.4-otp-29, and text order
// puts jruby-9 above jruby-10. The workflow editor relies on this for its hints.
func ListToolVersions(toolName string, versionPrefix string, tp provider.ToolProvider) ([]string, error) {
	canonicalName := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	if !IsSupported(toolName) {
		return nil, fmt.Errorf("%q is not a supported tool. Supported tools: %v", toolName, canonicalToolNames())
	}

	versions, err := tp.ListReleasedVersions(provider.ToolID(canonicalName))
	if err != nil {
		return nil, fmt.Errorf("list versions for %s: %w", toolName, err)
	}

	slices.Reverse(versions)

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
