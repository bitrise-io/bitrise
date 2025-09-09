package mise

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/hashicorp/go-retryablehttp"
)

const fallbackDownloadURLBase = "https://storage.googleapis.com/mise-release-mirror"

// installReleaseBinary installs the release binary for the specified version of Mise.
func installReleaseBinary(version string, checksums map[string]string, targetDir string) error {
	platformName, err := getPlatformName()
	if err != nil {
		return err
	}

	checksum, ok := checksums[platformName]
	if !ok {
		return fmt.Errorf("checksum not found for %s", platformName)
	}

	url := primaryDownloadURL(version, platformName)

	tempPath, err := downloadAndVerify(url, checksum)
	if err != nil {
		url = FallbackDownloadURL(version, platformName)
		tempPath, err = downloadAndVerify(url, checksum)
		if err != nil {
			return err
		}
	}
	defer func() {
		_ = os.Remove(tempPath)
	}()

	miseBinFound, err := extractTarball(tempPath, targetDir)
	if err != nil {
		return err
	}

	if !miseBinFound {
		return fmt.Errorf("mise binary not found in tarball from %s", url)
	}

	return nil
}

// downloadAndVerify downloads a file from the given URL, verifies its checksum,
// and returns the path to the downloaded file.
func downloadAndVerify(url, expectedChecksum string) (string, error) {
	// Get the tarball, check status
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	client := retryablehttp.NewClient()
	client.Logger = &log.HTTPLogAdaptor{Logger: logger}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: received status code %d", url, resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "mise-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
	}()

	// Compute SHA256 hash of the downloaded file and store contents in the temp file
	hash := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hash)
	if _, err := io.Copy(multiWriter, resp.Body); err != nil {
		return "", fmt.Errorf("save download to temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}
	calculatedChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	if calculatedChecksum != expectedChecksum {
		return "", fmt.Errorf("checksum validation failed: expected %s, got %s", expectedChecksum, calculatedChecksum)
	}

	return tempPath, nil
}

// extractTarball extracts a tarball file to the specified directory
// and returns whether the mise binary was found inside.
func extractTarball(tarballPath, targetDir string) (bool, error) {
	file, err := os.Open(tarballPath)
	if err != nil {
		return false, fmt.Errorf("open temp file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return false, fmt.Errorf("create gzip reader: %w", err)
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false, fmt.Errorf("create directory for %s: %w", targetDir, err)
	}

	tarReader := tar.NewReader(gzipReader)
	var miseBinFound bool

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("read tar header: %w", err)
		}

		if extractedPath, shouldExtract := processHeader(header, targetDir); shouldExtract {
			if err := extractFile(tarReader, header, extractedPath); err != nil {
				return false, err
			}
			if filepath.Base(extractedPath) == "mise" {
				miseBinFound = true
			}
		}
	}

	return miseBinFound, nil
}

// getPlatformName returns OS and architecture in a format used for Mise binaries.
func getPlatformName() (string, error) {
	osMap := map[string]string{
		"darwin": "macos",
		"linux":  "linux",
	}
	archMap := map[string]string{
		"amd64": "x64",
		"arm64": "arm64",
	}

	osString, ok := osMap[runtime.GOOS]
	if !ok {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	archString, ok := archMap[runtime.GOARCH]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return fmt.Sprintf("%s-%s", osString, archString), nil
}

func primaryDownloadURL(version, platformName string) string {
	version = strings.TrimPrefix(version, "v")
	artifactName := fmt.Sprintf("mise-v%s-%s.tar.gz", version, platformName)
	url := fmt.Sprintf("https://github.com/jdx/mise/releases/download/v%s/%s", version, artifactName)
	return url
}

func FallbackDownloadURL(version, platformName string) string {
	version = strings.TrimPrefix(version, "v")
	artifactName := fmt.Sprintf("mise-v%s-%s.tar.gz", version, platformName)
	url := fmt.Sprintf("%s/v%s/%s", fallbackDownloadURLBase, version, artifactName)
	return url
}

// processHeader processes a tar header and determines the target extraction path.
func processHeader(header *tar.Header, targetDir string) (string, bool) {
	// Skip the top-level "mise" directory and extract its contents directly
	pathParts := strings.Split(header.Name, "/")
	if len(pathParts) > 0 && pathParts[0] == "mise" {
		if len(pathParts) == 1 {
			// This is the top-level "mise" directory itself, skip it
			return "", false
		}
		// Remove the top-level "mise" directory from the path
		header.Name = strings.Join(pathParts[1:], "/")
	}

	// Clean the path to prevent directory traversal attacks
	targetPath := filepath.Join(targetDir, header.Name)
	if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)) {
		return "", false
	}

	return targetPath, true
}

// extractFile extracts a file from the tar reader to the target path.
func extractFile(tarReader *tar.Reader, header *tar.Header, targetPath string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		err := os.MkdirAll(targetPath, 0755)
		if err != nil {
			return fmt.Errorf("create directory %s: %w", targetPath, err)
		}
	case tar.TypeReg:
		dir := filepath.Dir(targetPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("create parent directory for %s: %w", targetPath, err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("create file %s: %w", targetPath, err)
		}
		defer func() {
			_ = outFile.Close()
		}()

		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			return fmt.Errorf("extract file %s: %w", targetPath, err)
		}

		if filepath.Base(targetPath) == "mise" {
			// Make mise binary executable
			err = os.Chmod(targetPath, 0755)
			if err != nil {
				return fmt.Errorf("make mise binary executable %s: %w", targetPath, err)
			}
		}
	default:
		// Skip other file types (symlinks, etc.)
	}
	return nil
}
