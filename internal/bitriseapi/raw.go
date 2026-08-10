package bitriseapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RawResponse is the unprocessed result of a RawRequest call. Unlike the
// client's typed methods, a non-2xx StatusCode is not an error here — the
// caller (the `api` command) decides how to present it.
type RawResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// RawRequest performs an arbitrary authenticated request against the Bitrise
// API, for callers that need direct control over method, path, query,
// headers, and body rather than one of the typed endpoints.
//
// path is resolved relative to the client's configured base URL, unless it
// is itself an absolute http(s):// URL, which is used verbatim (still with
// the same Authorization/Accept headers). query is merged with any query
// string already present in path.
func (c *Client) RawRequest(ctx context.Context, method, path string, query url.Values, header http.Header, body io.Reader) (*RawResponse, error) {
	reqURL, err := c.resolveURL(path, query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	for k, values := range header {
		req.Header.Del(k)
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &RawResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: respBody}, nil
}

func (c *Client) resolveURL(path string, query url.Values) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return mergeQuery(path, query)
	}
	p := path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return mergeQuery(strings.TrimSuffix(c.baseURL, "/")+p, query)
}

func mergeQuery(rawURL string, extra url.Values) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	if len(extra) > 0 {
		q := u.Query()
		for k, values := range extra {
			for _, v := range values {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}
