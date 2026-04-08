package versionresolver

import (
	"fmt"
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
	var best *candidate

	for _, raw := range availableVersions {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}

		if !constraint.Check(v) {
			continue
		}

		if best == nil || v.GreaterThan(best.ver) {
			best = &candidate{raw: raw, ver: v}
		}
	}

	if best == nil {
		return "", fmt.Errorf("no version matching constraint %q", constraintRaw)
	}

	return best.raw, nil
}
