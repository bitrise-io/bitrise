package steplibindex

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

// Paths to every file in the V2 inventory tree. There must be exactly one
// source of truth for the layout so the generator, the HTTP reader, and the
// validator never drift.
//
// Each file is one constructor returning a Path, which carries both addressing
// modes so they can't drift from each other:
//
//   - FS():  the inventory-relative path as a forward-slash string — the form
//     io/fs requires (os.DirFS, fs.ReadFile, fs.WalkDir, …). It is slash-
//     separated on every OS, so it is built with path.Join, never
//     filepath.Join; ids and versions are left raw (unescaped).
//   - URL(): absolute (leading slash), with url.PathEscape applied to the
//     dynamic id/version segments — for the HTTP reader.
//
// Dynamic segments (step id, version, asset file) are validated as safe single
// path components: a constructor returns an error if a segment is empty, a
// "."/".." traversal element, or contains a path separator or NUL. This is a
// security boundary — the segment is interpolated into filesystem and URL paths,
// so an unchecked separator or ".." could read or write outside the step's
// subtree. Directories (StepDirFS, IndexStepDirFS, StepAssetDirFS) are FS-only:
// they are path bases / walk roots, never fetched by URL.

// Top-level directories inside the inventory tree. The "steps" segment under
// index/ is the per-step index namespace — distinct from this top-level
// source-of-truth steps/ tree — so it is written inline, not via StepsRootFS.
const (
	StepsRootFS = "steps"
	IndexRootFS = "index"
)

// Path is a single file in the V2 inventory tree, addressable as a filesystem
// path (FS) or an HTTP URL path (URL). Both forms are built together so they
// cannot drift.
type Path struct {
	fs  string
	url string
}

// FS returns the slash-separated path relative to the inventory root.
func (p Path) FS() string { return p.fs }

// URL returns the absolute, percent-escaped URL path.
func (p Path) URL() string { return p.url }

// seg is one path segment. Dynamic segments (step id, version, asset file) are
// validated and percent-escaped in the URL form; static ones are taken verbatim.
type seg struct {
	v   string
	dyn bool
}

func lit(v string) seg { return seg{v: v, dyn: false} }
func dyn(v string) seg { return seg{v: v, dyn: true} }

// validateSegment rejects a dynamic segment that is not a safe single path
// component, so it can never escape or restructure the path it is spliced into.
func validateSegment(s string) error {
	switch {
	case s == "":
		return errors.New("path segment is empty")
	case s == "." || s == "..":
		return fmt.Errorf("path segment %q is a directory-traversal element", s)
	case strings.ContainsAny(s, `/\`):
		return fmt.Errorf("path segment %q contains a path separator", s)
	case strings.ContainsRune(s, 0):
		return fmt.Errorf("path segment %q contains a NUL byte", s)
	}
	return nil
}

// build assembles a Path from segments under the version dir (e.g. v2/),
// validating every dynamic segment. It is the single place each layout is
// spelled out, so the FS and URL forms are always derived from the same list.
func build(segs ...seg) (Path, error) {
	fsParts := []string{VersionDir()}
	urlParts := []string{VersionDir()}
	for _, s := range segs {
		if s.dyn {
			if err := validateSegment(s.v); err != nil {
				return Path{}, err
			}
			urlParts = append(urlParts, url.PathEscape(s.v))
		} else {
			urlParts = append(urlParts, s.v)
		}
		fsParts = append(fsParts, s.v)
	}
	return Path{fs: path.Join(fsParts...), url: "/" + path.Join(urlParts...)}, nil
}

// staticPath builds a Path from static segments only; with no dynamic input
// there is nothing to validate, so it cannot fail.
func staticPath(segs ...string) Path {
	joined := path.Join(append([]string{VersionDir()}, segs...)...)
	return Path{fs: joined, url: "/" + joined}
}

// MetaPath is v2/meta.json.
func MetaPath() Path { return staticPath("meta.json") }

// StepIDsPath is v2/index/step_ids.json.
func StepIDsPath() Path { return staticPath(IndexRootFS, "step_ids.json") }

// LatestPointerPath is v2/index/steps/<id>/latest.json.
func LatestPointerPath(stepID string) (Path, error) {
	return build(lit(IndexRootFS), lit("steps"), dyn(stepID), lit("latest.json"))
}

// VersionsPath is v2/index/steps/<id>/versions.json.
func VersionsPath(stepID string) (Path, error) {
	return build(lit(IndexRootFS), lit("steps"), dyn(stepID), lit("versions.json"))
}

// StepInfoPath is v2/steps/<id>/step-info.json.
func StepInfoPath(stepID string) (Path, error) {
	return build(lit(StepsRootFS), dyn(stepID), lit("step-info.json"))
}

// StepJSONPath is v2/steps/<id>/<version>/step.json.
func StepJSONPath(stepID, version string) (Path, error) {
	return build(lit(StepsRootFS), dyn(stepID), dyn(version), lit("step.json"))
}

// StepAssetPath is v2/steps/<id>/assets/<file>.
func StepAssetPath(stepID, file string) (Path, error) {
	return build(lit(StepsRootFS), dyn(stepID), lit("assets"), dyn(file))
}

// StepDirFS is the v2/steps/<id>/ directory (the per-step source subtree:
// step-info.json, assets/, and per-version dirs).
func StepDirFS(stepID string) (string, error) {
	p, err := build(lit(StepsRootFS), dyn(stepID))
	return p.FS(), err
}

// IndexStepDirFS is the v2/index/steps/<id>/ directory (the per-step subtree of
// derived index files).
func IndexStepDirFS(stepID string) (string, error) {
	p, err := build(lit(IndexRootFS), lit("steps"), dyn(stepID))
	return p.FS(), err
}

// StepAssetDirFS is the v2/steps/<id>/assets directory.
func StepAssetDirFS(stepID string) (string, error) {
	p, err := build(lit(StepsRootFS), dyn(stepID), lit("assets"))
	return p.FS(), err
}
