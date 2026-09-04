package steplib

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/require"
)

// Shared hermetic fixtures for the steplib package tests. A steplibrary.Client
// pointed at serveHelloStepInventory's server exercises the real read and source
// paths with no network and no git-cloned steplib; the inventory is built from
// the real steplibindex/models structs via the real Path helpers, so it cannot
// structurally drift from the reader.

const testSteplibURL = "https://github.com/bitrise-io/bitrise-steplib.git"

type apiTestLogger struct{ t *testing.T }

func (l apiTestLogger) Debugf(f string, a ...any) { l.t.Logf("DEBUG "+f, a...) }
func (l apiTestLogger) Errorf(f string, a ...any) { l.t.Logf("ERROR "+f, a...) }
func (l apiTestLogger) Warnf(f string, a ...any)  { l.t.Logf("WARN "+f, a...) }
func (l apiTestLogger) Infof(f string, a ...any)  { l.t.Logf("INFO "+f, a...) }

// helloStepVersions is newest-first, matching the versions.json contract.
var helloStepVersions = []string{"2.0.0", "1.1.0", "1.0.0"}

// serveHelloStepInventory writes and serves a minimal V2 inventory for
// "hello-step" and returns the server. The source archive for every version is
// served from the same server (keyed by step ID + version, the correct layout),
// so source activation resolves entirely against localhost.
func serveHelloStepInventory(t *testing.T) *httptest.Server {
	return serveHelloStepInventoryWithExecutables(t, nil)
}

// serveHelloStepInventoryWithExecutables is serveHelloStepInventory with an
// optional executables block applied to every version's step.json, so the
// precompiled-activation branch can be exercised.
func serveHelloStepInventoryWithExecutables(t *testing.T, executables *models.Executables) *httptest.Server {
	t.Helper()
	root := t.TempDir()

	writeInventoryJSON(t, root, steplibindex.StepIDsPath(),
		steplibindex.StepIDs{StepIDs: []string{"hello-step"}})

	latestPath, err := steplibindex.LatestPointerPath("hello-step")
	require.NoError(t, err)
	writeInventoryJSON(t, root, latestPath, steplibindex.LatestPointer{
		StepID:        "hello-step",
		Latest:        "2.0.0",
		LatestByMajor: map[string]string{"1": "1.1.0", "2": "2.0.0"},
	})

	versionsPath, err := steplibindex.VersionsPath("hello-step")
	require.NoError(t, err)
	writeInventoryJSON(t, root, versionsPath,
		steplibindex.Versions{StepID: "hello-step", Versions: helloStepVersions})

	infoPath, err := steplibindex.StepInfoPath("hello-step")
	require.NoError(t, err)
	writeInventoryJSON(t, root, infoPath,
		steplibindex.StepInfo{Maintainer: "bitrise", AssetURLs: []string{}})

	for _, v := range helloStepVersions {
		stepJSONPath, err := steplibindex.StepJSONPath("hello-step", v)
		require.NoError(t, err)
		title := "Hello Step " + v
		writeInventoryJSON(t, root, stepJSONPath, models.StepModel{
			Title: &title,
			Source: &models.StepSourceModel{
				Git:    "https://github.com/example/hello-step.git",
				Commit: "cccc3333cccc3333cccc3333cccc3333cccc3333",
			},
			Executables: executables,
		})
	}

	srv := httptest.NewServer(http.FileServer(http.Dir(root)))
	t.Cleanup(srv.Close)

	// meta.json declares a single zip download base pointing back at this server.
	// Written after the server starts so we know its URL; http.FileServer reads
	// lazily, so it's served on the next request. Only a zip location is listed
	// (no git) so source activation must resolve against localhost, never a clone.
	writeInventoryJSON(t, root, steplibindex.MetaPath(), steplibindex.Meta{
		FormatVersion: steplibindex.FormatVersion,
		UpdatedAt:     time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
		DownloadLocations: []models.DownloadLocationModel{
			{Type: "zip", Src: srv.URL + "/archive/"},
		},
	})

	// BuildStepSourceDownloadLocations forms "<base><stepID>/<version>/step.zip",
	// so the archive for each version lives under archive/hello-step/<version>/.
	for _, v := range helloStepVersions {
		writeStepSourceZip(t, filepath.Join(root, "archive", "hello-step", v, "step.zip"))
	}

	return srv
}

func writeInventoryJSON(t *testing.T, root string, p steplibindex.Path, v any) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(p.FS()))
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	data, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(full, data, 0o644))
}

// writeStepSourceZip writes a step source archive whose single entry is a marker
// file, so a successful activation can be verified by the marker's presence in
// the destination.
func writeStepSourceZip(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("activated_marker.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("activated from source"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

// errFakeDownload is returned by fakeExecutableFetcher when it is set to fail,
// so the executable-download-failure fallback can be driven deterministically.
var errFakeDownload = errors.New("simulated executable download failure")

// fakeExecutableFetcher is a real httpfetch.Client with only DownloadWithHash
// stubbed: it records the requested executable download and writes a stub file at
// destPath, so the executable-activation path can be exercised without a
// hash-matching served artifact. When downloadErr is set, DownloadWithHash
// returns it instead. Get and Download pass through to the embedded real client,
// which is what the step-source zip download needs — that keeps the download
// location ActivateStep builds under test, rather than stubbing past it.
type fakeExecutableFetcher struct {
	httpfetch.Client

	calledURL   string
	calledHash  string
	downloadErr error
}

var _ httpfetch.Client = (*fakeExecutableFetcher)(nil)

func newFakeExecutableFetcher(t *testing.T) *fakeExecutableFetcher {
	t.Helper()
	return &fakeExecutableFetcher{Client: httpfetch.NewClient(apiTestLogger{t})}
}

func (f *fakeExecutableFetcher) DownloadWithHash(_ context.Context, destPath, url, expectedHash string) error {
	f.calledURL = url
	f.calledHash = expectedHash
	if f.downloadErr != nil {
		return f.downloadErr
	}
	return os.WriteFile(destPath, []byte("stub binary"), 0o644)
}
