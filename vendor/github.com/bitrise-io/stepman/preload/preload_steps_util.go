package preload

import (
	"fmt"
	"time"

	"github.com/bitrise-io/stepman/models"
)

func filterPreloadedStepVersions(stepID string, steps map[string]models.StepModel, opts CacheOpts) (map[string]models.StepModel, error) {
	filteredSteps := map[string]models.StepModel{}
	allMajorMinor := map[uint64]map[uint64]models.Semver{}
	allLatestNMinor := map[uint64][]uint64{}

	// keep no version, as it is only used internally and takes up the most space
	if stepID == "project-scanner" {
		return filteredSteps, nil
	}

	for stepVersion, step := range steps {
		if step.PublishedAt == nil {
			return filteredSteps, fmt.Errorf("step %s@%s has no published_at date", stepID, stepVersion)
		}

		// Include all patch version releases
		publishDate := *(step.PublishedAt)
		if time.Since(publishDate.AddDate(0, opts.PatchesSinceMonths, 0)) > 0 {
			filteredSteps[stepVersion] = step
		}

		// All minor versions
		version, err := models.ParseSemver(stepVersion)
		if err != nil {
			return filteredSteps, fmt.Errorf("failed to parse version %s: %w", stepVersion, err)
		}

		if _, found := allMajorMinor[version.Major]; !found {
			allMajorMinor[version.Major] = map[uint64]models.Semver{}
			allMajorMinor[version.Major][version.Minor] = version

			continue
		}

		curVersion, found := allMajorMinor[version.Major][version.Minor]
		if !found {
			allMajorMinor[version.Major][version.Minor] = version

			continue
		} else if version.Patch > curVersion.Patch {
			allMajorMinor[version.Major][version.Minor] = version
		}
	}

	latestNMajors := make([]uint64, 0, opts.NumMajor)
	for major, minorToVersion := range allMajorMinor {
		latestNMajors = insertLatestNVersions(latestNMajors, major)
		latestNMinors := make([]uint64, 0, opts.NumMinor)
		for minor, version := range minorToVersion {
			latestNMinors = insertLatestNVersions(latestNMinors, minor)

			// The latest patch of any minor version
			publishDate := *steps[version.String()].PublishedAt
			if time.Since(publishDate.AddDate(0, opts.LatestMinorsSinceMonths, 0)) > 0 {
				filteredSteps[version.String()] = steps[version.String()]
			}
		}

		allLatestNMinor[major] = latestNMinors
	}

	for _, major := range latestNMajors {
		for _, minor := range allLatestNMinor[major] {
			latestPatch := allMajorMinor[major][minor]
			filteredSteps[latestPatch.String()] = steps[latestPatch.String()]
		}
	}

	return filteredSteps, nil
}

func insertLatestNVersions(latests []uint64, newVersion uint64) []uint64 {
	if len(latests) > 1 {
		for i, cur := range latests {
			if newVersion > cur {
				if len(latests) < cap(latests) {
					latests = append(latests[:i], append([]uint64{newVersion}, latests[i:]...)...)
					return latests
				}

				latests = append(latests[:i], append([]uint64{newVersion}, latests[i:len(latests)-1]...)...)
				return latests
			}
		}
	}

	if len(latests) == 1 {
		if newVersion > latests[0] {
			if len(latests) < cap(latests) {
				latests = append([]uint64{newVersion}, latests...)
				return latests
			}
			latests[0] = newVersion
			return latests
		}
	}

	if len(latests) < cap(latests) {
		latests = append(latests, newVersion)
	}

	return latests
}
