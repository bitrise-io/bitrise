package mise

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

func installReleaseBinary(version string, targetDir string) error {
	url, err := downloadURL(version)
	if err != nil {
		return err
	}

	resp, err := retryablehttp.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: received status code %d", url, resp.StatusCode)
	}

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory for %s: %w", targetDir, err)
	}

	tarReader := tar.NewReader(gzipReader)
	var miseBinFound bool

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		if extractedPath, shouldExtract := processHeader(header, targetDir); shouldExtract {
			if err := extractFile(tarReader, header, extractedPath); err != nil {
				return err
			}
			if filepath.Base(extractedPath) == "mise" {
				miseBinFound = true
			}
		}
	}

	if !miseBinFound {
		return fmt.Errorf("mise binary not found in tarball from %s", url)
	}

	return nil
}

func downloadURL(version string) (string, error) {
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
	version = strings.TrimPrefix(version, "v")
	artifactName := fmt.Sprintf("mise-v%s-%s-%s.tar.gz", version, osString, archString)
	url := fmt.Sprintf("https://github.com/jdx/mise/releases/download/v%s/%s", version, artifactName)
	return url, nil
}

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
