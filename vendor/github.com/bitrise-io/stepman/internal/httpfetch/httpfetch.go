// Package httpfetch is a minimal HTTP client wrapper that exposes a streaming
// Get plus an atomic Download (temp file in dest dir + rename). It's the
// shared transport for stepman's V2 inventory fetches and precompiled
// binary downloads.
package httpfetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-retryablehttp"
)

// Client streams or atomically downloads HTTP resources. Implementations
// returned by NewClient retry transient failures on both Get and Download.
type Client interface {
	// Get streams the body of url. The caller closes the returned reader.
	// Non-2xx responses are returned as an error.
	Get(ctx context.Context, url string) (io.ReadCloser, error)
	// Download fetches url and atomically writes it to destPath. Missing
	// parent directories are created. A temp file is created alongside
	// destPath and renamed on success so partial downloads never appear at
	// the final path.
	Download(ctx context.Context, destPath, url string) error
	// DownloadWithHash behaves like Download but also verifies that the
	// downloaded content matches expectedHash ("sha256-<hex>"). The temp file
	// is removed and an error is returned if the hash does not match, so a
	// mismatched file never appears at destPath.
	DownloadWithHash(ctx context.Context, destPath, url, expectedHash string) error
}

// Logger is the minimal logging interface required by Client.
// Both stepman.Logger and go-utils/v2/log.Logger satisfy it.
type Logger interface {
	Debugf(format string, v ...any)
	Warnf(format string, v ...any)
}

// retryhttpLogger adapts Logger to the retryablehttp.Logger interface (Printf only).
type retryhttpLogger struct{ l Logger }

func (r *retryhttpLogger) Printf(f string, v ...interface{}) { r.l.Debugf(f, v...) }

type client struct {
	httpClient *http.Client
}

// NewClient returns a Client backed by httpClient. When httpClient is nil
// a retryablehttp-backed client is used, so production callers get
// transient-failure retries by default.
func NewClient(httpClient *http.Client, logger Logger) Client {
	if httpClient == nil {
		rc := retryablehttp.NewClient()
		rc.Logger = &retryhttpLogger{l: logger}
		rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
		httpClient = rc.StandardClient()
	}
	return &client{httpClient: httpClient}
}

func (c *client) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", url, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Join(
			fmt.Errorf("GET %s: unexpected status %d", url, resp.StatusCode),
			resp.Body.Close(),
		)
	}
	return resp.Body, nil
}

func (c *client) Download(ctx context.Context, destPath, url string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create dest dir for %s: %w", destPath, err)
	}

	tmpPath, _, err := c.fetchToTemp(ctx, filepath.Dir(destPath), url)
	if err != nil {
		return err
	}

	if renameErr := os.Rename(tmpPath, destPath); renameErr != nil {
		return errors.Join(fmt.Errorf("rename %s to %s: %w", tmpPath, destPath, renameErr), os.Remove(tmpPath))
	}
	return nil
}

func (c *client) DownloadWithHash(ctx context.Context, destPath, url, expectedHash string) error {
	if expectedHash == "" {
		return fmt.Errorf("hash is empty")
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create dest dir for %s: %w", destPath, err)
	}
	tmpPath, hash, err := c.fetchToTemp(ctx, filepath.Dir(destPath), url)
	if err != nil {
		return err
	}

	if hash != expectedHash {
		return errors.Join(
			fmt.Errorf("hash mismatch (%s) expected %s, got %s", url, expectedHash, hash),
			os.Remove(tmpPath),
		)
	}
	if renameErr := os.Rename(tmpPath, destPath); renameErr != nil {
		return errors.Join(fmt.Errorf("rename %s to %s: %w", tmpPath, destPath, renameErr), os.Remove(tmpPath))
	}
	return nil
}

// fetchToTemp streams url into a new temp file under dir and returns its path
// and sha256 hash ("sha256-<hex>"). On error the temp file is removed and
// path/hash are empty; on success the caller owns cleanup.
func (c *client) fetchToTemp(ctx context.Context, dir, url string) (path string, hash string, err error) {
	// Place the temp file alongside destPath so the final rename stays on
	// one filesystem (cross-filesystem renames fail on most kernels).
	tmp, err := os.CreateTemp(dir, "download-*.tmp")
	if err != nil {
		return "", "", fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	defer func() {
		if closeErr := tmp.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close %s: %w", tmp.Name(), closeErr))
		}
		if err != nil {
			err = errors.Join(err, os.Remove(tmp.Name()))
			path = ""
			hash = ""
		}
	}()

	body, err := c.Get(ctx, url)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close response body: %w", closeErr))
		}
	}()

	h := sha256.New()
	if _, copyErr := io.Copy(io.MultiWriter(tmp, h), body); copyErr != nil {
		return "", "", fmt.Errorf("write to %s: %w", tmp.Name(), copyErr)
	}
	return tmp.Name(), "sha256-" + hex.EncodeToString(h.Sum(nil)), nil
}
