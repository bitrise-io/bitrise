package steplibrary

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/models"
)

// currentPlatform returns the runtime platform key (e.g. "darwin-arm64") used
// to look up an entry in models.StepModel.Executables.
func currentPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// resolveExecutable picks the binary for the current OS+arch from the step
// model. Returns false when no precompiled binary is available, when the
// platform isn't covered, or when the entry is missing storage_uri/hash.
func resolveExecutable(step models.StepModel) (models.Executable, bool) {
	if step.Executables == nil {
		return models.Executable{}, false
	}
	e, ok := (*step.Executables)[currentPlatform()]
	if !ok || e.StorageURI == "" || e.Hash == "" {
		return models.Executable{}, false
	}
	return e, true
}

// precompiledURLs builds the ordered list of download URLs for an executable.
// Bases come from BITRISE_PRECOMPILED_STEPS_STORAGE_URLS (comma-separated) or
// the two built-in defaults (GCS bucket + storage gateway).
func precompiledURLs(e models.Executable) ([]string, error) {
	bases := steplib.PrecompiledStepsDefaultStorageURLs
	if override := os.Getenv(steplib.PrecompiledStepsStorageURLsEnv); override != "" {
		bases = strings.Split(override, ",")
	}

	uri := strings.TrimLeft(e.StorageURI, "/")
	var urls []string
	for _, base := range bases {
		base = strings.TrimRight(strings.TrimSpace(base), "/")
		if base == "" {
			continue
		}
		url := fmt.Sprintf("%s/%s", base, uri)
		if strings.HasPrefix(url, "http://") {
			return nil, fmt.Errorf("http URL is unsupported, please use https: %s", url)
		}
		urls = append(urls, url)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no storage URLs configured")
	}
	return urls, nil
}

// downloadFromURLs tries each url in order, verifying expectedHash on each attempt.
func (s *Steplib) downloadFromURLs(ctx context.Context, destPath, expectedHash string, urls []string) error {
	var errs []error
	for _, url := range urls {
		if err := s.fetcher.DownloadWithHash(ctx, destPath, url, expectedHash); err == nil {
			return nil
		} else {
			s.log.Warnf("Failed to download from %s: %s\n", url, err)
			errs = append(errs, fmt.Errorf("%s: %w", url, err))
		}
	}
	return fmt.Errorf("failed to download executable: %w", errors.Join(errs...))
}

// downloadPrecompiled fetches `executable` for the current platform, verifies
// its SHA256, makes the file executable, and places it at destDir/<stepID>.
// Returns the final binary path.
func (s *Steplib) downloadPrecompiled(ctx context.Context, stepID string, executable models.Executable, destDir string) (binPath string, err error) {
	if executable.Hash == "" {
		return "", fmt.Errorf("hash is empty")
	}
	urls, err := precompiledURLs(executable)
	if err != nil {
		return "", err
	}

	binPath = filepath.Join(destDir, stepID)
	if err = s.downloadFromURLs(ctx, binPath, executable.Hash, urls); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(binPath))
		}
	}()

	if err = os.Chmod(binPath, 0o755); err != nil {
		return "", fmt.Errorf("chmod %s: %w", binPath, err)
	}
	return binPath, nil
}
