package steplibrary

import (
	"context"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/spec"
)

type ResolvedStepVersion struct {
	ID, Version string
}

type API interface {
	GetAllStepIDs(ctx context.Context) ([]string, error)
	GetLatestStepVersions(ctx context.Context, id string) (spec.LatestPointer, error)
	// GetAllStepVersions returns all available versions of a step.
	// Mirrors `spec/steps/<id>/versions.json` from the V2 inventory layout;
	// the per-version metadata is dropped for now since callers only need the
	// version strings to resolve MinorLocked constraints.
	GetAllStepVersions(ctx context.Context, id string) ([]string, error)
	// GetStepGroupInfo returns version-independent step metadata
	// (maintainer, deprecation, asset URLs). Mirrors `steps/<id>/step-info.json`.
	GetStepGroupInfo(ctx context.Context, id string) (spec.StepInfo, error)
	// GetStepModel fetches the V2 per-version step manifest (mirrors
	// `steps/<id>/<version>/step.json`, which serializes models.StepModel).
	GetStepModel(ctx context.Context, step ResolvedStepVersion) (models.StepModel, error)
}
