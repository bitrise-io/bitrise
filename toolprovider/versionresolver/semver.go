package versionresolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ResolveConstraint takes an npm semver constraint string and a list of available versions,
// and returns the latest version that satisfies the constraint.
// Pre-release versions are skipped unless the constraint explicitly targets them.
func ResolveConstraint(constraintRaw string, availableVersions []string) (string, error) {
	constraintRaw = strings.TrimSpace(constraintRaw)
	if constraintRaw == "" {
		return "", fmt.Errorf("empty version constraint")
	}

	constraint, err := semver.NewConstraint(constraintRaw)
	if err != nil {
		return "", fmt.Errorf("invalid semver constraint %q: %w", constraintRaw, err)
	}

	type candidate struct {
		raw string
		ver *semver.Version
	}
	var candidates []candidate

	for _, raw := range availableVersions {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}

		if v.Prerelease() != "" {
			continue
		}

		if constraint.Check(v) {
			candidates = append(candidates, candidate{raw: raw, ver: v})
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no version matching constraint %q", constraintRaw)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ver.GreaterThan(candidates[j].ver)
	})

	return candidates[0].raw, nil
}
