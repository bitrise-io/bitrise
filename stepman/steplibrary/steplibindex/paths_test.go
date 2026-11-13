package steplibindex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaths locks in the FS and URL forms of the dynamic inventory paths. The
// URL form url.PathEscape's the dynamic id/version segments (e.g. a space →
// %20); the FS form stays raw. Both come from one Path so they can't drift.
func TestPaths(t *testing.T) {
	cases := map[string]struct {
		build   func() (Path, error)
		wantFS  string
		wantURL string
	}{
		"latest pointer": {func() (Path, error) { return LatestPointerPath("git-clone") }, "v2/index/steps/git-clone/latest.json", "/v2/index/steps/git-clone/latest.json"},
		"versions":       {func() (Path, error) { return VersionsPath("git-clone") }, "v2/index/steps/git-clone/versions.json", "/v2/index/steps/git-clone/versions.json"},
		"step-info":      {func() (Path, error) { return StepInfoPath("git-clone") }, "v2/steps/git-clone/step-info.json", "/v2/steps/git-clone/step-info.json"},
		"step.json":      {func() (Path, error) { return StepJSONPath("git-clone", "1.0.0") }, "v2/steps/git-clone/1.0.0/step.json", "/v2/steps/git-clone/1.0.0/step.json"},
		"asset":          {func() (Path, error) { return StepAssetPath("git-clone", "icon.svg") }, "v2/steps/git-clone/assets/icon.svg", "/v2/steps/git-clone/assets/icon.svg"},

		// URL form escapes reserved characters in dynamic segments; FS stays raw.
		"id with space":      {func() (Path, error) { return LatestPointerPath("with space") }, "v2/index/steps/with space/latest.json", "/v2/index/steps/with%20space/latest.json"},
		"id with question":   {func() (Path, error) { return LatestPointerPath("weird?chars") }, "v2/index/steps/weird?chars/latest.json", "/v2/index/steps/weird%3Fchars/latest.json"},
		"version with space": {func() (Path, error) { return StepJSONPath("git-clone", "1.0.0 beta") }, "v2/steps/git-clone/1.0.0 beta/step.json", "/v2/steps/git-clone/1.0.0%20beta/step.json"},
		// url.PathEscape leaves "+" alone — it's valid unescaped in a path segment.
		"id with plus": {func() (Path, error) { return LatestPointerPath("with+plus") }, "v2/index/steps/with+plus/latest.json", "/v2/index/steps/with+plus/latest.json"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.build()
			require.NoError(t, err)
			assert.Equal(t, tc.wantFS, got.FS(), "FS")
			assert.Equal(t, tc.wantURL, got.URL(), "URL")
		})
	}
}

// TestStaticPaths locks in the paths with no dynamic input (they can't fail).
func TestStaticPaths(t *testing.T) {
	assert.Equal(t, "v2/meta.json", MetaPath().FS(), "MetaPath FS")
	assert.Equal(t, "/v2/meta.json", MetaPath().URL(), "MetaPath URL")
	assert.Equal(t, "v2/index/step_ids.json", StepIDsPath().FS(), "StepIDsPath FS")
	assert.Equal(t, "/v2/index/step_ids.json", StepIDsPath().URL(), "StepIDsPath URL")
}

// TestPaths_rejectsUnsafeSegments is the security-grade check: a dynamic segment
// that could escape or restructure the path (empty, "."/"..", a separator, NUL)
// is refused with an error rather than spliced in.
func TestPaths_rejectsUnsafeSegments(t *testing.T) {
	cases := map[string]func() (Path, error){
		"empty id":       func() (Path, error) { return StepInfoPath("") },
		"dot id":         func() (Path, error) { return StepInfoPath(".") },
		"dotdot id":      func() (Path, error) { return StepInfoPath("..") },
		"slash id":       func() (Path, error) { return StepInfoPath("a/b") },
		"backslash id":   func() (Path, error) { return StepInfoPath(`a\b`) },
		"nul id":         func() (Path, error) { return StepInfoPath("a\x00b") },
		"dotdot version": func() (Path, error) { return StepJSONPath("git-clone", "..") },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := build()
			assert.Error(t, err)
		})
	}
}

// TestDirPaths locks in the FS-only directory helpers.
func TestDirPaths(t *testing.T) {
	stepDir, err := StepDirFS("git-clone")
	require.NoError(t, err)
	assert.Equal(t, "v2/steps/git-clone", stepDir, "StepDirFS")

	indexDir, err := IndexStepDirFS("git-clone")
	require.NoError(t, err)
	assert.Equal(t, "v2/index/steps/git-clone", indexDir, "IndexStepDirFS")

	assetDir, err := StepAssetDirFS("git-clone")
	require.NoError(t, err)
	assert.Equal(t, "v2/steps/git-clone/assets", assetDir, "StepAssetDirFS")
}
