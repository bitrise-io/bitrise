package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleaseAsset(t *testing.T) {
	expected := map[string]string{
		"darwin/amd64": "bitrise-Darwin-x86_64",
		"darwin/arm64": "bitrise-Darwin-arm64",
		"linux/amd64":  "bitrise-Linux-x86_64",
		"linux/arm64":  "bitrise-Linux-arm64",
	}[runtime.GOOS+"/"+runtime.GOARCH]
	require.NotEmpty(t, expected, "no published asset for the platform running the tests")

	asset, err := releaseAsset()

	require.NoError(t, err)
	require.Equal(t, expected, asset)
}

func TestDownloaderInstall(t *testing.T) {
	t.Run("the asset of the running platform replaces the target", func(t *testing.T) {
		var requested string
		downloader := testDownloader(t, func(w http.ResponseWriter, r *http.Request) {
			requested = r.URL.Path
			_, _ = w.Write([]byte("new binary"))
		})
		target := existingBinary(t, "old binary", 0750)

		require.NoError(t, downloader.Install(t.Context(), "2.46.0", target))

		asset, err := releaseAsset()
		require.NoError(t, err)
		require.Equal(t, "/v2.46.0/"+asset, requested)
		requireFile(t, target, "new binary", 0750)
	})

	t.Run("a target that does not exist yet is executable", func(t *testing.T) {
		downloader := testDownloader(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("new binary"))
		})
		target := filepath.Join(t.TempDir(), "bitrise")

		require.NoError(t, downloader.Install(t.Context(), "2.46.0", target))

		requireFile(t, target, "new binary", 0755)
	})

	t.Run("a missing asset keeps the installed binary", func(t *testing.T) {
		downloader := testDownloader(t, func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		})
		target := existingBinary(t, "old binary", 0755)

		err := downloader.Install(t.Context(), "9.9.9", target)

		require.ErrorContains(t, err, "Not Found")
		requireUntouched(t, target, "old binary")
	})

	t.Run("an interrupted download keeps the installed binary", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		downloader := testDownloader(t, func(_ http.ResponseWriter, r *http.Request) {
			cancel()
			<-r.Context().Done()
		})
		target := existingBinary(t, "old binary", 0755)

		err := downloader.Install(ctx, "2.46.0", target)

		require.Error(t, err)
		requireUntouched(t, target, "old binary")
	})
}

func testDownloader(t *testing.T, handler http.HandlerFunc) *Downloader {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	downloader := NewDownloader()
	downloader.downloadURL = server.URL
	return downloader
}

func existingBinary(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bitrise")
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
	return path
}

// A failed update has to leave both the binary and its directory as they were.
func requireUntouched(t *testing.T, path, content string) {
	t.Helper()
	requireFile(t, path, content, 0755)
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func requireFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	actual, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, string(actual))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, mode, info.Mode().Perm())
}
