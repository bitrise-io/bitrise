package cli

import (
	"fmt"
	"time"

	"github.com/bitrise-io/stepman/models"
)

const (
	numMajors           = 2
	numMinors           = 1
	includeMinorsSince  = 24 * time.Hour * 31 * 0
	includePatchesSince = 24 * time.Hour * 31 * 0
	// includeMinorsSince  = 24 * time.Hour * 31 * 12 // 12 months
	// includePatchesSince = 24 * time.Hour * 31 * 6  // 6 months
)

func filterPreloadedStepVersions(stepID string, steps map[string]models.StepModel) (map[string]models.StepModel, error) {
	filteredSteps := map[string]models.StepModel{}
	allMajorMinor := map[uint64]map[uint64]models.Semver{}
	allLatestNMinor := map[uint64][]uint64{}

	// keep no version, as it is only used internally and takes up the most space
	if stepID == "project-scanner" {
		return filteredSteps, nil
	}

	for stepVersion, step := range steps {
		// Include all patch version releases
		if time.Since(*step.PublishedAt) < includePatchesSince {
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

	latestNMajors := make([]uint64, 0, numMajors)
	for major, minorToVersion := range allMajorMinor {
		latestNMajors = insertLatestNVersions(latestNMajors, major)
		latestNMinors := make([]uint64, 0, numMinors)
		for minor, version := range minorToVersion {
			latestNMinors = insertLatestNVersions(latestNMinors, minor)

			// The latest patch of any minor version
			if time.Since(*steps[version.String()].PublishedAt) < includeMinorsSince {
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

	if len(latests) < cap(latests) {
		latests = append(latests, newVersion)
	}

	return latests
}
