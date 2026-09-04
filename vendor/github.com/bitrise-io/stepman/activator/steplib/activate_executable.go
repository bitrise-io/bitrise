package steplib

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	stepID, version, platform string,
	executable models.Executable,
	logger stepman.Logger,
) (string, error) {
	cachePath, err := stepExecutableCachePath(stepID, version, platform)
	if err != nil {
		return "", fmt.Errorf("executable cache path: %w", err)
	}

	switch err := validateHash(cachePath, executable.Hash); {
	case err == nil:
		return cachePath, nil
	case errors.Is(err, fs.ErrNotExist):
		// cache miss, fall through to download
	default:
		logger.Warnf("Cached step executable failed validation, re-downloading: %s", err)
		if rmErr := os.Remove(cachePath); rmErr != nil && !errors.Is(rmErr, fs.ErrNotExist) {
			logger.Warnf("Failed to remove invalid cache entry %s: %s", cachePath, rmErr)
		}
	}

	if err := downloadExecutable(ctx, fetcher, executable, cachePath, logger); err != nil {
		return "", err
	}
	if err := os.Chmod(cachePath, 0755); err != nil {
		return "", fmt.Errorf("set executable permission on file: %s", err)
	}

	return cachePath, nil
}

func stepExecutableCachePath(stepID, version, platform string) (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(userCacheDir, "bitrise", "steps", "executables")
	return filepath.Join(base, stepID, version, platform, stepID), nil
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

	expectedSha256, err := parseExpectedHash(executable.Hash)
	if err != nil {
		return fmt.Errorf("parse expected hash: %w", err)
	}

	return downloadFromURLs(ctx, fetcher, urls, destPath, expectedSha256, logger)
}

// downloadFromURLs tries each URL in order via fetcher, verifying the expected
// hash on each attempt; a mismatch or failure falls through to the next mirror,
// logging each failed attempt so a mirror silently degrading isn't invisible on
// fallback success.
func downloadFromURLs(ctx context.Context, fetcher httpfetch.Client, urls []string, destPath, expectedSHA256 string, logger stepman.Logger) error {
	var errs []error
	for _, url := range urls {
		err := fetcher.DownloadWithHash(ctx, destPath, url, expectedSHA256)
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

func parseExpectedHash(hash string) (string, error) {
	if hash == "" {
		return "", fmt.Errorf("hash is empty")
	}
	if !strings.HasPrefix(hash, "sha256-") {
		return "", fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: %s", hash)
	}
	return strings.TrimPrefix(hash, "sha256-"), nil
}

func validateHash(filePath string, expectedHash string) error {
	hexHash, err := parseExpectedHash(expectedHash)
	if err != nil {
		return err
	}

	reader, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	h := sha256.New()
	_, err = io.Copy(h, reader)
	if err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}
	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != hexHash {
		return fmt.Errorf("hash mismatch: expected sha256-%s, got sha256-%s", hexHash, actualHash)
	}
	return nil
}
