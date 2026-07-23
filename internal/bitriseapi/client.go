// Package bitriseapi is a minimal client for the Bitrise API
// (https://api.bitrise.io/v0.1). It currently covers only the step-search
// and step-inputs endpoints; broader coverage (pagination, mutating
// requests) gets added when a command group that actually needs it is
// ported.
package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
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
		base += ": " + truncate(e.Body, 500)
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

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "…"
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
