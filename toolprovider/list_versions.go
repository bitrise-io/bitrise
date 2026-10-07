package toolprovider

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/bitrise-io/bitrise/v3/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v3/toolprovider/versionsort"
)

// ListToolVersions resolves aliases, validates the tool name against
// SupportedTools, and returns all released versions newest first, in the order
// the provider resolves them.
// If versionPrefix is set, only versions in that prefix's line are returned.
//
// mise lists oldest first and resolves a prefix to the last stable match in that
// list, so its list is reversed rather than sorted: semver puts 1.20.4 above
// 1.20.4-otp-29, and text order puts jruby-9 above jruby-10. Prereleases stay in,
// consumers skip them as mise does. The workflow editor relies on this order for
// its hints. asdf resolves by semver, so its list is sorted that way.
func ListToolVersions(toolName string, versionPrefix string, tp provider.ToolProvider) ([]string, error) {
	canonicalName := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	if !IsSupported(toolName) {
		return nil, fmt.Errorf("%q is not a supported tool. Supported tools: %v", toolName, canonicalToolNames())
	}

	versions, err := tp.ListReleasedVersions(provider.ToolID(canonicalName))
	if err != nil {
		return nil, fmt.Errorf("list versions for %s: %w", toolName, err)
	}

	if tp.ID() == "asdf" {
		versions = versionsort.SortSemverDescending(versions)
	} else {
		slices.Reverse(versions)
	}

	versionPrefix = strings.TrimRight(versionPrefix, ".")
	if versionPrefix != "" {
		var filtered []string
		for _, v := range versions {
			if isInVersionLine(v, versionPrefix) {
				filtered = append(filtered, v)
			}
		}
		versions = filtered
	}

	return versions, nil
}

// isInVersionLine mirrors how mise matches a prefix: 24.2 covers 24.2.0, not 24.20.0,
// and 1.19.0 covers 1.19.0-otp-28. A plus continues only a numeric line, as in
// 1.7.8+hotfix.4-stable, because after a name it joins flavours, as in truffleruby+graalvm.
func isInVersionLine(version, prefix string) bool {
	if version == prefix {
		return true
	}

	rest, ok := strings.CutPrefix(version, prefix)
	// mise wants something after the separator, so 22. is not in the 22 line.
	if !ok || len(rest) < 2 {
		return false
	}

	separators := ".-"
	if unicode.IsDigit(rune(prefix[0])) {
		separators += "+"
	}
	return strings.ContainsRune(separators, rune(rest[0]))
}
