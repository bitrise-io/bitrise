package indexgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flagMatching returns the first violation whose Path and Msg both contain the
// given substrings, or nil. Lets cases assert "the right violation fired"
// without coupling to exact message wording.
func flagMatching(errs []ValidationError, pathContains, msgContains string) *ValidationError {
	for i := range errs {
		if (pathContains == "" || strings.Contains(errs[i].Path, pathContains)) &&
			(msgContains == "" || strings.Contains(errs[i].Msg, msgContains)) {
			return &errs[i]
		}
	}
	return nil
}

func seedFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root, relPath), []byte(content), 0o644))
}

func removeFile(t *testing.T, root, relPath string) {
	t.Helper()
	require.NoError(t, os.Remove(filepath.Join(root, relPath)))
}

// mustFS returns p.FS() for a path constructor result, panicking if it errored.
// The test ids/versions are all valid, so the error is a test-setup bug.
func mustFS(p steplibindex.Path, err error) string {
	if err != nil {
		panic(err)
	}
	return p.FS()
}

// TestValidate mutates a freshly generated (and therefore valid) inventory and
// asserts the matching consistency check fires. The clean baseline must produce
// no violations; every other case proves a specific check. Comparison via the
// generator's own output keeps the fixtures honest — no hand-built tree to
// drift from what the generator actually emits.
func TestValidate(t *testing.T) {
	// wantMsg == "" means "expect no violations"; otherwise the run must produce
	// a violation whose Path and Msg contain wantPath / wantMsg.
	cases := map[string]struct {
		mutate   func(t *testing.T, root string)
		wantPath string
		wantMsg  string
	}{
		"clean generated output is valid": {
			mutate: func(*testing.T, string) {},
		},
		"missing meta.json": {
			mutate:   func(t *testing.T, root string) { removeFile(t, root, steplibindex.MetaPath().FS()) },
			wantPath: "meta.json", wantMsg: "missing",
		},
		"meta.json invalid JSON": {
			mutate:   func(t *testing.T, root string) { seedFile(t, root, steplibindex.MetaPath().FS(), "not json") },
			wantPath: "meta.json", wantMsg: "invalid JSON",
		},
		"meta.json wrong format_version": {
			mutate: func(t *testing.T, root string) {
				seedFile(t, root, steplibindex.MetaPath().FS(), `{"format_version": 99, "updated_at": "2026-05-15T12:00:00Z"}`)
			},
			wantPath: "meta.json", wantMsg: "format_version",
		},
		"empty step_ids": {
			mutate:   func(t *testing.T, root string) { seedFile(t, root, steplibindex.StepIDsPath().FS(), `{"step_ids":[]}`) },
			wantPath: "step_ids.json", wantMsg: "empty",
		},
		"unsorted step_ids": {
			mutate: func(t *testing.T, root string) {
				// Still references only real steps so only the sort check fires.
				seedFile(t, root, steplibindex.StepIDsPath().FS(), `{"step_ids": ["hello-step", "bash-step", "deprecated-step", "multi-platform-step"]}`)
			},
			wantPath: "step_ids.json", wantMsg: "not sorted",
		},
		"latest points at a missing version": {
			mutate: func(t *testing.T, root string) {
				seedFile(t, root, mustFS(steplibindex.LatestPointerPath("hello-step")), `{"step_id": "hello-step", "latest": "99.99.99", "latest_by_major": {"99": "99.99.99"}}`)
			},
			wantPath: "hello-step/latest.json", wantMsg: "not in",
		},
		"latest_by_major points at a different major": {
			mutate: func(t *testing.T, root string) {
				// hello-step has 1.0.0, 1.1.0, 2.0.0; point major "1" at 2.0.0.
				seedFile(t, root, mustFS(steplibindex.LatestPointerPath("hello-step")), `{"step_id": "hello-step", "latest": "2.0.0", "latest_by_major": {"1": "2.0.0", "2": "2.0.0"}}`)
			},
			wantPath: "hello-step/latest.json", wantMsg: "different major",
		},
		"latest.json step_id mismatch": {
			mutate: func(t *testing.T, root string) {
				seedFile(t, root, mustFS(steplibindex.LatestPointerPath("hello-step")), `{"step_id": "wrong-id", "latest": "2.0.0", "latest_by_major": {"1": "1.1.0", "2": "2.0.0"}}`)
			},
			wantPath: "hello-step/latest.json", wantMsg: "expected",
		},
		"declared version missing its step.json": {
			mutate: func(t *testing.T, root string) {
				removeFile(t, root, mustFS(steplibindex.StepJSONPath("hello-step", "1.0.0")))
			},
			wantPath: "hello-step/1.0.0/step.json", wantMsg: "missing",
		},
		"missing mandatory step-info.json": {
			mutate:   func(t *testing.T, root string) { removeFile(t, root, mustFS(steplibindex.StepInfoPath("hello-step"))) },
			wantPath: "hello-step/step-info.json", wantMsg: "missing",
		},
		"step-info asset that does not exist": {
			mutate: func(t *testing.T, root string) {
				removeFile(t, root, mustFS(steplibindex.StepAssetPath("hello-step", "icon.svg")))
			},
			wantPath: "hello-step/step-info.json", wantMsg: "does not exist",
		},
		"absolute asset_urls entry": {
			mutate: func(t *testing.T, root string) {
				// Keep the real relative asset so only the absolute-URL check fires.
				seedFile(t, root, mustFS(steplibindex.StepInfoPath("hello-step")), `{"maintainer":"bitrise","deprecation":null,"asset_urls":["assets/icon.svg","https://cdn.example/icon.svg"]}`)
			},
			wantPath: "hello-step/step-info.json", wantMsg: "absolute URL",
		},
		"absolute asset_urls path": {
			mutate: func(t *testing.T, root string) {
				// "/assets/icon.svg" mirrors the real relative asset, so it would
				// resolve and pass if absolute paths weren't rejected outright. Keep
				// the real relative entry so only the absolute-path check fires.
				seedFile(t, root, mustFS(steplibindex.StepInfoPath("hello-step")), `{"maintainer":"bitrise","deprecation":null,"asset_urls":["assets/icon.svg","/assets/icon.svg"]}`)
			},
			wantPath: "hello-step/step-info.json", wantMsg: "absolute path",
		},
		"asset_urls escapes the step directory": {
			mutate: func(t *testing.T, root string) {
				// "../../bash-step/..." resolves to a sibling step's real file, so it
				// would pass if only absolute paths were rejected. Keep the real
				// relative asset so only the escape check fires.
				seedFile(t, root, mustFS(steplibindex.StepInfoPath("hello-step")), `{"maintainer":"bitrise","deprecation":null,"asset_urls":["assets/icon.svg","assets/../../bash-step/step-info.json"]}`)
			},
			wantPath: "hello-step/step-info.json", wantMsg: "escapes",
		},
		"asset on disk missing from step-info.json": {
			mutate: func(t *testing.T, root string) {
				seedFile(t, root, mustFS(steplibindex.StepAssetPath("hello-step", "extra.svg")), "<svg/>")
			},
			wantPath: "hello-step/assets/extra.svg", wantMsg: "unexpected",
		},
		"stale file under index/": {
			mutate: func(t *testing.T, root string) {
				seedFile(t, root, filepath.Join(steplibindex.VersionDir(), steplibindex.IndexRootFS, "stale.json"), "{}")
			},
			wantPath: "stale.json", wantMsg: "unexpected",
		},
		"step dir not in step_ids.json": {
			mutate: func(t *testing.T, root string) {
				p := filepath.Join(root, mustFS(steplibindex.StepJSONPath("ghost-step", "1.0.0")))
				require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
				require.NoError(t, os.WriteFile(p, []byte(`{"title":"ghost"}`), 0o644))
			},
			wantPath: "ghost-step", wantMsg: "unexpected",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			root := runGenerateFromSteplibClone(t)
			tc.mutate(t, root)

			errs := Validate(os.DirFS(root))

			if tc.wantMsg == "" {
				assert.Empty(t, errs)
				return
			}
			assert.NotNil(t, flagMatching(errs, tc.wantPath, tc.wantMsg),
				"want a %q violation at %q; got %+v", tc.wantMsg, tc.wantPath, errs)
		})
	}
}
