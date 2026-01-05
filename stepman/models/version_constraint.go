package models

import (
	"fmt"
	"strconv"
	"strings"
)

// Semver represents a semantic version
type Semver struct {
	Major, Minor, Patch uint64
}

// String converts a Semver to string
func (v *Semver) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func ParseSemver(version string) (Semver, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		return Semver{}, fmt.Errorf("parse %s: should consist by 3 components", version)
	}

	major, err := strconv.ParseUint(versionParts[0], 10, 0)
	if err != nil {
		return Semver{}, fmt.Errorf("parse %s: invalid major version: %s", version, err)
	}
	minor, err := strconv.ParseUint(versionParts[1], 10, 0)
	if err != nil {
		return Semver{}, fmt.Errorf("parse %s: invalid minor version: %s", version, err)
	}
	patch, err := strconv.ParseUint(versionParts[2], 10, 0)
	if err != nil {
		return Semver{}, fmt.Errorf("parse %s: invalid patch version: %s", version, err)
	}

	return Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func CmpSemver(a, b Semver) int {
	if a.Major < b.Major {
		return -1
	}
	if a.Major > b.Major {
		return 1
	}

	if a.Minor < b.Minor {
		return -1
	}
	if a.Minor > b.Minor {
		return 1
	}

	if a.Patch < b.Patch {
		return -1
	}
	if a.Patch > b.Patch {
		return 1
	}

	return 0
}

// VersionLockType is the type of version lock
type VersionLockType int

const (
	// InvalidVersionConstraint is the value assigned to a VersionLockType if not explicitly initialized
	InvalidVersionConstraint VersionLockType = iota
	// Fixed is an exact version, e.g. 1.2.5
	Fixed
	// Latest means the latest available version
	Latest
	// MajorLocked means the latest available version with a given major version, e.g. 1.*.*
	MajorLocked
	// MinorLocked means the latest available version with a given major and minor version, e.g. 1.2.*
	MinorLocked
)

// VersionConstraint describes a version and a cosntraint (e.g. use latest major version available)
type VersionConstraint struct {
	VersionLockType VersionLockType
	Version         Semver
}

// ParseRequiredVersion returns VersionConstraint model from raw version string
func ParseRequiredVersion(version string) (VersionConstraint, error) {
	if version == "" {
		return VersionConstraint{
			VersionLockType: Latest,
			Version:         Semver{},
		}, nil
	}

	parts := strings.Split(version, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return VersionConstraint{}, fmt.Errorf("parse %s: should have more than 0 and not more than 3 components", version)
	}

	major, err := strconv.ParseUint(parts[0], 10, 0)
	if err != nil {
		return VersionConstraint{}, fmt.Errorf("parse %s: invalid major version: %s", version, err)
	}

	if len(parts) == 1 ||
		(len(parts) == 3 &&
			parts[1] == "x" && parts[2] == "x") {
		return VersionConstraint{
			VersionLockType: MajorLocked,
			Version: Semver{
				Major: major,
				Minor: 0,
				Patch: 0,
			},
		}, nil
	}

	minor, err := strconv.ParseUint(parts[1], 10, 0)
	if err != nil {
		return VersionConstraint{}, fmt.Errorf("parse %s: invalid minor version: %s", version, err)
	}

	if len(parts) == 2 ||
		(len(parts) == 3 && parts[2] == "x") {
		return VersionConstraint{
			VersionLockType: MinorLocked,
			Version: Semver{
				Major: major,
				Minor: minor,
				Patch: 0,
			},
		}, nil
	}

	patch, err := strconv.ParseUint(parts[2], 10, 0)
	if err != nil {
		return VersionConstraint{}, fmt.Errorf("parse %s: invalid patch version: %s", version, err)
	}

	return VersionConstraint{
		VersionLockType: Fixed,
		Version: Semver{
			Major: major,
			Minor: minor,
			Patch: patch,
		},
	}, nil
}

func latestMatchingStepVersion(constraint VersionConstraint, stepVersions StepGroupModel) (StepVersionModel, bool) {
	switch constraint.VersionLockType {
	case Fixed:
		{
			version := constraint.Version.String()
			latestStep, versionFound := stepVersions.Versions[version]

			if !versionFound {
				return StepVersionModel{}, false
			}

			return StepVersionModel{
				Step:                   latestStep,
				Version:                version,
				LatestAvailableVersion: stepVersions.LatestVersionNumber,
			}, true
		}
	case MinorLocked:
		{
			latestVersion := Semver{
				Major: constraint.Version.Major,
				Minor: constraint.Version.Minor,
				Patch: 0,
			}
			latestStep := StepModel{}

			for fullVersion, step := range stepVersions.Versions {
				stepVersion, err := ParseSemver(fullVersion)
				if err != nil {
					continue
				}
				if stepVersion.Major != constraint.Version.Major ||
					stepVersion.Minor != constraint.Version.Minor {
					continue
				}

				if stepVersion.Patch > latestVersion.Patch {
					latestVersion = stepVersion
					latestStep = step
				}
			}

			return StepVersionModel{
				Step:                   latestStep,
				Version:                latestVersion.String(),
				LatestAvailableVersion: stepVersions.LatestVersionNumber,
			}, true
		}
	case MajorLocked:
		{
			latestStepVersion := Semver{
				Major: constraint.Version.Major,
				Minor: 0,
				Patch: 0,
			}
			latestStep := StepModel{}

			for fullVersion, step := range stepVersions.Versions {
				stepVersion, err := ParseSemver(fullVersion)
				if err != nil {
					continue
				}
				if stepVersion.Major != constraint.Version.Major {
					continue
				}

				if stepVersion.Minor > latestStepVersion.Minor ||
					(stepVersion.Minor == latestStepVersion.Minor && stepVersion.Patch > latestStepVersion.Patch) {
					latestStepVersion = stepVersion
					latestStep = step
				}
			}

			return StepVersionModel{
				Step:                   latestStep,
				Version:                latestStepVersion.String(),
				LatestAvailableVersion: stepVersions.LatestVersionNumber,
			}, true
		}
	}

	return StepVersionModel{}, false
}
