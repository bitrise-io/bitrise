package buildcache

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_FromHostCache(t *testing.T) {
	tarball, checksum := releaseTarball(t, "#!/bin/sh\necho host\n")
	var garHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/gar/") {
			garHits.Add(1)
		}
		assert.Equal(t, "sha-256="+checksum, r.Header.Get("Digest"))
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()

	bin, err := testInstaller(srv.URL+"/host", srv.URL+"/gar", checksum).install(context.Background(), t.TempDir())

	require.NoError(t, err)
	assertExecutable(t, bin, "#!/bin/sh\necho host\n")
	assert.Zero(t, garHits.Load())
}

func TestInstall_FallsBackToGAR(t *testing.T) {
	tarball, checksum := releaseTarball(t, "#!/bin/sh\necho gar\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/host/") {
			http.NotFound(w, r)
			return
		}
		assert.Empty(t, r.Header.Get("Digest"))
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()

	bin, err := testInstaller(srv.URL+"/host", srv.URL+"/gar", checksum).install(context.Background(), t.TempDir())

	require.NoError(t, err)
	assertExecutable(t, bin, "#!/bin/sh\necho gar\n")
}

func TestInstall_ChecksumMismatch(t *testing.T) {
	tarball, _ := releaseTarball(t, "tampered")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()
	dir := t.TempDir()

	_, err := testInstaller(srv.URL+"/host", srv.URL+"/gar", strings.Repeat("0", 64)).install(context.Background(), dir)

	require.ErrorContains(t, err, "checksum validation failed")
	assert.NoFileExists(t, filepath.Join(dir, binaryName))
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, binaryName), []byte("existing"), 0o755))

	bin, err := testInstaller("http://127.0.0.1:1/host", "http://127.0.0.1:1/gar", "unused").install(context.Background(), dir)

	require.NoError(t, err)
	assertExecutable(t, bin, "existing")
}

func TestURLs(t *testing.T) {
	gar := "https://artifactregistry.googleapis.com/v1/projects/ip-build-cache-prod/locations/us-central1/repositories/build-cache-cli-releases/files/" +
		"bitrise-build-cache_linux_amd64.tar.gz:" + Version + ":bitrise-build-cache_" + Version + "_linux_amd64.tar.gz:download?alt=media"

	assert.Equal(t, []string{gar}, newInstaller(testLogger(), "").urls("linux_amd64"))
	assert.Equal(t, []string{
		"http://192.168.64.1:59020/build-cache-cli-releases/bitrise-build-cache_" + Version + "_linux_amd64.tar.gz",
		gar,
	}, newInstaller(testLogger(), "http://192.168.64.1:59020/build-cache-cli-releases/").urls("linux_amd64"))
}

func testInstaller(hostCacheURL, garURL, checksum string) installer {
	i := newInstaller(testLogger(), hostCacheURL)
	i.garFilesURL = garURL
	i.checksums = map[string]string{runtime.GOOS + "_" + runtime.GOARCH: checksum}
	return i
}

func releaseTarball(t *testing.T, binary string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range map[string]string{"README.md": "docs", binaryName: binary} {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	sum := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(sum[:])
}

func assertExecutable(t *testing.T, bin, content string) {
	t.Helper()
	got, err := os.ReadFile(bin)
	require.NoError(t, err)
	assert.Equal(t, content, string(got))
	assert.True(t, isExecutable(bin))
}

func testLogger() log.Logger {
	return log.NewLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Writer: io.Discard})
}
