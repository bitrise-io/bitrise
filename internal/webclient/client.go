// Package webclient is a short-lived HTTP client for app.bitrise.io's
// Rails-Devise endpoints (signup, email/password sign-in).
//
// Separate from internal/bitriseapi (bearer tokens against api.bitrise.io):
// the website uses cookie sessions plus CSRF, and the jar lives only inside
// one command invocation — never persisted.
package webclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/version"
)

// defaultTimeout bounds a single request against app.bitrise.io, mirroring
// internal/bitriseapi's client so a stalled server can't hang a command
// indefinitely.
const defaultTimeout = 30 * time.Second

// userAgent identifies this CLI to app.bitrise.io, which sits behind a
// CDN/WAF that can block Go's default User-Agent.
var userAgent = "bitrise/" + version.VERSION

// Client wraps an http.Client with a per-call cookie jar and CSRF priming.
// Construct one per command invocation; do not reuse across commands.
type Client struct {
	baseURL    string
	httpClient *http.Client
	csrfToken  string
}

// New builds a Client targeting baseURL, with its own cookie jar — pass one
// Client through a single sequence of calls (Prime → PostJSON …) and discard it.
func New(baseURL string) (*Client, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	return &Client{
		baseURL: normalized,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: defaultTimeout,
			// This is a JSON API flow; a redirect means something unexpected
			// happened (e.g. a failed sign-in bouncing to an HTML page).
			// ErrUseLastResponse stops at the redirect response itself so it
			// surfaces through the normal non-2xx status handling below,
			// rather than being followed silently or resending the request
			// body to whatever host Location names.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// metaCSRFRegexp extracts the token from the <meta name="csrf-token"> tag
// Rails emits on every page, rather than the CSRF-TOKEN cookie — Rails masks
// the cookie value differently on each request, while the meta value is the
// canonical request-validating token.
var metaCSRFRegexp = regexp.MustCompile(`<meta\s+name="csrf-token"\s+content="([^"]+)"`)

// Prime issues a GET to an HTML path (e.g. "/users/sign_up"), captures
// cookies, and extracts the CSRF token for subsequent PostJSON calls.
func (c *Client) Prime(ctx context.Context, path string) error {
	u, err := c.url(path)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build prime request: %w", err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("prime: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return fmt.Errorf("prime: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("prime %s: unexpected status %d", path, resp.StatusCode)
	}
	m := metaCSRFRegexp.FindSubmatch(body)
	if m == nil {
		return fmt.Errorf("prime %s: no CSRF token found in response", path)
	}
	c.csrfToken = string(m[1])
	return nil
}

// Response carries the status code and body of a website request.
type Response struct {
	Status int
	Body   []byte
}

// PostJSON sends body as JSON to path, attaching X-CSRF-Token when Prime
// captured one. Any "skip CSRF" exception lives at the call site.
func (c *Client) PostJSON(ctx context.Context, path string, body any) (Response, error) {
	u, err := c.url(path)
	if err != nil {
		return Response{}, err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return Response{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("POST %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	return Response{
		Status: resp.StatusCode,
		Body:   respBody,
	}, nil
}

func (c *Client) url(path string) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	return u.String(), nil
}

// normalizeBaseURL rejects a relative URL or a plaintext http:// URL against
// a non-loopback host (this flow sends a password), and trims a trailing
// slash so concatenating a leading-slash path in url() can't produce "//".
// Loopback is allowed over http so tests can point at an httptest server.
func normalizeBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse base URL %q: %w", raw, err)
	}
	if !u.IsAbs() || u.Host == "" {
		return "", fmt.Errorf("base URL %q must be an absolute URL", raw)
	}
	if u.Scheme != "https" && !isLoopbackHost(u.Hostname()) {
		return "", fmt.Errorf("base URL %q must use https (got %q)", raw, u.Scheme)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String(), nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
