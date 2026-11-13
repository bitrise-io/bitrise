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
	"strings"
	"sync"
	"sync/atomic"
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

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func redirectCacheDir(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")
}

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
			expectedHash: "f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
			expectedErr:  nil,
		},
		{
			name:         "Hash mismatch",
			filePath:     "testdata/file.txt",
			expectedHash: "1234567890abcdef",
			expectedErr:  fmt.Errorf("hash mismatch: expected sha256-1234567890abcdef, got sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2"),
		},
		{
			name:         "Nonexistent file",
			filePath:     "testdata/nonexistent.txt",
			expectedHash: "3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("open testdata/nonexistent.txt: no such file or directory"),
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

func TestParseExpectedHash(t *testing.T) {
	tests := []struct {
		name        string
		hash        string
		expectedHex string
		expectedErr error
	}{
		{
			name:        "Valid hash",
			hash:        "sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
			expectedHex: "f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
		},
		{
			name:        "Empty hash",
			hash:        "",
			expectedErr: fmt.Errorf("hash is empty"),
		},
		{
			name:        "Invalid hash type",
			hash:        "md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr: fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexHash, err := parseExpectedHash(tt.hash)
			if tt.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, tt.expectedHex, hexHash)
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
			name:  "Default list: steplib.bitrise.io first, GCS second",
			bases: DefaultPrecompiledStorageURLs,
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

func TestActivateStepExecutable(t *testing.T) {
	ctx := context.Background()
	logger := log.NewDefaultLogger(false)
	storageURI := "steps/hello-step/2.0.0/bin/hello-step"
	const hash = "sha256-1111111111111111111111111111111111111111111111111111111111111111"

	t.Run("default bases: first mirror + StorageURI, hash threaded through", func(t *testing.T) {
		redirectCacheDir(t)
		fake := newFakeExecutableFetcher(t)
		destDir := t.TempDir()

		path, err := activateStepExecutable(ctx, fake, "https://github.com/bitrise-io/bitrise-steplib.git", "hello-step", "2.0.0", "linux-amd64",
			models.Executable{StorageURI: storageURI, Hash: hash}, destDir, logger, DefaultPrecompiledStorageURLs)
		require.NoError(t, err)

		require.Equal(t, filepath.Join(destDir, "hello-step"), path, "the served path must live under the caller's destination dir, not the cache")
		require.FileExists(t, path)
		require.Equal(t, DefaultPrecompiledStorageURLs[0]+"/"+storageURI, fake.calledURL)
		hexHash, err := parseExpectedHash(hash)
		require.NoError(t, err)
		require.Equal(t, hexHash, fake.calledHash, "the fetcher sees a bare hex digest, not the \"sha256-\" tagged form")
	})

	t.Run("configured storage URLs win over the defaults", func(t *testing.T) {
		redirectCacheDir(t)
		fake := newFakeExecutableFetcher(t)

		_, err := activateStepExecutable(ctx, fake, "https://github.com/bitrise-io/bitrise-steplib.git", "hello-step", "2.0.0", "linux-amd64",
			models.Executable{StorageURI: storageURI, Hash: hash}, t.TempDir(), logger,
			[]string{"https://custom.example.com"})
		require.NoError(t, err)

		require.Equal(t, "https://custom.example.com/"+storageURI, fake.calledURL)
	})
}

// newExecutableTestServer spins up a self-signed TLS test server (the download
// path enforces https) and returns a real httpfetch.Client configured to trust
// its certificate, plus a storageURLs slice pointing at it - so
// activateStepExecutable's real download path can be exercised without
// depending on OS-specific system cert trust behavior.
func newExecutableTestServer(t *testing.T, handler http.HandlerFunc) (httpfetch.Client, []string) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return httpfetch.NewWithClient(server.Client()), []string{server.URL}
}

func TestActivateStepExecutableCache(t *testing.T) {
	ctx := context.Background()
	logger := log.NewDefaultLogger(false)

	t.Run("cache miss downloads and populates the cache", func(t *testing.T) {
		redirectCacheDir(t)
		var hits int32
		content := []byte("step binary contents v1")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			_, _ = w.Write(content)
		})

		executable := models.Executable{StorageURI: "steps/step1.bin", Hash: sha256Hash(content)}
		destDir := t.TempDir()
		path, err := activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step1", "1.0.0", "linux-amd64", executable, destDir, logger, storageURLs)
		require.NoError(t, err)
		require.EqualValues(t, 1, atomic.LoadInt32(&hits))
		require.Equal(t, filepath.Join(destDir, "step1"), path)

		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, content, got)

		cachePath, err := stepExecutableCachePath("https://github.com/bitrise-io/bitrise-steplib.git", "step1", "1.0.0", "linux-amd64")
		require.NoError(t, err)
		require.FileExists(t, cachePath, "a successful download must populate the shared cache")
	})

	t.Run("cache hit skips the download", func(t *testing.T) {
		redirectCacheDir(t)
		var hits int32
		content := []byte("step binary contents v2")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			_, _ = w.Write(content)
		})

		executable := models.Executable{StorageURI: "steps/step2.bin", Hash: sha256Hash(content)}
		firstPath, err := activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step2", "1.0.0", "linux-amd64", executable, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)
		require.EqualValues(t, 1, atomic.LoadInt32(&hits))

		// A second, independent destination dir: the cache is shared, but each
		// activation still gets its own served copy.
		secondPath, err := activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step2", "1.0.0", "linux-amd64", executable, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)
		require.EqualValues(t, 1, atomic.LoadInt32(&hits), "second activation must not hit the network")
		require.NotEqual(t, firstPath, secondPath, "each activation gets its own destination copy")

		firstContent, err := os.ReadFile(firstPath)
		require.NoError(t, err)
		secondContent, err := os.ReadFile(secondPath)
		require.NoError(t, err)
		require.Equal(t, firstContent, secondContent)
	})

	t.Run("corrupt cache entry is detected and re-downloaded", func(t *testing.T) {
		redirectCacheDir(t)
		var hits int32
		content := []byte("step binary contents v3")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			_, _ = w.Write(content)
		})

		executable := models.Executable{StorageURI: "steps/step3.bin", Hash: sha256Hash(content)}
		cachePath, err := stepExecutableCachePath("https://github.com/bitrise-io/bitrise-steplib.git", "step3", "1.0.0", "linux-amd64")
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0755))
		require.NoError(t, os.WriteFile(cachePath, []byte("corrupted"), 0644))

		path, err := activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step3", "1.0.0", "linux-amd64", executable, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)
		require.EqualValues(t, 1, atomic.LoadInt32(&hits))

		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, content, got)
	})

	t.Run("cache entry with valid content but wrong mode is repaired, not just trusted", func(t *testing.T) {
		redirectCacheDir(t)
		content := []byte("step binary contents v3b")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("must not re-download when the cached content is already valid")
		})

		executable := models.Executable{StorageURI: "steps/step3b.bin", Hash: sha256Hash(content)}
		cachePath, err := stepExecutableCachePath("https://github.com/bitrise-io/bitrise-steplib.git", "step3b", "1.0.0", "linux-amd64")
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0755))
		// Simulate a cache hit landing between the download's atomic publish
		// and the mode fixup: correct content, wrong (non-executable) mode.
		require.NoError(t, os.WriteFile(cachePath, content, 0600))

		path, err := activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step3b", "1.0.0", "linux-amd64", executable, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)

		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0755), info.Mode().Perm(), "cache hit must repair the executable mode")
	})

	t.Run("different steplib sources for the same step ID/version/platform get separate cache entries", func(t *testing.T) {
		redirectCacheDir(t)
		contentA := []byte("library A's binary")
		contentB := []byte("library B's binary")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "a-only") {
				_, _ = w.Write(contentA)
				return
			}
			_, _ = w.Write(contentB)
		})

		executableA := models.Executable{StorageURI: "steps/a-only.bin", Hash: sha256Hash(contentA)}
		executableB := models.Executable{StorageURI: "steps/b-only.bin", Hash: sha256Hash(contentB)}

		pathA, err := activateStepExecutable(ctx, fetcher, "https://github.com/example/library-a.git", "shared-step", "1.0.0", "linux-amd64", executableA, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)
		pathB, err := activateStepExecutable(ctx, fetcher, "https://github.com/example/library-b.git", "shared-step", "1.0.0", "linux-amd64", executableB, t.TempDir(), logger, storageURLs)
		require.NoError(t, err)

		cachePathA, err := stepExecutableCachePath("https://github.com/example/library-a.git", "shared-step", "1.0.0", "linux-amd64")
		require.NoError(t, err)
		cachePathB, err := stepExecutableCachePath("https://github.com/example/library-b.git", "shared-step", "1.0.0", "linux-amd64")
		require.NoError(t, err)
		require.NotEqual(t, cachePathA, cachePathB, "different libraries must not share a cache path for the same step ID/version/platform")

		gotA, err := os.ReadFile(pathA)
		require.NoError(t, err)
		require.Equal(t, contentA, gotA)

		gotB, err := os.ReadFile(pathB)
		require.NoError(t, err)
		require.Equal(t, contentB, gotB)
	})

	t.Run("concurrent activations for the same key do not corrupt the cache", func(t *testing.T) {
		redirectCacheDir(t)
		content := []byte("step binary contents v4")
		fetcher, storageURLs := newExecutableTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(content)
		})

		executable := models.Executable{StorageURI: "steps/step4.bin", Hash: sha256Hash(content)}
		const n = 10
		destDirs := make([]string, n)
		for i := range n {
			destDirs[i] = t.TempDir()
		}
		var wg sync.WaitGroup
		paths := make([]string, n)
		errs := make([]error, n)
		for i := range n {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				paths[i], errs[i] = activateStepExecutable(ctx, fetcher, "https://github.com/bitrise-io/bitrise-steplib.git", "step4", "1.0.0", "linux-amd64", executable, destDirs[i], logger, storageURLs)
			}(i)
		}
		wg.Wait()

		for i := range n {
			require.NoError(t, errs[i])
			got, err := os.ReadFile(paths[i])
			require.NoError(t, err)
			require.Equal(t, content, got)
		}
	})
}

func TestDownloadFromURLs(t *testing.T) {
	ctx := context.Background()
	logger := log.NewDefaultLogger(false)
	fetcher := httpfetch.NewClient(logger)
	payload := []byte("primary payload")
	hash := sha256Hex(payload)

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
