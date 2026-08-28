// Package rdeapi is the HTTP client for the Bitrise Remote Dev Environments
// API (https://api.bitrise.io/rde).
//
// This is a sibling of internal/bitriseapi, not a sub-package: the RDE
// service uses a Bearer authorization header, lives under a different base
// URL, and emits camelCase JSON (grpc-gateway swagger output) with
// google.rpc.Status errors. Wire-format DTOs in this package match the
// backend's shape; the CLI-facing layer in internal/rde converts them into
// the stable snake_case --output json/yml shape.
package rdeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/baseurl"
	"github.com/bitrise-io/bitrise/v2/internal/stringutil"
)

const defaultTimeout = 30 * time.Second

// UserAgent is sent on every RDE request. cli/cli.go overrides it at startup
// to include the binary's version. The backend uses this to attribute
// traffic to CLI vs MCP vs other clients.
var UserAgent = "bitrise-cli"

// requestSource is the value of the X-Request-Source header on every RDE
// request, mirroring the MCP's "X-Request-Source: mcp" so the backend can
// distinguish CLI from MCP traffic without parsing the UA.
const requestSource = "cli"

// Client is an authenticated HTTP client for the RDE API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client (useful for tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a Client authenticated with the given token and base URL.
// rawBaseURL should be the RDE API root (e.g. https://api.bitrise.io/rde) —
// resource paths are appended verbatim. Validated via internal/baseurl
// (https, loopback exempted): this client sends a bearer token on every
// request, so a misconfigured rde_api_base_url must not be able to send it
// over plaintext http. The default client has a defaultTimeout deadline;
// doStream's caller gets streamHTTPClient() instead, for requests that must
// outlive it.
func New(rawBaseURL, token string, opts ...Option) (*Client, error) {
	if _, err := baseurl.Validate("RDE API base URL", rawBaseURL); err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    strings.TrimRight(rawBaseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// wsPath builds a workspace-scoped path under /v1/workspaces/{wsID}/...
// Used by every RDE resource except /v1/me and /v1/saved-inputs.
func wsPath(wsID, p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "/v1/workspaces/" + url.PathEscape(wsID) + p
}

// userPath builds a non-workspace-scoped path (currently /v1/me and
// /v1/saved-inputs/...).
func userPath(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "/v1" + p
}

// APIError represents a non-2xx response from the RDE API. Message is the
// human-readable text extracted from the {"message": "..."} field RDE uses
// universally; Violations holds field-level validation messages pulled from
// the gRPC error details (details[].fieldViolations[]), which carry the
// actionable "why" for 400s (e.g. "missing required input: BUILD_TOKEN");
// Body is the raw response body, surfaced only when no structured field was
// found; RequestInfo is "METHOD /path" so a failure names the endpoint,
// mirroring bitriseapi.APIError.RequestInfo.
type APIError struct {
	StatusCode  int
	Message     string
	Violations  []string
	Body        string
	RequestInfo string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = stringutil.Truncate(e.Body, 500)
	}
	if detail := strings.Join(e.Violations, "; "); detail != "" {
		if msg != "" {
			msg = msg + ": " + detail
		} else {
			msg = detail
		}
	}
	var base string
	if msg != "" {
		base = fmt.Sprintf("RDE API %d: %s", e.StatusCode, msg)
	} else {
		base = fmt.Sprintf("RDE API %d", e.StatusCode)
	}
	if e.RequestInfo != "" {
		return e.RequestInfo + ": " + base
	}
	return base
}

// do executes req and returns the response body on 2xx. Non-2xx responses
// are returned as *APIError.
func (c *Client) do(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Request-Source", requestSource)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(req, resp.StatusCode, body)
	}
	return body, nil
}

// streamHTTPClient returns an http.Client sharing the configured client's
// transport, but with no overall timeout — for streaming endpoints (e.g.
// session logs) that can legitimately run longer than defaultTimeout.
// Cancellation is left to the caller's context. Mirrors
// bitriseapi.Client.streamHTTPClient.
func (c *Client) streamHTTPClient() *http.Client {
	return &http.Client{
		Transport:     c.httpClient.Transport,
		CheckRedirect: c.httpClient.CheckRedirect,
		Jar:           c.httpClient.Jar,
	}
}

// doStream executes req and, on a 2xx response, returns the live
// *http.Response without buffering the body — the caller owns it and MUST
// close resp.Body. Used by streaming endpoints (e.g. session logs) where
// do()'s ReadAll would defeat the point, and where do()'s defaultTimeout
// would kill a long-lived stream, hence streamHTTPClient(). Non-2xx
// responses are drained, closed, and returned as an *APIError, matching
// do().
func (c *Client) doStream(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Request-Source", requestSource)

	resp, err := c.streamHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, parseAPIError(req, resp.StatusCode, body)
	}
	return resp, nil
}

// getJSON performs a GET against path and decodes the response into out.
func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	body, err := c.do(req)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// sendJSON marshals reqBody as JSON, sends it via method to path, and
// decodes the response into out (skipped when out is nil).
func (c *Client) sendJSON(ctx context.Context, method, path string, reqBody, out any) error {
	var r io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		r = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, r)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	body, err := c.do(req)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// del performs a DELETE on path; responses are discarded.
func (c *Client) del(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	_, err = c.do(req)
	return err
}

// errorBody covers the JSON error envelope RDE uses: a gRPC-gateway status
// ({"code": int, "message": string, "details": [...]}). Each details entry
// is a google.rpc.* message; BadRequest entries carry fieldViolations whose
// descriptions explain validation failures.
type errorBody struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details"`
}

// errorDetail is one entry of the gRPC status details array. Only the
// fieldViolations of google.rpc.BadRequest are consumed; other detail types
// unmarshal with an empty FieldViolations and are ignored.
type errorDetail struct {
	Type            string           `json:"@type"`
	FieldViolations []fieldViolation `json:"fieldViolations"`
}

// fieldViolation is a single google.rpc.BadRequest.FieldViolation.
type fieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

// parseAPIError turns a non-2xx (statusCode, body) into an *APIError, naming
// the failing endpoint via req. Shared by do and doStream so the buffering
// and streaming paths surface errors identically.
func parseAPIError(req *http.Request, statusCode int, body []byte) *APIError {
	var e errorBody
	_ = json.Unmarshal(body, &e)
	apiErr := &APIError{
		StatusCode:  statusCode,
		Message:     e.Message,
		Violations:  violationsFromDetails(e.Details),
		RequestInfo: req.Method + " " + req.URL.RequestURI(),
	}
	// Fall back to the raw body only when neither a message nor any field
	// violation gave us something human-readable.
	if e.Message == "" && len(apiErr.Violations) == 0 {
		apiErr.Body = strings.TrimSpace(string(body))
	}
	return apiErr
}

// violationsFromDetails pulls the human-readable field-violation strings out
// of a gRPC status' details array (preferring the description, falling back
// to the field name).
func violationsFromDetails(details []errorDetail) []string {
	var out []string
	for _, d := range details {
		for _, v := range d.FieldViolations {
			switch {
			case v.Description != "":
				out = append(out, v.Description)
			case v.Field != "":
				out = append(out, v.Field)
			}
		}
	}
	return out
}
