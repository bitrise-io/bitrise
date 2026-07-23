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
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

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
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Jar: jar},
	}, nil
}

// WithHTTPClient replaces the default http.Client. Used by tests; the
// supplied client must have a non-nil cookie Jar.
func (c *Client) WithHTTPClient(hc *http.Client) *Client {
	c.httpClient = hc
	return c
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
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("prime: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return fmt.Errorf("prime: read body: %w", err)
	}
	if m := metaCSRFRegexp.FindSubmatch(body); m != nil {
		c.csrfToken = string(m[1])
	}
	return nil
}

// Response carries the decoded body, status code, and final URL of a
// website request.
type Response struct {
	Status   int
	Location string
	Body     []byte
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
		Status:   resp.StatusCode,
		Location: resp.Header.Get("Location"),
		Body:     respBody,
	}, nil
}

func (c *Client) url(path string) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	return u.String(), nil
}
