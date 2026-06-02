package steplibrary

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/spec"
)

func (s *Steplib) getStepVersionInfo(ctx context.Context, stepID, version string) (models.StepInfoModel, ResolvedStepVersion, error) {
	var err error
	if stepID == "" {
		err = errors.New("missing required input: step id")
	}

	var allSteps []string
	if err == nil {
		allSteps, err = s.api.GetAllStepIDs(ctx)
		if err != nil {
			err = fmt.Errorf("fetching avaialble step IDs: %w", err)
		}
	}
	if err == nil && !slices.Contains(allSteps, stepID) {
		err = fmt.Errorf("%s steplib does not contain %s step", s.steplibURI, stepID)
	}

	var versionConstraint models.VersionConstraint
	if err == nil {
		versionConstraint, err = models.ParseRequiredVersion(version)
		if err != nil {
			err = fmt.Errorf("invalid step version constraint: %w", err)
		}
	}
	if err == nil && versionConstraint.VersionLockType == models.InvalidVersionConstraint {
		err = fmt.Errorf("invalid step version constraint: %s", version)
	}

	var latestVersions spec.LatestPointer
	if err == nil {
		latestVersions, err = s.api.GetLatestStepVersions(ctx, stepID)
		if err != nil {
			err = fmt.Errorf("fetching latest versions of `%s`: %w", stepID, err)
		}
	}

	var groupInfo spec.StepInfo
	if err == nil {
		groupInfo, err = s.api.GetStepGroupInfo(ctx, stepID)
		if err != nil {
			err = fmt.Errorf("fetching group info of `%s`: %w", stepID, err)
		}
	}

	var resolvedVersion string
	if err == nil {
		switch versionConstraint.VersionLockType {
		case models.Latest:
			resolvedVersion = latestVersions.Latest
		case models.Fixed:
			resolvedVersion = versionConstraint.Version.String()
			// ToDo: check version exists, otherwise error:
			// "%s steplib does not contain %s step %s version"
		case models.MajorLocked:
			majorKey := strconv.FormatUint(versionConstraint.Version.Major, 10)
			v, ok := latestVersions.LatestByMajor[majorKey]
			if !ok {
				err = fmt.Errorf("%s steplib does not contain %s step with major version %s", s.steplibURI, stepID, majorKey)
			} else {
				resolvedVersion = v
			}
		case models.MinorLocked:
			var allVersions []string
			allVersions, err = s.api.GetAllStepVersions(ctx, stepID)
			if err != nil {
				err = fmt.Errorf("fetching all versions of `%s`: %w", stepID, err)
			}
			if err == nil {
				resolvedVersion, err = resolveMinorLocked(allVersions, versionConstraint.Version)
				if err != nil {
					err = fmt.Errorf("%s steplib: %w", s.steplibURI, err)
				}
			}
		default:
			err = fmt.Errorf("unknown version constraint: %s", version)
		}
	}

	if err != nil {
		return models.StepInfoModel{}, ResolvedStepVersion{}, err
	}
	//nolint:exhaustruct // Step and DefinitionPth aren't surfaced by the v2 API yet
	return models.StepInfoModel{
		Library:         s.steplibURI,
		ID:              stepID,
		Version:         resolvedVersion,
		OriginalVersion: version,
		LatestVersion:   latestVersions.Latest,
		GroupInfo:       toStepGroupInfoModel(groupInfo),
	}, ResolvedStepVersion{ID: stepID, Version: resolvedVersion}, nil
}

// toStepGroupInfoModel flattens v2's nested `deprecation` object into v1's
// `RemovalDate` + `DeprecateNotes` fields so the rest of the codebase keeps
// reading the same model shape.
func toStepGroupInfoModel(info spec.StepInfo) models.StepGroupInfoModel {
	out := models.StepGroupInfoModel{
		Maintainer:     info.Maintainer,
		AssetURLs:      info.AssetURLs,
		RemovalDate:    "",
		DeprecateNotes: "",
	}
	if info.Deprecation != nil {
		out.RemovalDate = info.Deprecation.RemovalDate
		out.DeprecateNotes = info.Deprecation.Notes
	}
	return out
}

// resolveMinorLocked picks the highest patch within `versions` matching the
// constraint's Major+Minor. Unparseable entries are skipped.
func resolveMinorLocked(versions []string, constraint models.Semver) (string, error) {
	var best models.Semver
	found := false
	for _, raw := range versions {
		sv, err := models.ParseSemver(raw)
		if err != nil {
			continue
		}
		if sv.Major != constraint.Major || sv.Minor != constraint.Minor {
			continue
		}
		if !found || sv.Patch > best.Patch {
			best = sv
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("no version matches %d.%d.x", constraint.Major, constraint.Minor)
	}
	return best.String(), nil
}
