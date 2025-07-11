package steplib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/models"
	"github.com/hashicorp/go-retryablehttp"
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

	err = os.MkdirAll(destinationDir, 0755)
	if err != nil {
		return "", fmt.Errorf("create directory %s: %w", destinationDir, err)
	}

	path := filepath.Join(destinationDir, stepID)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file %s: %w", path, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("Failed to close file %s: %s\n", path, err)
		}
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("download %s to %s: %w", executable.Url, path, err)
	}

	err = validateHash(path, executable.Hash)
	if err != nil {
		return "", fmt.Errorf("validate hash: %s", err)
	}

	err = os.Chmod(path, 0755)
	if err != nil {
		return "", fmt.Errorf("set executable permission on file: %s", err)
	}

	if err := copyStepYML(stepLibURI, stepID, version, destinationStepYML); err != nil {
		return "", fmt.Errorf("copy step.yml: %s", err)
	}

	return path, nil
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
