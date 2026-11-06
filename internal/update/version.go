package update

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

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
