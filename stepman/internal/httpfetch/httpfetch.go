// Package httpfetch is a minimal HTTP client wrapper that exposes a streaming
// Get plus an atomic Download (temp file in dest dir + rename). It's the
// shared transport for stepman's V2 inventory fetches and precompiled
// binary downloads.
package httpfetch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/bartventer/httpcache/store/memcache"
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

// Clients are a pair of caching and passthrough clients.
// They share one connection pool, for performance.
type Clients struct {
	// Caching reads through a HTTP cache that honours Cache-Control and revalidates with ETags.
	Caching Client
	// Passthrough fetches step archives and precompiled executables, uncached.
	Passthrough Client
}

// Logger is the minimal logging interface Client needs; the retry adapter only
// emits debug lines.
type Logger interface {
	Debugf(format string, v ...any)
}

// retryhttpLogger adapts Logger to the retryablehttp.Logger interface (Printf only).
type retryhttpLogger struct{ l Logger }

func (r *retryhttpLogger) Printf(f string, v ...any) { r.l.Debugf(f, v...) }

type client struct {
	httpClient *http.Client
}

// newRetryingClient returns an *http.Client that retries transient failures.
func newRetryingClient(logger Logger) *http.Client {
	rc := retryablehttp.NewClient()
	rc.Logger = &retryhttpLogger{l: logger}
	rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return rc.StandardClient()
}

// NewClient returns a Client backed by a retryablehttp client.
func NewClient(logger Logger) Client {
	return &client{httpClient: newRetryingClient(logger)}
}

// NewCachingClient returns a caching and a passthrough Client. Both run over the
// same retrying transport, so they share one connection pool.
func NewCachingClient(logger Logger) (Clients, error) {
	// The cache is layered above the retries, so a 500 is retried before it is stored.
	passthrough := newRetryingClient(logger)
	caching, err := NewCachingWithClient(logger, passthrough)
	if err != nil {
		return Clients{}, err
	}

	return Clients{
		Caching:     caching,
		Passthrough: &client{httpClient: passthrough},
	}, nil
}

// NewCachingWithClient returns a caching Client.
// Prefer NewCachingClient unless you need a specific transport.
func NewCachingWithClient(logger Logger, httpClient *http.Client) (Client, error) {
	upstream := httpClient.Transport
	if upstream == nil {
		upstream = http.DefaultTransport // what net/http itself uses for a nil Transport
	}

	cached, err := newCachingTransport(logger, upstream)
	if err != nil {
		return nil, err
	}

	caching := *httpClient
	caching.Transport = cached
	return &client{httpClient: &caching}, nil
}

// NewWithClient returns a Client backed by the given httpClient.
// Prefer NewClient unless you need a specific transport.
func NewWithClient(httpClient *http.Client) Client {
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
		snippet, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, errors.Join(
			&StatusError{URL: url, Code: resp.StatusCode, Body: string(bytes.TrimSpace(snippet))},
			readErr,
			resp.Body.Close(),
		)
	}
	return resp.Body, nil
}

// StatusError is returned by Get when the server responds with a non-2xx
// status, so callers can branch on the code (e.g. treat 404 as "not found")
// via errors.As.
type StatusError struct {
	URL  string
	Code int
	Body string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("GET %s: unexpected status %d: %s", e.URL, e.Code, e.Body)
}

func (c *client) Download(ctx context.Context, destPath, url string) error {
	return c.download(ctx, destPath, url, "")
}

func (c *client) DownloadWithHash(ctx context.Context, destPath, url, expectedHash string) error {
	if expectedHash == "" {
		return fmt.Errorf("hash is empty")
	}
	return c.download(ctx, destPath, url, expectedHash)
}

// download fetches url into a temp file alongside destPath and atomically
// renames it into place. When expectedHash is non-empty the content is verified
// against it ("sha256-<hex>") before the rename, so a mismatched or partial
// file never lands at destPath.
func (c *client) download(ctx context.Context, destPath, url, expectedHash string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create dest dir for %s: %w", destPath, err)
	}

	tmpPath, hash, err := c.fetchToTemp(ctx, filepath.Dir(destPath), url)
	if err != nil {
		return err
	}
	// Best-effort cleanup: on a verify/rename failure this removes the temp
	// file; after a successful rename tmpPath is gone, so Remove fails harmlessly.
	defer func() { _ = os.Remove(tmpPath) }()

	if expectedHash != "" && hash != expectedHash {
		return fmt.Errorf("hash mismatch (%s) expected %s, got %s", url, expectedHash, hash)
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmpPath, destPath, err)
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
		// A failed Close is intentionally a hard failure: it can mean the final
		// write never flushed to disk, so the temp file may be incomplete.
		if closeErr := tmp.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close %s: %w", tmp.Name(), closeErr))
		}
		if err != nil {
			// Best-effort cleanup of the partial temp file; the original error is
			// what matters, so a failed remove is intentionally ignored.
			_ = os.Remove(tmp.Name())
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
