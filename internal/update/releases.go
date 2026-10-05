package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bitrise-io/bitrise/v3/internal/stringutil"
)

const (
	defaultReleasesURL = "https://api.github.com/repos/bitrise-io/bitrise/releases"
	releasesPerPage    = 100
	requestTimeout     = 30 * time.Second
	errorBodyLimit     = 500
)

// Release is the part of a GitHub release the CLI reads.
type Release struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// Client reads the published releases of the Bitrise CLI repository.
type Client struct {
	releasesURL string
	httpClient  *http.Client
}

type Option func(*Client)

func NewClient(opts ...Option) *Client {
	c := &Client{
		releasesURL: defaultReleasesURL,
		httpClient:  &http.Client{Timeout: requestTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithReleasesURL overrides the GitHub endpoint the client reads - tests point
// it at an httptest server.
func WithReleasesURL(releasesURL string) Option {
	return func(c *Client) { c.releasesURL = releasesURL }
}

// WithHTTPClient overrides the default *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// ReleasesPage returns one page of releases, newest first. page is 1-based.
func (c *Client) ReleasesPage(ctx context.Context, page int) ([]Release, error) {
	u, err := url.Parse(c.releasesURL)
	if err != nil {
		return nil, fmt.Errorf("parse releases URL: %w", err)
	}
	q := u.Query()
	q.Set("per_page", strconv.Itoa(releasesPerPage))
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
		return nil, fmt.Errorf("GitHub releases API %d: %s", resp.StatusCode, stringutil.Truncate(string(body), errorBodyLimit))
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("parse releases response: %w", err)
	}
	return releases, nil
}
