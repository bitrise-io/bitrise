package stepman

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

type stubLogger struct{ t *testing.T }

func (l stubLogger) Debugf(format string, v ...any) { l.t.Logf(format, v...) }
func (l stubLogger) Errorf(format string, v ...any) { l.t.Logf(format, v...) }
func (l stubLogger) Warnf(format string, v ...any)  { l.t.Logf(format, v...) }
func (l stubLogger) Infof(format string, v ...any)  { l.t.Logf(format, v...) }

// zipOf builds an in-memory zip of files, then appends an empty entry for each
// name in trailingNames. The trailing entries are ordered after the map's, so a
// test can place a deliberately bad entry behind a good one.
func zipOf(t *testing.T, files map[string]string, trailingNames ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	write := func(name, content string) {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	for name, content := range files {
		write(name, content)
	}
	for _, name := range trailingNames {
		write(name, "")
	}
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestDownloadStepZip(t *testing.T) {
	payload := zipOf(t, map[string]string{"step.yml": "title: Test\n"})
	fetcher := httpfetch.NewClient(stubLogger{t})

	// downloadStepZip is exercised against a real httptest server (rather than a
	// fake fetcher) so the zip-download-and-extract path is covered end to end.
	t.Run("downloads and extracts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer server.Close()

		destDir := t.TempDir()
		err := downloadStepZip(fetcher, server.URL, destDir, stubLogger{t})
		require.NoError(t, err)

		got, err := os.ReadFile(filepath.Join(destDir, "step.yml"))
		require.NoError(t, err)
		require.Equal(t, "title: Test\n", string(got))
	})

	t.Run("propagates HTTP failures", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		err := downloadStepZip(fetcher, server.URL, t.TempDir(), stubLogger{t})
		require.Error(t, err)
		require.Contains(t, err.Error(), "download step zip")
	})
}

// TestDownloadStep covers the cache-dir contract: the cache is keyed by the
// existence of the step's cache dir alone, so the dir must appear only once it is
// fully populated. A failed download must leave no cache dir behind for the next
// run to serve as a hit.
func TestDownloadStep(t *testing.T) {
	const (
		stepID      = "hello-step"
		stepVersion = "1.0.0"
		steplibURI  = "https://github.com/example/steplib.git"
	)

	// setup points HOME at a temp dir and registers a route, so the cache lands
	// under the temp dir rather than the developer's real ~/.stepman. It returns the
	// step's cache dir and its parent (where the staging dir is created).
	setup := func(t *testing.T) (stepCacheDir, cacheParentDir string) {
		t.Helper()
		t.Setenv("HOME", t.TempDir())
		require.NoError(t, CreateStepManDirIfNeeded())

		route := SteplibRoute{SteplibURI: steplibURI, FolderAlias: "test-alias"}
		require.NoError(t, SteplibRoutes{route}.writeToFile())
		// GetRoute only resolves a route whose collection dir exists on disk.
		require.NoError(t, os.MkdirAll(filepath.Join(GetCollectionsDirPath(), route.FolderAlias), 0o755))

		stepCacheDir = GetStepCacheDirPath(route, stepID, stepVersion)
		return stepCacheDir, filepath.Dir(stepCacheDir)
	}

	collectionFor := func(zipBaseURL string) models.StepCollectionModel {
		return models.StepCollectionModel{
			SteplibSource:     steplibURI,
			DownloadLocations: []models.DownloadLocationModel{{Type: "zip", Src: zipBaseURL}},
			Steps: models.StepHash{
				stepID: models.StepGroupModel{
					Versions: map[string]models.StepModel{
						stepVersion: {Source: &models.StepSourceModel{Git: "https://github.com/example/hello-step.git"}},
					},
					LatestVersionNumber: stepVersion,
				},
			},
		}
	}

	// serve responds to every request with payload, or with status when non-zero.
	serve := func(t *testing.T, payload []byte, status int) string {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if status != 0 {
				w.WriteHeader(status)
				return
			}
			_, _ = w.Write(payload)
		}))
		t.Cleanup(server.Close)
		return server.URL + "/"
	}

	fetcher := httpfetch.NewClient(stubLogger{t})

	t.Run("populates the cache dir and leaves no staging dir behind", func(t *testing.T) {
		stepCacheDir, cacheParentDir := setup(t)
		baseURL := serve(t, zipOf(t, map[string]string{"step.yml": "title: Test\n"}), 0)

		require.NoError(t, DownloadStep(steplibURI, collectionFor(baseURL), stepID, stepVersion, "", stubLogger{t}, fetcher))

		got, err := os.ReadFile(filepath.Join(stepCacheDir, "step.yml"))
		require.NoError(t, err)
		require.Equal(t, "title: Test\n", string(got))

		entries, err := os.ReadDir(cacheParentDir)
		require.NoError(t, err)
		require.Len(t, entries, 1, "only the version dir should remain, got %v", entries)
		require.Equal(t, stepVersion, entries[0].Name())
	})

	t.Run("leaves no cache dir when the download fails", func(t *testing.T) {
		stepCacheDir, cacheParentDir := setup(t)
		baseURL := serve(t, nil, http.StatusNotFound)

		err := DownloadStep(steplibURI, collectionFor(baseURL), stepID, stepVersion, "", stubLogger{t}, fetcher)
		require.Error(t, err)

		require.NoDirExists(t, stepCacheDir)
		entries, err := os.ReadDir(cacheParentDir)
		require.NoError(t, err)
		require.Empty(t, entries, "staging dir must be cleaned up, got %v", entries)
	})

	// A zip that fails partway through extraction is the case the staging dir
	// exists for: without it the cache dir would survive half-populated and be
	// served as a valid hit on the next run.
	t.Run("leaves no cache dir when extraction fails partway", func(t *testing.T) {
		stepCacheDir, cacheParentDir := setup(t)
		// The first entry extracts fine; the second escapes destDir and is rejected.
		baseURL := serve(t, zipOf(t, map[string]string{"step.yml": "title: Test\n"}, "../escaped.txt"), 0)

		err := DownloadStep(steplibURI, collectionFor(baseURL), stepID, stepVersion, "", stubLogger{t}, fetcher)
		require.Error(t, err)

		require.NoDirExists(t, stepCacheDir)
		entries, err := os.ReadDir(cacheParentDir)
		require.NoError(t, err)
		require.Empty(t, entries, "staging dir must be cleaned up, got %v", entries)
	})
}

