package steplibrary

import (
	"context"
	"errors"
	"fmt"
	"path"
	"slices"
	"strconv"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

func (s *Steplib) getStepVersionInfo(ctx context.Context, stepID, version string) (models.StepInfoModel, ResolvedStepVersion, error) {
	if stepID == "" {
		return models.StepInfoModel{}, ResolvedStepVersion{}, errors.New("missing required input: step id")
	}

	versionConstraint, err := models.ParseRequiredVersion(version)
	if err != nil {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("invalid step version constraint: %w", err)
	}
	if versionConstraint.VersionLockType == models.InvalidVersionConstraint {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("invalid step version constraint: %s", version)
	}

	allSteps, err := s.api.GetAllStepIDs(ctx)
	if err != nil {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("fetching available step IDs: %w", err)
	}
	if !slices.Contains(allSteps, stepID) {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("%s steplib does not contain %s step", s.steplibURI, stepID)
	}

	latestVersions, err := s.api.GetLatestStepVersions(ctx, stepID)
	if err != nil {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("fetching latest versions of `%s`: %w", stepID, err)
	}

	groupInfo, err := s.api.GetStepGroupInfo(ctx, stepID)
	if err != nil {
		return models.StepInfoModel{}, ResolvedStepVersion{}, fmt.Errorf("fetching group info of `%s`: %w", stepID, err)
	}

	resolvedVersion, err := s.resolveVersion(ctx, stepID, version, versionConstraint, latestVersions)
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

// resolveVersion turns a parsed version constraint into a concrete version
// string, fetching the step's version list when the constraint needs it.
func (s *Steplib) resolveVersion(ctx context.Context, stepID, version string, constraint models.VersionConstraint, latestVersions steplibindex.LatestPointer) (string, error) {
	switch constraint.VersionLockType {
	case models.Latest:
		return latestVersions.Latest, nil
	case models.Fixed:
		resolved := constraint.Version.String()
		// Verify the pinned version actually exists; otherwise a typo'd pin
		// surfaces later as an opaque fetch/decode error instead of a clear
		// "no such version".
		allVersions, err := s.api.GetAllStepVersions(ctx, stepID)
		if err != nil {
			return "", fmt.Errorf("fetching all versions of `%s`: %w", stepID, err)
		}
		if !slices.Contains(allVersions, resolved) {
			return "", fmt.Errorf("%s steplib does not contain %s step %s version", s.steplibURI, stepID, resolved)
		}
		return resolved, nil
	case models.MajorLocked:
		majorKey := strconv.FormatUint(constraint.Version.Major, 10)
		v, ok := latestVersions.LatestByMajor[majorKey]
		if !ok {
			return "", fmt.Errorf("%s steplib does not contain %s step with major version %s", s.steplibURI, stepID, majorKey)
		}
		return v, nil
	case models.MinorLocked:
		allVersions, err := s.api.GetAllStepVersions(ctx, stepID)
		if err != nil {
			return "", fmt.Errorf("fetching all versions of `%s`: %w", stepID, err)
		}
		resolved, err := resolveMinorLocked(allVersions, constraint.Version)
		if err != nil {
			return "", fmt.Errorf("%s steplib: %w", s.steplibURI, err)
		}
		return resolved, nil
	default:
		return "", fmt.Errorf("unknown version constraint: %s", version)
	}
}

// toStepGroupInfoModel flattens v2's nested `deprecation` object into v1's
// `RemovalDate` + `DeprecateNotes` fields so the rest of the codebase keeps
// reading the same model shape.
func toStepGroupInfoModel(info steplibindex.StepInfo) models.StepGroupInfoModel {
	// v2's asset_urls is a []string of step-relative URLs; v1's model keys them
	// by asset filename. Rebuild that map (e.g. "assets/icon.svg" -> {"icon.svg": ...}).
	var assetURLs map[string]string
	if len(info.AssetURLs) > 0 {
		assetURLs = make(map[string]string, len(info.AssetURLs))
		for _, u := range info.AssetURLs {
			assetURLs[path.Base(u)] = u
		}
	}
	out := models.StepGroupInfoModel{
		Maintainer:     info.Maintainer,
		AssetURLs:      assetURLs,
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
// constraint's Major+Minor. An unparseable entry is an error — versions.json is
// expected to hold only valid semver.
func resolveMinorLocked(versions []string, constraint models.Semver) (string, error) {
	parsed := make([]models.Semver, 0, len(versions))
	for _, raw := range versions {
		sv, err := models.ParseSemver(raw)
		if err != nil {
			return "", fmt.Errorf("parse version %q: %w", raw, err)
		}
		parsed = append(parsed, sv)
	}

	best, ok := models.HighestForMajorMinor(parsed, constraint)
	if !ok {
		return "", fmt.Errorf("no version matches %d.%d.x", constraint.Major, constraint.Minor)
	}
	return best.String(), nil
}
