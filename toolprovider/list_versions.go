package toolprovider

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// ListToolVersions resolves aliases, validates the tool name against
// SupportedTools, and returns all released versions sorted newest-first.
func ListToolVersions(toolName string, tp provider.ToolProvider) ([]string, error) {
	canonicalName := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	if !slices.Contains(SupportedTools(), canonicalName) {
		return nil, fmt.Errorf("%q is not a supported tool. Supported tools: %v", toolName, SupportedTools())
	}

	versions, err := tp.ListReleasedVersions(provider.ToolID(canonicalName))
	if err != nil {
		return nil, fmt.Errorf("list versions for %s: %w", toolName, err)
	}

	sortVersionsDescending(versions)

	return versions, nil
}

// sortVersionsDescending sorts versions newest-first using semver where possible.
// Non-semver versions are placed after semver versions in their original order.
func sortVersionsDescending(versions []string) {
	type entry struct {
		raw string
		ver *semver.Version
		idx int
	}

	entries := make([]entry, len(versions))
	for i, raw := range versions {
		v, _ := semver.NewVersion(raw)
		entries[i] = entry{raw: raw, ver: v, idx: i}
	}

	slices.SortStableFunc(entries, func(a, b entry) int {
		if a.ver != nil && b.ver != nil {
			return b.ver.Compare(a.ver) // descending
		}
		if a.ver != nil {
			return -1 // semver before non-semver
		}
		if b.ver != nil {
			return 1
		}
		return cmp.Compare(a.idx, b.idx) // preserve original order for non-semver
	})

	for i, e := range entries {
		versions[i] = e.raw
	}
}
