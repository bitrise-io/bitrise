//go:build steplib_e2e

// Package steplibe2e is an end-to-end suite that activates steplib steps through
// the real bitrise CLI across the legacy (v1) and API (v2) steplib paths,
// crossing precompiled-binary vs source fetch and several step-version forms,
// then captures and diffs the activation logs of the two implementations.
//
// It is build-tagged `steplib_e2e` and excluded from the normal integration
// test run because it depends on network access and hosted steplib infra.
// Drive it via bitrise_e2e_steplib.yml (see the repo root).
package steplibe2e

import "os"

// canonicalSteplibURL must match stepman's inventoryAPIClientFactory check: the
// v2 API path only engages when the workflow's default_step_lib_source is this
// exact URL and the migrate experiment flag is on.
const canonicalSteplibURL = "https://github.com/bitrise-io/bitrise-steplib.git"

// defaultDevInventoryURL is the v2 API inventory base URL used when
// E2E_STEPLIB_API_URL is not set. The orchestrator yml sets the env var.
const defaultDevInventoryURL = "https://storage.googleapis.com/steplib-storage-dev"

// defaultGitCloneURL is the repository git-clone checks out at runtime. Small,
// public, and known-good; overridable via E2E_GIT_CLONE_URL.
const defaultGitCloneURL = "https://github.com/bitrise-io/stepman"

func devInventoryURL() string {
	if v := os.Getenv("E2E_STEPLIB_API_URL"); v != "" {
		return v
	}
	return defaultDevInventoryURL
}

func gitCloneURL() string {
	if v := os.Getenv("E2E_GIT_CLONE_URL"); v != "" {
		return v
	}
	return defaultGitCloneURL
}

// pathVariant selects which stepman activation path a cell exercises.
//
//	v1-source:      legacy local-steplib path, source fetch (no precompiled support)
//	v2-source:      API path, source fetch
//	v2-precompiled: API path, precompiled-binary fetch (falls back to source if
//	                the step has no prebuilt executable — e.g. bash steps)
type pathVariant struct {
	name        string
	useAPI      bool
	precompiled bool
}

var variants = []pathVariant{
	{name: "v1-source", useAPI: false, precompiled: false},
	{name: "v2-source", useAPI: true, precompiled: false},
	{name: "v2-precompiled", useAPI: true, precompiled: true},
}

// versionForm is a way of referencing a step version in a workflow. The empty
// version string means "no @version" (latest). The label drives the version
// constraint type stepman resolves (exact / minor-locked / major-locked / latest).
type versionForm struct {
	label   string
	version string
}

// stepSpec is a step under test plus the version forms and runtime inputs used
// to drive it. inputs values may reference workflow env vars like
// $BITRISE_SOURCE_DIR; they are emitted verbatim into the generated workflow.
type stepSpec struct {
	id             string
	versions       []versionForm
	inputs         map[string]string
	hasPrecompiled bool // documents whether the step ships prebuilt binaries
}

// steps is the curated matrix. Kept deliberately small; adding a step or a
// version is a one-line change here.
//
//   - git-clone: Go toolkit with prebuilt binaries -> exercises precompiled and
//     source fetch across multiple concrete versions + latest.
//   - script:    Bash toolkit, source-only -> exercises the non-Go path and the
//     v2 precompiled->source fallback log.
//
// Only exact pins and "latest" are used as version forms: a bare workflow
// step-ref like `git-clone@8.4` (minor/major lock) is resolved by the v2 API
// resolver but NOT by the v1 activation layer (which expects a concrete version
// dir), so comparing those across paths would be apples-to-oranges. The chosen
// exact versions exist in both the v1 prod steplib and the v2 dev inventory and
// (for git-clone) ship prebuilt binaries.
func steps() []stepSpec {
	return []stepSpec{
		{
			id:             "git-clone",
			hasPrecompiled: true,
			versions: []versionForm{
				{label: "8.5.0", version: "8.5.0"},
				{label: "8.4.2", version: "8.4.2"},
				{label: "latest", version: ""},
			},
			inputs: map[string]string{
				"repository_url": gitCloneURL(),
				"clone_into_dir": "$BITRISE_SOURCE_DIR/_repo",
				"branch":         "master",
				"clone_depth":    "1",
			},
		},
		{
			id:             "script",
			hasPrecompiled: false,
			versions: []versionForm{
				{label: "1.2.1", version: "1.2.1"},
				{label: "latest", version: ""},
			},
			inputs: map[string]string{
				"content": "#!/usr/bin/env bash\necho steplib-e2e-script-ran",
			},
		},
	}
}

// cell is one matrix point: a step at a version form, activated via one path
// variant.
type cell struct {
	step    stepSpec
	version versionForm
	variant pathVariant
}

// allCells expands the matrix. Every step runs every version form through every
// path variant.
func allCells() []cell {
	var cells []cell
	for _, s := range steps() {
		for _, v := range s.versions {
			for _, variant := range variants {
				cells = append(cells, cell{step: s, version: v, variant: variant})
			}
		}
	}
	return cells
}
