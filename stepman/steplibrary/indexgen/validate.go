package indexgen

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

// ValidationError is a single consistency violation found by Validate.
// Path is the slash-separated path of the file the violation belongs to,
// or "" for tree-level issues. Msg explains what's wrong.
type ValidationError struct {
	Path string
	Msg  string
}

func (e ValidationError) Error() string {
	if e.Path == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

// violationf builds a ValidationError with a formatted message.
func violationf(p, format string, args ...any) ValidationError {
	return ValidationError{Path: p, Msg: fmt.Sprintf(format, args...)}
}

// Validate walks the V2 inventory tree rooted at inventoryFS and returns the
// list of consistency violations. An empty slice means the tree is internally
// consistent.
//
// Intended uses:
//   - Pre-deploy CI gate: run against a freshly generated tree before
//     publishing to the V2 host. Fail the build on any violation.
//   - Generator test smoke check: run a generated tree through Validate to
//     catch cross-file inconsistencies that per-file assertions would miss.
//
// A filesystem traversal failure (e.g. an unreadable directory) is itself
// reported as a violation, so callers only need to check whether the returned
// slice is empty.
func Validate(inventoryFS fs.FS) []ValidationError {
	v := &validator{fs: inventoryFS, seen: map[string]bool{}}
	issues := v.collect()

	// Sort violations by (Path, Msg) so the output is deterministic across runs.
	// Without this, map-iteration order leaks into the output and breaks any
	// consumer that diffs error logs (golden files, CI dashboards, etc.).
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		return issues[i].Msg < issues[j].Msg
	})
	return issues
}

// validator carries the read-only inventory and the set of paths consumed so
// far. Each check returns its own violations; the consumed set is the one piece
// of cross-check state (populated as files are read, drained by the stale-file
// sweep at the end).
type validator struct {
	fs   fs.FS
	seen map[string]bool // paths the validator has consumed; remainder is stale
}

func (v *validator) consume(p string) { v.seen[p] = true }

func (v *validator) collect() []ValidationError {
	var issues []ValidationError

	var meta steplibindex.Meta
	metaIssues, ok := v.readJSON(steplibindex.MetaPath().FS(), &meta)
	issues = append(issues, metaIssues...)
	if ok {
		issues = append(issues, v.checkMeta(meta)...)
	}

	var stepIDs steplibindex.StepIDs
	stepIDsIssues, haveIDs := v.readJSON(steplibindex.StepIDsPath().FS(), &stepIDs)
	issues = append(issues, stepIDsIssues...)
	if haveIDs {
		if len(stepIDs.StepIDs) == 0 {
			issues = append(issues, violationf(steplibindex.StepIDsPath().FS(), "step_ids is empty: a steplib must contain at least one step"))
		}
		issues = append(issues, v.checkStepIDsSorted(stepIDs)...)
		for _, id := range stepIDs.StepIDs {
			issues = append(issues, v.checkStep(id)...)
		}
	}

	issues = append(issues, v.staleFileViolations()...)
	return issues
}

// readJSON marks the file as consumed and parses it into `into`. ok is false
// (and a violation is returned) on missing/unreadable/invalid JSON; on success
// it returns no violations.
func (v *validator) readJSON(p string, into any) (issues []ValidationError, ok bool) {
	v.consume(p)
	bytes, err := fs.ReadFile(v.fs, p)
	if err != nil {
		return []ValidationError{violationf(p, "missing or unreadable: %s", err)}, false
	}
	if err := json.Unmarshal(bytes, into); err != nil {
		return []ValidationError{violationf(p, "invalid JSON: %s", err)}, false
	}
	return nil, true
}

func (v *validator) checkMeta(m steplibindex.Meta) []ValidationError {
	var issues []ValidationError
	if m.FormatVersion != steplibindex.FormatVersion {
		issues = append(issues, violationf(steplibindex.MetaPath().FS(), "format_version is %d, expected %d", m.FormatVersion, steplibindex.FormatVersion))
	}
	if m.UpdatedAt.IsZero() {
		issues = append(issues, violationf(steplibindex.MetaPath().FS(), "updated_at is zero"))
	}
	return issues
}

func (v *validator) checkStepIDsSorted(ids steplibindex.StepIDs) []ValidationError {
	var issues []ValidationError
	if !sort.StringsAreSorted(ids.StepIDs) {
		issues = append(issues, violationf(steplibindex.StepIDsPath().FS(), "step_ids is not sorted lexicographically"))
	}
	// Duplicate detection.
	seen := make(map[string]bool, len(ids.StepIDs))
	for _, id := range ids.StepIDs {
		if seen[id] {
			issues = append(issues, violationf(steplibindex.StepIDsPath().FS(), "duplicate step id %q", id))
		}
		seen[id] = true
	}
	return issues
}

