// Package bitriseapi is a minimal client for the Bitrise API
// (https://api.bitrise.io/v0.1).
package bitriseapi

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

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type Option func(*Client)

// WithHTTPClient overrides the default *http.Client — tests use it to point
// the client at an httptest server.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New validates baseURL (via internal/baseurl: https, loopback exempted —
// this client sends a bearer token on every request) before constructing a
// Client against it.
func New(rawBaseURL, token string, opts ...Option) (*Client, error) {
	u, err := baseurl.Validate("API base URL", rawBaseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    u.String(),
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// APIError represents an error response from the Bitrise API. Body is only
// populated when the response had no recognized JSON error field, so
// unexpected error shapes still surface something to the user.
type APIError struct {
	StatusCode  int
	Message     string
	Body        string
	RequestInfo string
}

func (e *APIError) Error() string {
	base := fmt.Sprintf("bitrise API %d", e.StatusCode)
	if e.Message != "" {
		base += ": " + e.Message
	} else if e.Body != "" {
		base += ": " + stringutil.Truncate(e.Body, 500)
	}
	if e.RequestInfo != "" {
		return e.RequestInfo + ": " + base
	}
	return base
}

// errorBody covers the JSON error shapes Bitrise services actually return:
// {"message":...}, {"error_msg":...}, {"error":...}, {"errors":[...]}.
type errorBody struct {
	Message  string   `json:"message"`
	ErrorMsg string   `json:"error_msg"`
	Error    string   `json:"error"`
	Errors   []string `json:"errors"`
}

func (e errorBody) pick() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Error != "" {
		return e.Error
	}
	if len(e.Errors) > 0 {
		return strings.Join(e.Errors, "; ")
	}
	return ""
}

func (c *Client) newRequest(ctx context.Context, path string, params url.Values) (*http.Request, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// post builds and executes a JSON POST request, returning the raw response
// body. params, when non-nil, are appended to the URL as a query string —
// e.g. the validate endpoint's optional app_slug.
func (c *Client) post(ctx context.Context, path string, params url.Values, body any) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return c.do(req)
}

// get performs an authenticated GET against path and decodes the JSON
// response body into T.
func get[T any](ctx context.Context, c *Client, path string, params url.Values) (T, error) {
	var zero T
	req, err := c.newRequest(ctx, path, params)
	if err != nil {
		return zero, err
	}
	body, err := c.do(req)
	if err != nil {
		return zero, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

// envelope unwraps the {"data": ...} shape used by single-resource endpoints
// (e.g. GET /apps/{app-slug}).
type envelope[T any] struct {
	Data T `json:"data"`
}

// getEnvelope performs a GET and unwraps a {"data": T} envelope.
func getEnvelope[T any](ctx context.Context, c *Client, path string, params url.Values) (T, error) {
	env, err := get[envelope[T]](ctx, c, path, params)
	return env.Data, err
}

// page is the {"data": [...], "paging": {"next": ...}} shape used by
// cursor-paginated list endpoints.
type page[T any] struct {
	Data   []T `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

// getPage performs a GET against a cursor-paginated list endpoint, returning
// the page's items and the cursor for the next page ("" when there isn't one).
func getPage[T any](ctx context.Context, c *Client, path string, params url.Values) ([]T, string, error) {
	p, err := get[page[T]](ctx, c, path, params)
	if err != nil {
		return nil, "", err
	}
	return p.Data, p.Paging.Next, nil
}

// postDecode performs an authenticated POST with reqBody as the JSON body,
// and decodes the JSON response into Resp.
func postDecode[Resp any](ctx context.Context, c *Client, path string, params url.Values, reqBody any) (Resp, error) {
	var zero Resp
	body, err := c.post(ctx, path, params, reqBody)
	if err != nil {
		return zero, err
	}
	var resp Resp
	if err := json.Unmarshal(body, &resp); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return resp, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
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
		var e errorBody
		_ = json.Unmarshal(body, &e)
		msg := e.pick()
		apiErr := &APIError{
			StatusCode:  resp.StatusCode,
			Message:     msg,
			RequestInfo: req.Method + " " + req.URL.RequestURI(),
		}
		if msg == "" {
			// No structured field — keep the raw body so the user has
			// something concrete to see (e.g. an unmarshalable Rails 500
			// HTML page or an undocumented error shape).
			apiErr.Body = strings.TrimSpace(string(body))
		}
		return nil, apiErr
	}
	return body, nil
}
