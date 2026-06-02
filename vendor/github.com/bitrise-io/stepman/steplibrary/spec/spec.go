// Package spec defines the Go types for the V2 step library inventory wire
// format described in STEP-2374-plan.md. It is shared between the generator
// (steplibrary/specgen) and the read path (steplibrary).
//
// The V2 layout splits the inventory into two URL prefixes:
//   - steps/  — source of truth, self-contained per step, immutable per version
//   - spec/   — derived index files, regeneratable from steps/, short-TTL
//
// All files are JSON. Per-version step manifests (steps/<id>/<v>/step.json)
// use models.StepModel directly; the types below describe the new
// inventory-level and index files that have no V1 equivalent.
package spec

import (
	"time"

	"github.com/bitrise-io/stepman/models"
)

// FormatVersion is the on-disk schema version recorded in Meta. Bump only on
// breaking changes; additive changes (new optional fields) do not bump.
const FormatVersion = 2

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
// The Deprecation field is always present in JSON (null for active steps).
type StepInfo struct {
	Maintainer  string            `json:"maintainer,omitempty"`
	Deprecation *Deprecation      `json:"deprecation"`
	AssetURLs   map[string]string `json:"asset_urls,omitempty"`
}

// Deprecation carries the removal_date and notes for a deprecated step.
// A nil Deprecation on StepInfo means the step is active.
type Deprecation struct {
	RemovalDate string `json:"removal_date,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// StepIDs is spec/step_ids.json: sorted list of all step IDs in the steplib.
type StepIDs struct {
	StepIDs []string `json:"step_ids"`
}

// Catalog is spec/latest_versions.json: fat catalog for browse views.
// A single fetch gives everything needed to render a catalog without per-step
// round trips.
type Catalog struct {
	GeneratedAt      time.Time              `json:"generated_at"`
	SteplibCommitSHA string                 `json:"steplib_commit_sha,omitempty"`
	Steps            map[string]CatalogEntry `json:"steps"`
}

// CatalogEntry is one step's entry in Catalog.Steps. Asset URLs are
// inventory-root-relative so consumers can resolve them against the base URL
// the catalog was fetched from, without any hosting URL baked in.
type CatalogEntry struct {
	LatestVersion   string            `json:"latest_version"`
	PublishedAt     *time.Time        `json:"published_at,omitempty"`
	Title           string            `json:"title,omitempty"`
	Summary         string            `json:"summary,omitempty"`
	Maintainer      string            `json:"maintainer,omitempty"`
	TypeTags        []string          `json:"type_tags,omitempty"`
	ProjectTypeTags []string          `json:"project_type_tags,omitempty"`
	HostOsTags      []string          `json:"host_os_tags,omitempty"`
	Website         string            `json:"website,omitempty"`
	SourceCodeURL   string            `json:"source_code_url,omitempty"`
	SupportURL      string            `json:"support_url,omitempty"`
	AssetURLs       map[string]string `json:"asset_urls,omitempty"`
	HasExecutable   bool              `json:"has_executable"`
	Deprecation     *Deprecation      `json:"deprecation"`
}

// LatestPointer is spec/steps/<id>/latest.json: per-step latest pointers.
// Answers Latest and MajorLocked constraints in a single small fetch.
type LatestPointer struct {
	StepID        string            `json:"step_id"`
	Latest        string            `json:"latest"`
	LatestByMajor map[string]string `json:"latest_by_major,omitempty"`
}

// Versions is spec/steps/<id>/versions.json: per-step version list with the
// metadata stepman needs for MinorLocked resolution and binary-availability
// checks. Ordered newest-first.
type Versions struct {
	StepID   string         `json:"step_id"`
	Latest   string         `json:"latest"`
	Versions []VersionEntry `json:"versions"`
}

// VersionEntry is a single entry in Versions.Versions.
type VersionEntry struct {
	Version       string     `json:"version"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	HasExecutable bool       `json:"has_executable"`
	Commit        string     `json:"commit,omitempty"`
}
