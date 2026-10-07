// Package update discovers the published Bitrise CLI releases, decides whether
// the running build may be told about them, and keeps the result of the last
// check in its own state file.
package update

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// maxReleasePages bounds the paging, so a current major with no published
// release cannot walk the whole release history.
const maxReleasePages = 5

type Available struct {
	Version string `yaml:"version"`
	URL     string `yaml:"url"`
}

// NewMajor is kept apart from Update because a major version can break the user,
// so it is reported and never installed implicitly.
type Versions struct {
	Update   *Available `yaml:"update,omitempty"`
	NewMajor *Available `yaml:"new_major,omitempty"`
	// CurrentMajorFound is false when the pages read held no release of the
	// running major, which a nil Update cannot be told apart from on its own. The
	// caller must not report "up to date" in that case.
	CurrentMajorFound bool `yaml:"current_major_found"`
}

type usableRelease struct {
	version *semver.Version
	url     string
}

// Resolve returns the newest release inside currentVersion's major, and
// separately the newest release above it. Drafts, pre-releases and tags that are
// not a clean MAJOR.MINOR.PATCH are never offered.
func Resolve(ctx context.Context, client *Client, currentVersion string) (Versions, error) {
	current, err := parseVersion(currentVersion)
	if err != nil {
		return Versions{}, fmt.Errorf("current version is not comparable: %w", err)
	}

	// The newest release of each major is on page 1 by construction, so a further
	// page is only worth asking for while nothing of the current major has shown up.
	var candidates []usableRelease
	for page := 1; page <= maxReleasePages; page++ {
		releases, err := client.ReleasesPage(ctx, page)
		if err != nil {
			// The pages already read go with it: cached as a successful check, a
			// partial result would hide a release until the next check is due.
			return Versions{}, err
		}
		candidates = append(candidates, usableReleases(releases)...)
		if len(releases) < releasesPerPage || hasMajor(candidates, current.Major()) {
			break
		}
	}

	versions := selectVersions(candidates, current)
	versions.CurrentMajorFound = hasMajor(candidates, current.Major())
	return versions, nil
}

func usableReleases(releases []Release) []usableRelease {
	var usable []usableRelease
	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		v, err := parseVersion(strings.TrimPrefix(r.TagName, "v"))
		if err != nil {
			continue
		}
		usable = append(usable, usableRelease{version: v, url: r.HTMLURL})
	}
	return usable
}

func hasMajor(candidates []usableRelease, major uint64) bool {
	return slices.ContainsFunc(candidates, func(c usableRelease) bool { return c.version.Major() == major })
}

func selectVersions(candidates []usableRelease, current *semver.Version) Versions {
	var update, newMajor *usableRelease
	for i, c := range candidates {
		switch {
		case c.version.Major() == current.Major():
			if c.version.GreaterThan(current) && (update == nil || c.version.GreaterThan(update.version)) {
				update = &candidates[i]
			}
		case c.version.Major() > current.Major():
			if newMajor == nil || c.version.GreaterThan(newMajor.version) {
				newMajor = &candidates[i]
			}
		}
	}
	return Versions{Update: update.available(), NewMajor: newMajor.available()}
}

func (r *usableRelease) available() *Available {
	if r == nil {
		return nil
	}
	return &Available{Version: r.version.String(), URL: r.url}
}
