package steplibrary

import (
	"context"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

type ResolvedStepVersion struct {
	ID, Version string
}

type API interface {
	GetAllStepIDs(ctx context.Context) ([]string, error)
	GetLatestStepVersions(ctx context.Context, id string) (steplibindex.LatestPointer, error)
	// GetAllStepVersions returns all available versions of a step.
	// Mirrors `index/steps/<id>/versions.json` from the V2 inventory layout.
	GetAllStepVersions(ctx context.Context, id string) ([]string, error)
	// GetStepGroupInfo returns version-independent step metadata
	// (maintainer, deprecation, asset URLs). Mirrors `steps/<id>/step-info.json`.
	GetStepGroupInfo(ctx context.Context, id string) (steplibindex.StepInfo, error)
	// GetStepModel fetches the V2 per-version step manifest (mirrors
	// `steps/<id>/<version>/step.json`, which serializes models.StepModel).
	GetStepModel(ctx context.Context, step ResolvedStepVersion) (models.StepModel, error)
	// GetStepSourceDownloadLocations returns the inventory-wide base download locations
	// (mirrors meta.json's download_locations): a zip base and a git marker,
	// from which per-step source URLs are built.
	GetStepSourceDownloadLocations(ctx context.Context) ([]models.DownloadLocationModel, error)
}
