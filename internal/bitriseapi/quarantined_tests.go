package bitriseapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// QuarantinedTest is a test the app's test steps exclude from their runs.
type QuarantinedTest struct {
	ID                 string                 `json:"id"`
	TestCaseName       string                 `json:"test_case_name"`
	TestCaseIdentifier string                 `json:"test_case_identifier"`
	TestSuiteName      []string               `json:"test_suite_name"`
	ClassName          string                 `json:"class_name"`
	Comment            string                 `json:"comment"`
	QuarantinedAt      time.Time              `json:"quarantined_at"`
	QuarantinedBy      string                 `json:"quarantined_by"`
	Filters            QuarantinedTestFilters `json:"filters"`
	QuarantinedFrom    QuarantinedTestSource  `json:"quarantined_from"`
}

// QuarantinedTestFilters limits a quarantine to builds on matching branches
// and of matching Workflows or Pipelines (targets). An empty list matches
// everything.
type QuarantinedTestFilters struct {
	Branches []string `json:"branches"`
	Targets  []string `json:"targets"`
}

// QuarantinedTestSource is where a test was quarantined from. Slug and
// Number identify the build or pipeline; source "api" has neither.
type QuarantinedTestSource struct {
	Source string `json:"source"`
	Slug   string `json:"slug"`
	Number int    `json:"number"`
}

// CreateQuarantinedTestRequest is the body of POST /apps/{app-slug}/quarantined-tests.
type CreateQuarantinedTestRequest struct {
	TestCaseName       string                  `json:"test_case_name"`
	TestSuiteName      []string                `json:"test_suite_name"`
	ClassName          string                  `json:"class_name,omitempty"`
	TestCaseIdentifier string                  `json:"test_case_identifier,omitempty"`
	Comment            string                  `json:"comment,omitempty"`
	BuildSlug          string                  `json:"build_slug,omitempty"`
	Filters            *QuarantinedTestFilters `json:"filters,omitempty"`
}

// UpdateQuarantinedTestRequest is the body of PATCH
// /apps/{app-slug}/quarantined-tests/{id}. The API changes only the keys
// sent, so a nil field is left out, and a filter set to a nil list is sent
// as null, which clears it.
type UpdateQuarantinedTestRequest struct {
	Comment  *string
	Branches *[]string
	Targets  *[]string
}

func (r UpdateQuarantinedTestRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if r.Comment != nil {
		body["comment"] = *r.Comment
	}
	filters := map[string]any{}
	if r.Branches != nil {
		filters["branches"] = *r.Branches
	}
	if r.Targets != nil {
		filters["targets"] = *r.Targets
	}
	if len(filters) > 0 {
		body["filters"] = filters
	}
	return json.Marshal(body)
}

// AlreadyQuarantinedError is CreateQuarantinedTest's error when the test is
// already quarantined. The API matches tests by identifier, or by name,
// suites and class name, whatever their filters.
type AlreadyQuarantinedError struct {
	*APIError
	Existing QuarantinedTest
}

func (e *AlreadyQuarantinedError) Unwrap() error { return e.APIError }

// QuarantinedTests lists the quarantined tests of an app, oldest first.
// Endpoint: GET /apps/{app-slug}/quarantined-tests. The list is not paginated.
func (c *Client) QuarantinedTests(ctx context.Context, appSlug string) ([]QuarantinedTest, error) {
	resp, err := get[struct {
		QuarantinedTests []QuarantinedTest `json:"quarantined_tests"`
	}](ctx, c, quarantinedTestsPath(appSlug), nil)
	return resp.QuarantinedTests, err
}

// QuarantinedTest fetches one quarantined test.
// Endpoint: GET /apps/{app-slug}/quarantined-tests/{id}.
func (c *Client) QuarantinedTest(ctx context.Context, appSlug, id string) (QuarantinedTest, error) {
	return get[QuarantinedTest](ctx, c, quarantinedTestPath(appSlug, id), nil)
}

// CreateQuarantinedTest quarantines a test, returning *AlreadyQuarantinedError
// when it already is.
// Endpoint: POST /apps/{app-slug}/quarantined-tests.
func (c *Client) CreateQuarantinedTest(ctx context.Context, appSlug string, req CreateQuarantinedTestRequest) (QuarantinedTest, error) {
	qt, err := postDecode[QuarantinedTest](ctx, c, quarantinedTestsPath(appSlug), nil, req)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
		var conflict struct {
			QuarantinedTest *QuarantinedTest `json:"quarantined_test"`
		}
		if json.Unmarshal(apiErr.rawBody, &conflict) == nil && conflict.QuarantinedTest != nil {
			return QuarantinedTest{}, &AlreadyQuarantinedError{APIError: apiErr, Existing: *conflict.QuarantinedTest}
		}
	}
	return qt, err
}

// UpdateQuarantinedTest changes the comment or filters of a quarantined test.
// Endpoint: PATCH /apps/{app-slug}/quarantined-tests/{id}.
func (c *Client) UpdateQuarantinedTest(ctx context.Context, appSlug, id string, req UpdateQuarantinedTestRequest) (QuarantinedTest, error) {
	return sendDecode[QuarantinedTest](ctx, c, http.MethodPatch, quarantinedTestPath(appSlug, id), nil, req)
}

// DeleteQuarantinedTest removes a test from quarantine.
// Endpoint: DELETE /apps/{app-slug}/quarantined-tests/{id}.
func (c *Client) DeleteQuarantinedTest(ctx context.Context, appSlug, id string) error {
	_, err := c.send(ctx, http.MethodDelete, quarantinedTestPath(appSlug, id), nil, nil)
	return err
}

func quarantinedTestsPath(appSlug string) string {
	return "/apps/" + url.PathEscape(appSlug) + "/quarantined-tests"
}

func quarantinedTestPath(appSlug, id string) string {
	return quarantinedTestsPath(appSlug) + "/" + url.PathEscape(id)
}
