// Package steplibindex defines the Go types for the V2 step library inventory wire
// format. It is shared between the generator (steplibrary/indexgen) and the
// read path (steplibrary).
//
// The V2 layout splits the inventory into two URL prefixes:
//   - steps/  — source of truth, self-contained per step, immutable per version
//   - spec/   — derived index files, regeneratable from steps/, short-TTL
//
// All files are JSON. Per-version step manifests (steps/<id>/<v>/step.json)
// use models.StepModel directly; the types below describe the new
// inventory-level and index files that have no V1 equivalent.
//
// Wire shape is kept stable: index and per-step collections always render as
// [] or {} (never null) and nullable values (Deprecation, published_at) as
// null. Optional inventory metadata (Meta.steplib_commit_sha,
// Meta.steplib_source, Meta.download_locations) uses omitempty and is dropped
// when empty.
package steplibindex

import (
	"fmt"
	"time"

	"github.com/bitrise-io/stepman/models"
)

// FormatVersion is the on-disk schema version recorded in Meta. Bump only on
// breaking changes; additive changes (new optional fields) do not bump.
const FormatVersion = 2

// VersionDir is the inventory's top-level directory for this format version
// (e.g. "v2"). The whole tree is rooted under it so multiple format versions
// can be hosted side by side; readers prefix their fetch URLs with it.
func VersionDir() string { return fmt.Sprintf("v%d", FormatVersion) }

// Meta is the inventory-level metadata file at the inventory root (meta.json).
// It is the only file that carries FormatVersion.
type Meta struct {
	FormatVersion     int                            `json:"format_version"`
	UpdatedAt         time.Time                      `json:"updated_at"`
	SteplibCommitSHA  string                         `json:"steplib_commit_sha,omitempty"`
	SteplibSource     string                         `json:"steplib_source,omitempty"`
	DownloadLocations []models.DownloadLocationModel `json:"download_locations,omitempty"`
}

// StepInfo is the per-step metadata file at steps/<id>/step-info.json.
// Holds facts that span versions: maintainer, deprecation, asset list.
// Asset URLs are relative to the file's own location for self-containment.
// Deprecation is null for active steps; AssetURLs is [] for steps with no assets.
type StepInfo struct {
	Maintainer  string       `json:"maintainer"`
	Deprecation *Deprecation `json:"deprecation"`
	AssetURLs   []string     `json:"asset_urls"`
}

// Deprecation carries the removal_date and notes for a deprecated step.
// A nil Deprecation on StepInfo means the step is active.
type Deprecation struct {
	RemovalDate string `json:"removal_date"`
	Notes       string `json:"notes"`
}

// StepIDs is spec/step_ids.json: sorted list of all step IDs in the steplib.
type StepIDs struct {
	StepIDs []string `json:"step_ids"`
}

// LatestPointer is spec/steps/<id>/latest.json: per-step latest pointers.
// Answers Latest and MajorLocked constraints in a single small fetch.
type LatestPointer struct {
	StepID        string            `json:"step_id"`
	Latest        string            `json:"latest"`
	LatestByMajor map[string]string `json:"latest_by_major"`
}

// Versions is spec/steps/<id>/versions.json: the per-step list of version
// strings, newest-first (so versions[0] is the latest). Per-version detail
// (commit, published_at, ...) lives in each steps/<id>/<version>/step.json; the
// latest pointer lives in latest.json.
type Versions struct {
	StepID   string   `json:"step_id"`
	Versions []string `json:"versions"`
}
