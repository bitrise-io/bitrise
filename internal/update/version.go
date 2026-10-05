package update

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// isComparable reports whether a running CLI version can be compared against
// the published releases. Only a clean MAJOR.MINOR.PATCH qualifies: a plain
// `go build` reports "dev" and a goreleaser snapshot reports
// "<next patch>-next", and neither of those builds corresponds to a release.
func isComparable(version string) bool {
	_, err := parseVersion(version)
	return err == nil
}

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
