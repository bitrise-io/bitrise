package steplib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/require"
)

// Tests for activateStepSourceWithAPI (V2 source activation) — its guard clauses.
// The happy-path id → step.zip URL flow is covered end-to-end by
// TestActivateStep_ResolvesExactVersion in activate_test.go, which serves the
// archive at the id-keyed path (so a wrong id would 404 there).

func TestActivateStepSourceWithAPI_OfflineFails(t *testing.T) {
	log := apiTestLogger{t}
	client := steplibrary.New(log, "http://unused.invalid")

	err := activateStepSourceWithAPI(client, "hello-step", "2.0.0",
		&models.StepSourceModel{Git: "https://github.com/example/hello-step.git"}, t.TempDir(), log, true, httpfetch.NewClient(log))

	require.Error(t, err)
	require.Contains(t, err.Error(), "offline")
}

func TestActivateStepSourceWithAPI_MissingSourceGitFails(t *testing.T) {
	log := apiTestLogger{t}
	client := steplibrary.New(log, "http://unused.invalid")

	t.Run("nil source", func(t *testing.T) {
		err := activateStepSourceWithAPI(client, "hello-step", "2.0.0", nil, t.TempDir(), log, false, httpfetch.NewClient(log))
		require.Error(t, err)
		require.Contains(t, err.Error(), "source git")
	})

	t.Run("empty git URL", func(t *testing.T) {
		err := activateStepSourceWithAPI(client, "hello-step", "2.0.0",
			&models.StepSourceModel{Git: ""}, t.TempDir(), log, false, httpfetch.NewClient(log))
		require.Error(t, err)
		require.Contains(t, err.Error(), "source git")
	})
}

// When the inventory publishes no download locations, source activation fails
// with a clear error rather than silently doing nothing.
func TestActivateStepSourceWithAPI_NoDownloadLocationFails(t *testing.T) {
	root := t.TempDir()
	// meta.json with no download_locations; that's all StepSourceDownloadLocations reads.
	writeInventoryJSON(t, root, steplibindex.MetaPath(), steplibindex.Meta{
		FormatVersion: steplibindex.FormatVersion,
	})
	srv := httptest.NewServer(http.FileServer(http.Dir(root)))
	t.Cleanup(srv.Close)

	log := apiTestLogger{t}
	client := steplibrary.New(log, srv.URL)

	err := activateStepSourceWithAPI(client, "hello-step", "2.0.0",
		&models.StepSourceModel{Git: "https://github.com/example/hello-step.git"}, t.TempDir(), log, false, httpfetch.NewClient(log))

	require.Error(t, err)
	require.Contains(t, err.Error(), "no download location")
}