func TestAddStepVersionToStepGroup(t *testing.T) {
	step := models.StepModel{
		Title: pointers.NewStringPtr("name 1"),
	}

	group := models.StepGroupModel{
		Versions: map[string]models.StepModel{
			"1.0.0": step,
			"2.0.0": step,
		},
		LatestVersionNumber: "2.0.0",
	}

	group, err := addStepVersionToStepGroup(step, "2.1.0", group)
	require.Equal(t, nil, err)
	require.Equal(t, 3, len(group.Versions))
	require.Equal(t, "2.1.0", group.LatestVersionNumber)
}

func Test_parseStepModel(t *testing.T) {
	empty := ""
	falseBool := false
	zero := 0
	tests := []struct {
		name     string
		bytes    []byte
		validate bool
		want     models.StepModel
		wantErr  bool
	}{
		{
			name:     "Meta field",
			bytes:    []byte(stepDefinitionMetaFieldOnly),
			validate: false,
			want: models.StepModel{
				Title: &empty,
				Meta: map[string]interface{}{
					"bitrise.io.addons.optional.2": []interface{}{
						map[string]interface{}{
							"addon_id": "addons-testing",
						},
					},
					"bitrise.io.addons.required": []interface{}{
						map[string]interface{}{
							"addon_id": "addons-testing",
							"addon_options": map[string]interface{}{
								"required": true,
								"title":    "Testing Addon",
							},
							"addon_params": "--token TOKEN",
						},
						map[string]interface{}{
							"addon_id": "addons-ship",
							"addon_options": map[string]interface{}{
								"required": true,
								"title":    "Ship Addon",
							},
							"addon_params": "--token TOKEN",
						},
					},
				},
				Summary:             &empty,
				Description:         &empty,
				Website:             &empty,
				SourceCodeURL:       &empty,
				SupportURL:          &empty,
				IsRequiresAdminUser: &falseBool,
				IsAlwaysRun:         &falseBool,
				IsSkippable:         &falseBool,
				RunIf:               &empty,
				Timeout:             &zero,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStepModel(tt.bytes, tt.validate)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStepModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStepModel() = %+v, want %v,\n Diff: %s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

const stepDefinitionMetaFieldOnly = `
meta:
  bitrise.io.addons.required: 
    - addon_id: "addons-testing"
      addon_params: "--token TOKEN"
      addon_options: 
        required: true
        title: "Testing Addon"
    - addon_id: "addons-ship"
      addon_params: "--token TOKEN"
      addon_options: 
        required: true
        title: "Ship Addon"
  bitrise.io.addons.optional.2: [{"addon_id":"addons-testing"}]
`
