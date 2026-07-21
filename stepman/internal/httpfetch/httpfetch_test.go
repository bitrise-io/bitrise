package httpfetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/require"
)

func sha256Hash(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256-" + hex.EncodeToString(sum[:])
}

func newClient() Client {
	return NewClient(log.NewDefaultLogger(false))
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	fetcher := newClient()

	t.Run("2xx returns the body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello"))
		}))
		defer server.Close()

		body, err := fetcher.Get(ctx, server.URL)
		require.NoError(t, err)
		defer func() { _ = body.Close() }()

		got, err := io.ReadAll(body)
		require.NoError(t, err)
		require.Equal(t, "hello", string(got))
	})

	t.Run("non-2xx returns a StatusError with the status code and body snippet", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		defer server.Close()

		_, err := fetcher.Get(ctx, server.URL)
		require.Error(t, err)

		var statusErr *StatusError
		require.True(t, errors.As(err, &statusErr), "expected a *StatusError, got %T: %v", err, err)
		require.Equal(t, http.StatusNotFound, statusErr.Code)
		require.Equal(t, "not found", statusErr.Body)
	})
}

func TestDownload(t *testing.T) {
	ctx := context.Background()
	fetcher := newClient()
	payload := []byte("binary content")

	t.Run("writes the fetched content to destPath", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer server.Close()

		destPath := filepath.Join(t.TempDir(), "nested", "dir", "out.bin")
		err := fetcher.Download(ctx, destPath, server.URL)
		require.NoError(t, err)

		got, err := os.ReadFile(destPath)
		require.NoError(t, err)
		require.Equal(t, payload, got)
	})

	t.Run("a failed fetch leaves no file at destPath", func(t *testing.T) {
		// 404 (not 500) so retryablehttp's internal retry policy doesn't retry
		// this and slow the test down: only 5xx/network errors are retried.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		destPath := filepath.Join(t.TempDir(), "out.bin")
		err := fetcher.Download(ctx, destPath, server.URL)
		require.Error(t, err)

		_, statErr := os.Stat(destPath)
		require.True(t, os.IsNotExist(statErr), "destPath should not exist after a failed download")

		// No leftover temp files in the destination directory either.
		entries, err := os.ReadDir(filepath.Dir(destPath))
		require.NoError(t, err)
		require.Empty(t, entries, "temp file should have been cleaned up")
	})
}

func TestDownloadWithHash(t *testing.T) {
	ctx := context.Background()
	fetcher := newClient()
	payload := []byte("binary content")
	hash := sha256Hash(payload)

	t.Run("matching hash succeeds and writes destPath", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer server.Close()

		destPath := filepath.Join(t.TempDir(), "out.bin")
		err := fetcher.DownloadWithHash(ctx, destPath, server.URL, hash)
		require.NoError(t, err)

		got, err := os.ReadFile(destPath)
		require.NoError(t, err)
		require.Equal(t, payload, got)
	})

	t.Run("mismatched hash fails and leaves no file at destPath", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("different content"))
		}))
		defer server.Close()

		destPath := filepath.Join(t.TempDir(), "out.bin")
		err := fetcher.DownloadWithHash(ctx, destPath, server.URL, hash)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash mismatch")

		_, statErr := os.Stat(destPath)
		require.True(t, os.IsNotExist(statErr), "destPath should not exist after a hash mismatch")
	})

	t.Run("empty expected hash is rejected without making a request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("server should not have been contacted")
		}))
		defer server.Close()

		destPath := filepath.Join(t.TempDir(), "out.bin")
		err := fetcher.DownloadWithHash(ctx, destPath, server.URL, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash is empty")
	})
}
