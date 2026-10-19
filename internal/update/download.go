package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bitrise-io/bitrise/v3/tools"
)

const (
	defaultDownloadURL = "https://github.com/bitrise-io/bitrise/releases/download"
	downloadTimeout    = 5 * time.Minute
)

type Downloader struct {
	downloadURL string
	httpClient  *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		downloadURL: defaultDownloadURL,
		httpClient:  &http.Client{Timeout: downloadTimeout},
	}
}

// Install replaces the binary at targetPath with the release build for the running machine.
func (d *Downloader) Install(ctx context.Context, version, targetPath string) error {
	asset, err := releaseAsset()
	if err != nil {
		return err
	}

	// Written beside the target, so the rename below stays inside one filesystem
	// and the user is never left without a CLI.
	tmp, err := os.CreateTemp(filepath.Dir(targetPath), ".bitrise-update-")
	if err != nil {
		return fmt.Errorf("create a temporary file next to %s: %w", targetPath, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	err = d.fetch(ctx, fmt.Sprintf("%s/v%s/%s", d.downloadURL, version, asset), tmp)
	if closeErr := tmp.Close(); err == nil && closeErr != nil {
		err = fmt.Errorf("write %s: %w", tmp.Name(), closeErr)
	}
	if err != nil {
		return err
	}

	if err := os.Chmod(tmp.Name(), targetMode(targetPath)); err != nil {
		return fmt.Errorf("set the permissions of %s: %w", tmp.Name(), err)
	}

	if err := os.Rename(tmp.Name(), targetPath); err != nil {
		return fmt.Errorf("replace %s: %w", targetPath, err)
	}
	return nil
}

// BinaryPath resolves the running binary through its symlinks, so an update replaces
// the installed file and not the link pointing at it.
func BinaryPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate the running Bitrise CLI: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", executable, err)
	}
	return resolved, nil
}

// Release assets are named after uname's output, not Go's GOOS and GOARCH.
func releaseAsset() (string, error) {
	osName, osErr := tools.UnameGOOS()
	arch, archErr := tools.UnameGOARCH()
	if osErr != nil || archErr != nil {
		return "", fmt.Errorf("no Bitrise CLI release is published for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("bitrise-%s-%s", osName, arch), nil
}

func (d *Downloader) fetch(ctx context.Context, url string, target io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, http.StatusText(resp.StatusCode))
	}

	if _, err := io.Copy(target, resp.Body); err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	return nil
}

// A target that is not there yet is about to become the CLI, so it needs the executable bit.
func targetMode(targetPath string) os.FileMode {
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0755
	}
	return info.Mode().Perm()
}
