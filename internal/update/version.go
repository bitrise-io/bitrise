package update

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func isComparable(version string) bool {
	_, err := parseVersion(version)
	return err == nil
}

// Only a clean MAJOR.MINOR.PATCH corresponds to a published release: a plain
// `go build` reports "dev" and a goreleaser snapshot "<next patch>-next".
func parseVersion(version string) (*semver.Version, error) {
	v, err := semver.StrictNewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("parse version %q: %w", version, err)
	}
	if v.Prerelease() != "" || v.Metadata() != "" {
		return nil, fmt.Errorf("parse version %q: not a MAJOR.MINOR.PATCH version", version)
	}
	return v, nil
}

// The `v` prefix is accepted because that is how the releases are tagged, and a
// pre-release is kept, since naming one is the only way to install it.
func parseRequestedVersion(raw string) (*semver.Version, error) {
	v, err := semver.StrictNewVersion(strings.TrimPrefix(raw, "v"))
	if err != nil || v.Metadata() != "" {
		return nil, fmt.Errorf("invalid version %q: expected MAJOR.MINOR.PATCH, for example 2.46.0", raw)
	}
	return v, nil
}