func (v *validator) checkStep(id string) []ValidationError {
	// Build the step's paths once; a non-nil error means the id itself is not a
	// safe path segment, which is a violation of step_ids.json.
	latestP, errLatest := steplibindex.LatestPointerPath(id)
	versionsP, errVersions := steplibindex.VersionsPath(id)
	infoP, errInfo := steplibindex.StepInfoPath(id)
	stepDir, errDir := steplibindex.StepDirFS(id)
	if err := cmp.Or(errLatest, errVersions, errInfo, errDir); err != nil {
		return []ValidationError{violationf(steplibindex.StepIDsPath().FS(), "step id %q is invalid: %s", id, err)}
	}
	latestPath := latestP.FS()
	versionsPath := versionsP.FS()

	var issues []ValidationError

	var latest steplibindex.LatestPointer
	latestIssues, haveLatest := v.readJSON(latestPath, &latest)
	issues = append(issues, latestIssues...)

	var versions steplibindex.Versions
	versionsIssues, haveVersions := v.readJSON(versionsPath, &versions)
	issues = append(issues, versionsIssues...)

	if haveLatest && latest.StepID != id {
		issues = append(issues, violationf(latestPath, "step_id is %q, expected %q", latest.StepID, id))
	}
	if haveVersions && versions.StepID != id {
		issues = append(issues, violationf(versionsPath, "step_id is %q, expected %q", versions.StepID, id))
	}

	// Cross-check pointers against the versions list.
	declaredVersions := map[string]bool{}
	if haveVersions {
		for _, ver := range versions.Versions {
			declaredVersions[ver] = true
		}
	}
	if haveLatest && haveVersions {
		if !declaredVersions[latest.Latest] {
			issues = append(issues, violationf(latestPath, "latest %q is not in %s", latest.Latest, versionsPath))
		}
		for major, ver := range latest.LatestByMajor {
			if !declaredVersions[ver] {
				issues = append(issues, violationf(latestPath, "latest_by_major[%q]=%q is not in versions.json", major, ver))
			}
			if !strings.HasPrefix(ver, major+".") {
				issues = append(issues, violationf(latestPath, "latest_by_major[%q]=%q has a different major", major, ver))
			}
		}
	}

	// Every declared version must have its step.json on disk.
	if haveVersions {
		for _, ver := range versions.Versions {
			issues = append(issues, v.checkStepJSON(id, ver, versionsPath)...)
		}
	}

	issues = append(issues, v.checkStepInfo(infoP.FS(), stepDir)...)
	return issues
}

// checkStepJSON validates v2/steps/<id>/<version>/step.json. versionsPath is the
// step's versions.json, where an invalid version string is reported (the version
// comes from there).
func (v *validator) checkStepJSON(id, version, versionsPath string) []ValidationError {
	stepJSON, err := steplibindex.StepJSONPath(id, version)
	if err != nil {
		return []ValidationError{violationf(versionsPath, "version %q is invalid: %s", version, err)}
	}
	p := stepJSON.FS()
	var step models.StepModel
	if issues, ok := v.readJSON(p, &step); !ok {
		return issues
	}
	if step.Source == nil {
		return []ValidationError{violationf(p, "missing source")}
	}
	var issues []ValidationError
	if step.Source.Git == "" {
		issues = append(issues, violationf(p, "missing source.git"))
	}
	if step.Source.Commit == "" {
		issues = append(issues, violationf(p, "missing source.commit"))
	}
	return issues
}

// checkStepInfo validates the step's step-info.json (at infoPath) and that each
// asset_urls entry is a step-relative path resolving to a real file under stepDir.
func (v *validator) checkStepInfo(infoPath, stepDir string) []ValidationError {
	if _, err := fs.Stat(v.fs, infoPath); err != nil {
		// step-info.json is mandatory: the generator writes it for every step.
		v.consume(infoPath)
		return []ValidationError{violationf(infoPath, "missing or unreadable: %s", err)}
	}
	var info steplibindex.StepInfo
	if issues, ok := v.readJSON(infoPath, &info); !ok {
		return issues
	}
	var issues []ValidationError
	for _, rel := range info.AssetURLs {
		issues = append(issues, v.checkAssetURL(infoPath, stepDir, rel)...)
	}
	return issues
}

// checkAssetURL validates one asset_urls entry (attributed to infoPath): it must
// be a clean step-relative reference — no absolute URL or path, no parent-dir
// traversal — that resolves to a real file under stepDir. On success it marks
// the resolved asset consumed so the stale-file sweep won't flag it.
func (v *validator) checkAssetURL(infoPath, stepDir, rel string) []ValidationError {
	switch {
	case strings.Contains(rel, "://"):
		return []ValidationError{violationf(infoPath, "asset_urls entry %q is an absolute URL; must be step-relative", rel)}
	case path.IsAbs(rel):
		return []ValidationError{violationf(infoPath, "asset_urls entry %q is an absolute path; must be step-relative", rel)}
	}
	// path.Join cleans the result, collapsing any "../"; if it no longer sits
	// under stepDir the entry escaped the step's own directory.
	resolved := path.Join(stepDir, rel)
	if resolved != stepDir && !strings.HasPrefix(resolved, stepDir+"/") {
		return []ValidationError{violationf(infoPath, "asset_urls entry %q escapes the step directory; must be step-relative", rel)}
	}
	if _, err := fs.Stat(v.fs, resolved); err != nil {
		return []ValidationError{violationf(infoPath, "asset_urls entry %q points to %q which does not exist", rel, resolved)}
	}
	v.consume(resolved)
	return nil
}

// staleFileViolations walks v2/steps and v2/index once each and returns a
// violation for every file the checks above did not consume — left-over files
// from a previous generation (a removed step), or a stray file from a generator
// bug. It reads the accumulated seen set, so it must run after all other checks.
func (v *validator) staleFileViolations() []ValidationError {
	root := steplibindex.VersionDir()
	issues := []ValidationError{}
	walkErr := fs.WalkDir(v.fs, root, func(p string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			// A missing or unreadable expected root is itself a violation.
			issues = append(issues, violationf(p, "walk failed: %s", err))
		case !d.IsDir() && !v.seen[p]:
			issues = append(issues, violationf(p, "unexpected file under %s/", root))
		}
		return nil
	})
	if walkErr != nil {
		// The callback handles each entry's error and always returns nil, so
		// this only fires if that ever changes — don't let it be dropped.
		issues = append(issues, violationf(root, "walk aborted: %s", walkErr))
	}
	return issues
}
