package steplib

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

func sha256Hash(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256-" + hex.EncodeToString(sum[:])
}

func TestBuildDownloadURLs(t *testing.T) {
	tests := []struct {
		name         string
		bases        []string
		executable   models.Executable
		expectedURLs []string
		expectedErr  error
	}{
		{
			name:  "Default list: steplib.bitrise.io first, GCS second",
			bases: precompiledStepsDefaultStorageURLs,
			executable: models.Executable{
				StorageURI: "steps/step1.tar.gz",
			},
			expectedURLs: []string{
				"https://steplib.bitrise.io/steps/step1.tar.gz",
				"https://storage.googleapis.com/bitrise-steplib-storage/steps/step1.tar.gz",
			},
		},
		{
			name:  "Multiple bases",
			bases: []string{"https://a.example.com", "https://b.example.com"},
			executable: models.Executable{
				StorageURI: "steps/step2.tar.gz",
			},
			expectedURLs: []string{
				"https://a.example.com/steps/step2.tar.gz",
				"https://b.example.com/steps/step2.tar.gz",
			},
		},
		{
			name:  "URL normalization: trailing slashes and leading StorageURI slash",
			bases: []string{"https://a.example.com/// ", " https://b.example.com///"},
			executable: models.Executable{
				StorageURI: "/steps/step3.tar.gz",
			},
			expectedURLs: []string{
				"https://a.example.com/steps/step3.tar.gz",
				"https://b.example.com/steps/step3.tar.gz",
			},
		},
		{
			name:  "Input parsing: spaces and empty entries",
			bases: []string{"", " https://a.example.com ", "", " https://b.example.com ", ""},
			executable: models.Executable{
				StorageURI: "steps/step4.tar.gz",
			},
			expectedURLs: []string{
				"https://a.example.com/steps/step4.tar.gz",
				"https://b.example.com/steps/step4.tar.gz",
			},
		},
		{
			name:  "http URL is rejected",
			bases: []string{"http://a.example.com"},
			executable: models.Executable{
				StorageURI: "steps/step5.tar.gz",
			},
			expectedErr: fmt.Errorf("http URL is unsupported, please use https: http://a.example.com/steps/step5.tar.gz"),
		},
		{
			name:  "All-empty list yields a configuration error",
			bases: []string{"", "", ""},
			executable: models.Executable{
				StorageURI: "steps/step6.tar.gz",
			},
			expectedErr: fmt.Errorf("no storage URLs configured"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildDownloadURLs(tt.bases, tt.executable)
			if tt.expectedErr != nil {
				require.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedURLs, got)
			}
		})
	}
}

func TestDownloadFromURLs(t *testing.T) {
	ctx := context.Background()
	logger := log.NewDefaultLogger(false)
	fetcher := httpfetch.NewClient(logger)
	payload := []byte("primary payload")
	hash := sha256Hash(payload)

	t.Run("primary succeeds, secondary is not called", func(t *testing.T) {
		secondaryHits := 0
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secondaryHits++
		}))
		defer secondary.Close()

		destPath := filepath.Join(t.TempDir(), "executable")
		err := downloadFromURLs(ctx, fetcher, []string{primary.URL, secondary.URL}, destPath, hash, logger)
		require.NoError(t, err)
		require.Equal(t, 0, secondaryHits)

		got, err := os.ReadFile(destPath)
		require.NoError(t, err)
		require.Equal(t, payload, got)
	})

	t.Run("primary 404 falls back to secondary", func(t *testing.T) {
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer secondary.Close()

		destPath := filepath.Join(t.TempDir(), "executable")
		err := downloadFromURLs(ctx, fetcher, []string{primary.URL, secondary.URL}, destPath, hash, logger)
		require.NoError(t, err)

		got, err := os.ReadFile(destPath)
		require.NoError(t, err)
		require.Equal(t, payload, got)
	})

	t.Run("primary hash mismatch falls back to secondary", func(t *testing.T) {
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("corrupted"))
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer secondary.Close()

		destPath := filepath.Join(t.TempDir(), "executable")
		err := downloadFromURLs(ctx, fetcher, []string{primary.URL, secondary.URL}, destPath, hash, logger)
		require.NoError(t, err)

		got, err := os.ReadFile(destPath)
		require.NoError(t, err)
		require.Equal(t, payload, got)
	})

	t.Run("all URLs fail and the error lists each one", func(t *testing.T) {
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer secondary.Close()

		destPath := filepath.Join(t.TempDir(), "executable")
		err := downloadFromURLs(ctx, fetcher, []string{primary.URL, secondary.URL}, destPath, hash, logger)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to download executable")
		require.Contains(t, err.Error(), primary.URL)
		require.Contains(t, err.Error(), "404")
		require.Contains(t, err.Error(), secondary.URL)
		require.Contains(t, err.Error(), "403")
	})

	t.Run("hash mismatch on every URL is a final error", func(t *testing.T) {
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("corrupted-1"))
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("corrupted-2"))
		}))
		defer secondary.Close()

		destPath := filepath.Join(t.TempDir(), "executable")
		err := downloadFromURLs(ctx, fetcher, []string{primary.URL, secondary.URL}, destPath, hash, logger)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to download executable")
		require.Contains(t, err.Error(), "hash mismatch")
	})
}
