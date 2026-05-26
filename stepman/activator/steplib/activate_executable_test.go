package steplib

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedHash string
		expectedErr  error
	}{
		{
			name:         "Valid hash",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
			expectedErr:  nil,
		},
		{
			name:         "Hash mismatch",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-1234567890abcdef",
			expectedErr:  fmt.Errorf("hash mismatch: expected sha256-1234567890abcdef, got sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2"),
		},
		{
			name:         "Nonexistent file",
			filePath:     "testdata/nonexistent.txt",
			expectedHash: "sha256-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("open testdata/nonexistent.txt: no such file or directory"),
		},
		{
			name:         "Empty hash",
			filePath:     "testdata/file.txt",
			expectedHash: "",
			expectedErr:  fmt.Errorf("hash is empty"),
		},
		{
			name:         "Invalid hash type",
			filePath:     "testdata/file.txt",
			expectedHash: "md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHash(tt.filePath, tt.expectedHash)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
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
			name:  "Default list: GCS first, gateway second",
			bases: precompiledStepsDefaultStorageURLs,
			executable: models.Executable{
				StorageURI: "steps/step1.tar.gz",
			},
			expectedURLs: []string{
				"https://storage.googleapis.com/bitrise-steplib-storage/steps/step1.tar.gz",
				"https://storage-gateway.services.bitrise.io/steps/step1.tar.gz",
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
	t.Run("primary succeeds, secondary is not called", func(t *testing.T) {
		secondaryHits := 0
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("from primary"))
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secondaryHits++
		}))
		defer secondary.Close()

		body, err := downloadFromURLs([]string{primary.URL, secondary.URL})
		require.NoError(t, err)
		defer func() { _ = body.Close() }()

		b, err := io.ReadAll(body)
		require.NoError(t, err)
		require.Equal(t, "from primary", string(b))
		require.Equal(t, 0, secondaryHits)
	})

	t.Run("primary 404 falls back to secondary", func(t *testing.T) {
		primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer primary.Close()
		secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("from secondary"))
		}))
		defer secondary.Close()

		body, err := downloadFromURLs([]string{primary.URL, secondary.URL})
		require.NoError(t, err)
		defer func() { _ = body.Close() }()

		b, err := io.ReadAll(body)
		require.NoError(t, err)
		require.Equal(t, "from secondary", string(b))
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

		_, err := downloadFromURLs([]string{primary.URL, secondary.URL})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to download executable")
		require.Contains(t, err.Error(), primary.URL)
		require.Contains(t, err.Error(), "status 404")
		require.Contains(t, err.Error(), secondary.URL)
		require.Contains(t, err.Error(), "status 403")
	})
}
