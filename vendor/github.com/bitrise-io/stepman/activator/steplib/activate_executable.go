package steplib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman/filelock"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func activateStepExecutable(
	stepLibURI string,
	stepID string,
	version string,
	executable models.Executable,
	destinationDir string,
	destinationStepYML string,
) (string, error) {
	if strings.HasPrefix(executable.Url, "http://") {
		return "", fmt.Errorf("http URL is unsupported, please use https: %s", executable.Url)
	}

	finalPath := filepath.Join(destinationDir, stepID)
	lockPath := finalPath + ".download.lock"

	// Acquire lock to prevent concurrent downloads
	lock := filelock.NewFileLock(lockPath)
	if err := lock.TryLock(); err != nil {
		// Another process is downloading, wait and check if file exists
		log.Warnf("Another process is downloading %s, waiting...", stepID)
		time.Sleep(2 * time.Second)
		if exists, _ := pathutil.IsPathExists(finalPath); exists {
			// File was created by other process, verify hash and return
			if hashErr := validateHash(finalPath, executable.Hash); hashErr == nil {
				return finalPath, nil
			}
		}
		// Try to acquire lock with timeout
		if err := lock.Lock(); err != nil {
			return "", fmt.Errorf("failed to acquire download lock: %w", err)
		}
	}
	defer func() { _ = lock.Unlock() }()

	// Check if file was created while waiting for lock
	if exists, _ := pathutil.IsPathExists(finalPath); exists {
		if err := validateHash(finalPath, executable.Hash); err == nil {
			return finalPath, nil
		}
		// File exists but hash is invalid, remove and re-download
		_ = os.Remove(finalPath)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", destinationDir, err)
	}

	// Download to temporary file first
	tempPath := finalPath + fmt.Sprintf(".tmp.%d", os.Getpid())
	defer func() { _ = os.Remove(tempPath) }() // Clean up temp file on any error

	resp, err := retryablehttp.Get(executable.Url)
	if err != nil {
		return "", fmt.Errorf("fetch from %s: %w", executable.Url, err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Warnf("Failed to close response body: %s\n", err)
		}
	}()

	file, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("create temp file %s: %w", tempPath, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("Failed to close temp file %s: %s\n", tempPath, err)
		}
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("download %s to %s: %w", executable.Url, tempPath, err)
	}

	_ = file.Close() // Close before validation

	err = validateHash(tempPath, executable.Hash)
	if err != nil {
		return "", fmt.Errorf("validate hash: %s", err)
	}

	// Make executable before moving
	err = os.Chmod(tempPath, 0755)
	if err != nil {
		return "", fmt.Errorf("set executable permission on temp file: %s", err)
	}

	// Atomic move to final location
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", fmt.Errorf("move temp file to final location: %w", err)
	}

	if err := copyStepYML(stepLibURI, stepID, version, destinationStepYML); err != nil {
		return "", fmt.Errorf("copy step.yml: %s", err)
	}

	return finalPath, nil
}

func validateHash(filePath string, expectedHash string) error {
	if expectedHash == "" {
		return fmt.Errorf("hash is empty")
	}

	if !strings.HasPrefix(expectedHash, "sha256-") {
		return fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: %s", expectedHash)
	}

	expectedHash = strings.TrimPrefix(expectedHash, "sha256-")

	reader, err := os.Open(filePath)
	if err != nil {
		return err
	}

	h := sha256.New()
	_, err = io.Copy(h, reader)
	if err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}
	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected sha256-%s, got sha256-%s", expectedHash, actualHash)
	}
	return nil
}
