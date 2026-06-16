//go:build integration

package indexgen

import (
	"cmp"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestRoundTrip_stepYAML_to_stepJSON locks in the central V2 wire-format
// contract: a generated step.json must be exactly the V1 step.yml that
// produced it, just JSON-encoded.
//
// It runs against the real default steplib — every step, every version — so it
// needs network and builds only under the "integration" build tag (excluded
// from a normal `go test ./...`); STEPLIB_URI overrides the source. For each
// step it parses the source step.yml through the same pipeline the generator
// uses, parses the generated step.json, and asserts the two serialize to equal
// JSON.
//
// The comparison is via the JSON wire format, not reflect.DeepEqual on the Go
// structs, on purpose: a nil and an empty slice serialize identically under
// omitempty, so they're equivalent to every consumer of the bytes. The contract
// is "the bytes carry the same step", not "the in-memory struct is identical".
func TestRoundTrip_stepYAML_to_stepJSON(t *testing.T) {
	uri := cmp.Or(os.Getenv("STEPLIB_URI"), "https://github.com/bitrise-io/bitrise-steplib.git")

	// Generate the V2 tree (Generate clones the steplib into stepman's local
	// cache), then read the clone back to pair each step.yml with its step.json.
	out := t.TempDir()
	_, err := Generate(uri, out, Options{}, testLogger{t})
	require.NoError(t, err, "Generate")

	route, found := stepman.ReadRoute(uri)
	require.True(t, found, "no route for %s after Generate", uri)
	inputFS := os.DirFS(stepman.GetLibraryBaseDirPath(route))
	v2Dir := filepath.Join(out, steplibindex.VersionDir())

	pairs := collectStepYMLAndJSONPaths(t, inputFS, v2Dir)
	require.NotEmpty(t, pairs, "expected at least one step.yml/step.json pair to compare")
	t.Logf("round-tripping %d step versions from %s", len(pairs), uri)

	for _, p := range pairs {
		t.Run(p.yamlPath, func(t *testing.T) {
			fromYAML := parseStepYMLForRoundTrip(t, inputFS, p.yamlPath)
			fromJSON := parseStepJSONForRoundTrip(t, p.jsonPath)

			yamlAsJSON, err := json.Marshal(fromYAML)
			require.NoError(t, err, "re-marshal fromYAML")
			jsonAsJSON, err := json.Marshal(fromJSON)
			require.NoError(t, err, "re-marshal fromJSON")

			assert.JSONEq(t, string(yamlAsJSON), string(jsonAsJSON),
				"step.yml and the generated step.json must serialize to semantically equal JSON")
		})
	}
}

type stepYAMLJSONPair struct {
	yamlPath string // path within inputFS, e.g. "steps/hello-step/1.0.0/step.yml"
	jsonPath string // absolute path on disk, e.g. "<v2>/steps/hello-step/1.0.0/step.json"
}

// collectStepYMLAndJSONPaths walks inputFS's source steps/ tree for every
// step.yml and returns the matching pair of (input step.yml path, generated
// step.json path under the v2 output dir). Asserts each generated step.json
// exists.
func collectStepYMLAndJSONPaths(t *testing.T, inputFS fs.FS, outV2Dir string) []stepYAMLJSONPair {
	t.Helper()
	var pairs []stepYAMLJSONPair

	require.NoError(t, fs.WalkDir(inputFS, steplibindex.StepsRootFS, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		jsonPath, ok := generatedStepJSONPath(p, outV2Dir)
		if !ok {
			return nil
		}
		require.FileExists(t, jsonPath, "generated step.json missing for %s", p)
		pairs = append(pairs, stepYAMLJSONPair{yamlPath: p, jsonPath: jsonPath})
		return nil
	}))
	return pairs
}

// generatedStepJSONPath maps a source "steps/<id>/<version>/step.yml" path to
// its generated step.json path under v2Dir. ok is false for paths the
// generator doesn't emit a step.json for: non-step.yml files and non-semver
// version dirs.
func generatedStepJSONPath(p, v2Dir string) (jsonPath string, ok bool) {
	if filepath.Base(p) != "step.yml" {
		return "", false
	}
	rel := strings.TrimSuffix(p, "/step.yml") // steps/<id>/<version>
	segs := strings.Split(rel, "/")
	if len(segs) != 3 {
		return "", false // not a steps/<id>/<version> path
	}
	if _, err := models.ParseSemver(segs[2]); err != nil {
		return "", false // non-semver version dir: generator skips these
	}
	return filepath.Join(v2Dir, rel, "step.json"), true
}

// parseStepYMLForRoundTrip mirrors the generator's canonical step.yml parse
// pipeline. It is duplicated here on purpose: the round-trip property we're
// asserting is "the bytes the generator emits decode back to whatever its
// pipeline produced", so we re-derive the expected model from first principles
// rather than calling the generator's internal helper.
func parseStepYMLForRoundTrip(t *testing.T, inputFS fs.FS, ymlPath string) models.StepModel {
	t.Helper()
	bytes, err := fs.ReadFile(inputFS, ymlPath)
	require.NoError(t, err, "read %s", ymlPath)

	var step models.StepModel
	require.NoError(t, yaml.Unmarshal(bytes, &step), "yaml.Unmarshal %s", ymlPath)
	require.NoError(t, step.Normalize(), "Normalize %s", ymlPath)
	require.NoError(t, step.FillMissingDefaults(), "FillMissingDefaults %s", ymlPath)
	return step
}

func parseStepJSONForRoundTrip(t *testing.T, jsonPath string) models.StepModel {
	t.Helper()
	bytes, err := os.ReadFile(jsonPath)
	require.NoError(t, err, "read %s", jsonPath)

	var step models.StepModel
	require.NoError(t, json.Unmarshal(bytes, &step), "json.Unmarshal %s", jsonPath)
	return step
}
