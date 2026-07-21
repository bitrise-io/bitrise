package steplib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

func activateStepExecutable(
	ctx context.Context,
	fetcher httpfetch.Client,
	stepID string,
	executable models.Executable,
	destinationDir string,
	logger stepman.Logger,
) (string, error) {
	path := filepath.Join(destinationDir, stepID)

	if err := downloadExecutable(ctx, fetcher, executable, path, logger); err != nil {
		return "", err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return "", fmt.Errorf("set executable permission on file: %s", err)
	}

	return path, nil
}

func buildDownloadURLs(bases []string, executable models.Executable) ([]string, error) {
	uri := strings.TrimLeft(executable.StorageURI, "/")
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

func downloadExecutable(ctx context.Context, fetcher httpfetch.Client, executable models.Executable, destPath string, logger stepman.Logger) error {
	bases := precompiledStepsDefaultStorageURLs
	if override := os.Getenv(precompiledStepsStorageURLsEnv); override != "" {
		bases = strings.Split(override, ",")
	}

	urls, err := buildDownloadURLs(bases, executable)
	if err != nil {
		return err
	}
	return downloadFromURLs(ctx, fetcher, urls, destPath, executable.Hash, logger)
}

// downloadFromURLs tries each URL in order via fetcher, verifying executable.Hash
// on each attempt; a mismatch or failure falls through to the next mirror, logging
// each failed attempt so a mirror silently degrading isn't invisible on fallback success.
func downloadFromURLs(ctx context.Context, fetcher httpfetch.Client, urls []string, destPath, hash string, logger stepman.Logger) error {
	var errs []error
	for _, url := range urls {
		err := fetcher.DownloadWithHash(ctx, destPath, url, hash)
		if err == nil {
			return nil
		}
		// err already names the failing URL (fetcher wraps it in the underlying
		// GET/status/hash-mismatch error), so it isn't repeated here.
		logger.Warnf("Failed to download step executable: %s", err)
		errs = append(errs, fmt.Errorf("%s: %w", url, err))
	}
	return fmt.Errorf("failed to download executable: %w", errors.Join(errs...))
}
