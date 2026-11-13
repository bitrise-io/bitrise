package indexgen_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/internal/specfixtures"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/steplibrary/indexgen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogger routes stepman log output to t.Log, so it stays quiet on success
// but surfaces in the failure output of a broken test.
type testLogger struct{ t *testing.T }

func (l testLogger) Debugf(f string, a ...any) { l.t.Logf("DEBUG "+f, a...) }
func (l testLogger) Errorf(f string, a ...any) { l.t.Logf("ERROR "+f, a...) }
func (l testLogger) Warnf(f string, a ...any)  { l.t.Logf("WARN "+f, a...) }
func (l testLogger) Infof(f string, a ...any)  { l.t.Logf("INFO "+f, a...) }

// TestHTTPAPI_Integration generates a V2 inventory from the specfixtures
// testdata into a temp dir, serves it over httptest, and exercises
// steplibrary.HTTPAPI against it. This is a true generator → HTTP → reader
// end-to-end check: the schema shapes are validated against freshly generated
// output, with no checked-in fixture to drift out of sync with the generator.
//
// It lives in indexgen's test package so it can drive the unexported fs.FS
// generator via the GenerateFromSteplibCloneForTest hook without exporting it
// from the production API.
func TestHTTPAPI_Integration(t *testing.T) {
	outDir := t.TempDir()
	_, err := indexgen.GenerateFromSteplibCloneForTest(
		specfixtures.SteplibClone(),
		outDir,
		indexgen.Options{GeneratedAt: time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC), SteplibCommitSHA: ""},
		testLogger{t},
	)
	require.NoError(t, err, "generate V2 inventory")

	// The tree is rooted under v2/; serving outDir lets the reader's v2/-prefixed
	// path helpers resolve against srv.URL directly.
	srv := httptest.NewServer(http.FileServer(http.Dir(outDir)))
	t.Cleanup(srv.Close)

	api := steplibrary.NewHTTPAPI(srv.URL, httpfetch.NewWithClient(srv.Client()))
	ctx := t.Context()

	t.Run("GetAllStepIDs returns sample step IDs", func(t *testing.T) {
		got, gotErr := api.GetAllStepIDs(ctx)
		require.NoError(t, gotErr, "GetAllStepIDs")
		assert.Equal(t, []string{"bash-step", "deprecated-step", "hello-step", "multi-platform-step"}, got, "step IDs")
	})

	t.Run("GetLatestStepVersions(hello-step)", func(t *testing.T) {
		got, gotErr := api.GetLatestStepVersions(ctx, "hello-step")
		require.NoError(t, gotErr, "GetLatestStepVersions")
		assert.Equal(t, "hello-step", got.StepID, "StepID")
		assert.Equal(t, "2.0.0", got.Latest, "Latest")
		assert.Equal(t, "1.1.0", got.LatestByMajor["1"], "LatestByMajor[1]")
		assert.Equal(t, "2.0.0", got.LatestByMajor["2"], "LatestByMajor[2]")
	})

	t.Run("GetAllStepVersions(hello-step) returns newest-first", func(t *testing.T) {
		got, gotErr := api.GetAllStepVersions(ctx, "hello-step")
		require.NoError(t, gotErr, "GetAllStepVersions")
		assert.Equal(t, []string{"2.0.0", "1.1.0", "1.0.0"}, got, "versions")
	})

	t.Run("GetStepGroupInfo(hello-step) shows active step", func(t *testing.T) {
		got, gotErr := api.GetStepGroupInfo(ctx, "hello-step")
		require.NoError(t, gotErr, "GetStepGroupInfo")
		assert.Equal(t, "bitrise", got.Maintainer, "Maintainer")
		assert.Nil(t, got.Deprecation, "Deprecation")
		assert.Equal(t, []string{"assets/icon.svg"}, got.AssetURLs, "AssetURLs")
	})

	t.Run("GetStepGroupInfo(deprecated-step) exposes deprecation metadata", func(t *testing.T) {
		got, gotErr := api.GetStepGroupInfo(ctx, "deprecated-step")
		require.NoError(t, gotErr, "GetStepGroupInfo")
		assert.Equal(t, "bitrise", got.Maintainer, "Maintainer")
		require.NotNil(t, got.Deprecation, "Deprecation")
		assert.Equal(t, "2025-04-11", got.Deprecation.RemovalDate, "RemovalDate")
		assert.Contains(t, got.Deprecation.Notes, "key-based caching", "Notes")
	})

	t.Run("GetStepModel(hello-step, 2.0.0) decodes step.json", func(t *testing.T) {
		got, gotErr := api.GetStepModel(ctx, steplibrary.ResolvedStepVersion{ID: "hello-step", Version: "2.0.0"})
		require.NoError(t, gotErr, "GetStepModel")
		require.NotNil(t, got.Title, "Title")
		assert.Equal(t, "Hello Step", *got.Title, "Title")
		require.NotNil(t, got.SourceCodeURL, "SourceCodeURL")
		assert.Equal(t, "https://github.com/example/hello-step", *got.SourceCodeURL, "SourceCodeURL")
		require.NotNil(t, got.Source, "Source")
		assert.Equal(t, "https://github.com/example/hello-step.git", got.Source.Git, "Source.Git")
	})
}
