package versionsort

import (
	"sort"

	"github.com/hashicorp/go-version"
)

// SortSemverDescending returns a sorted copy of versions, newest first, using semver where possible.
// Non-semver versions are placed after semver versions in reverse lexicographic order.
func SortSemverDescending(versions []string) []string {
	var semverVersions version.Collection
	var nonSemverVersions []string
	for _, v := range versions {
		semverV, err := version.NewVersion(v)
		if err != nil {
			nonSemverVersions = append(nonSemverVersions, v)
			continue
		}
		semverVersions = append(semverVersions, semverV)
	}

	// semverVersions is of type version.Collection, which implements sort.Interface according to the semver spec.
	sort.Sort(sort.Reverse(semverVersions))
	// nonSemverVersions are only lexicographically sortable
	sort.Sort(sort.Reverse(sort.StringSlice(nonSemverVersions)))

	var sortedVersions []string
	for _, v := range semverVersions {
		sortedVersions = append(sortedVersions, v.Original())
	}

	sortedVersions = append(sortedVersions, nonSemverVersions...)
	return sortedVersions
}
