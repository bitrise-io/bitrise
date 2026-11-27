package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
)

// QuarantinedTest is a test the app's test steps exclude from their runs.
type QuarantinedTest struct {
	ID                 string            `json:"id" yaml:"id"`
	TestCaseName       string            `json:"test_case_name" yaml:"test_case_name"`
	TestCaseIdentifier string            `json:"test_case_identifier,omitempty" yaml:"test_case_identifier,omitempty"`
	TestSuiteName      []string          `json:"test_suite_name" yaml:"test_suite_name"`
	ClassName          string            `json:"class_name,omitempty" yaml:"class_name,omitempty"`
	Comment            string            `json:"comment,omitempty" yaml:"comment,omitempty"`
	QuarantinedAt      time.Time         `json:"quarantined_at" yaml:"quarantined_at"`
	QuarantinedBy      string            `json:"quarantined_by,omitempty" yaml:"quarantined_by,omitempty"`
	Filters            QuarantineFilters `json:"filters" yaml:"filters"`
	QuarantinedFrom    QuarantineSource  `json:"quarantined_from" yaml:"quarantined_from"`
}

// QuarantineFilters limits a quarantine to builds on matching branches and of
// matching Workflows or Pipelines (targets). Values may contain * wildcards,
// and an empty list matches everything.
type QuarantineFilters struct {
	Branches []string `json:"branches" yaml:"branches"`
	Targets  []string `json:"targets" yaml:"targets"`
}

// QuarantineSource is where a test was quarantined from: source is api,
// build, insights or pipeline, and ID and Number identify the build or
// pipeline.
type QuarantineSource struct {
	Source string `json:"source" yaml:"source"`
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Number int    `json:"number,omitempty" yaml:"number,omitempty"`
}

// QuarantinedTestsResult is the quarantined tests of an app, oldest first.
type QuarantinedTestsResult struct {
	Items []QuarantinedTest `json:"items" yaml:"items"`
}

// AddQuarantinedTestOptions identifies the test to quarantine and the
// scope of the quarantine.
//
// Without BuildID, the test fields are sent as given, and TestSuiteName is
// required. With BuildID, the test case is looked up in that build's test
// results and quarantined with the values it has there, as the build page
// does: TestCaseName matches its name, or its name as test results list it
// (internalbuild.TestCaseDisplayName), and TestSuiteName, ClassName and
// TestCaseIdentifier only narrow the lookup.
type AddQuarantinedTestOptions struct {
	TestCaseName       string
	TestSuiteName      []string // outermost first
	ClassName          string
	TestCaseIdentifier string
	Comment            string
	BuildID            string
	Branches           []string
	Targets            []string
}

// UpdateQuarantinedTestOptions changes only its non-nil fields. A filter set
// to a nil list is cleared, so the quarantine applies to every branch or
// target.
type UpdateQuarantinedTestOptions struct {
	Comment  *string
	Branches *[]string
	Targets  *[]string
}

// ListQuarantinedTests returns the quarantined tests of an app.
func (s *Service) ListQuarantinedTests(ctx context.Context, appSlug string) (QuarantinedTestsResult, error) {
	tests, err := s.client.QuarantinedTests(ctx, appSlug)
	if err != nil {
		return QuarantinedTestsResult{}, err
	}
	items := make([]QuarantinedTest, 0, len(tests))
	for _, t := range tests {
		items = append(items, toQuarantinedTest(t))
	}
	return QuarantinedTestsResult{Items: items}, nil
}

// GetQuarantinedTest returns one quarantined test of an app.
func (s *Service) GetQuarantinedTest(ctx context.Context, appSlug, id string) (QuarantinedTest, error) {
	t, err := s.client.QuarantinedTest(ctx, appSlug, id)
	if err != nil {
		return QuarantinedTest{}, err
	}
	return toQuarantinedTest(t), nil
}

// AddQuarantinedTest quarantines a test. If the test is already
// quarantined, the error names the existing quarantine, whose scope the
// caller can update instead.
func (s *Service) AddQuarantinedTest(ctx context.Context, appSlug string, opts AddQuarantinedTestOptions) (QuarantinedTest, error) {
	if opts.BuildID == "" && len(opts.TestSuiteName) == 0 {
		return QuarantinedTest{}, errors.New("pass --build to look the test case up in a build, or --suite with the suites it belongs to")
	}
	if opts.BuildID != "" {
		tc, err := s.findTestCase(ctx, appSlug, opts)
		if err != nil {
			return QuarantinedTest{}, err
		}
		opts.TestCaseName = tc.Name
		opts.TestSuiteName = tc.TestSuites
		opts.ClassName = tc.ClassName
		opts.TestCaseIdentifier = tc.Identifier
	}

	req := bitriseapi.CreateQuarantinedTestRequest{
		TestCaseName:       opts.TestCaseName,
		TestSuiteName:      opts.TestSuiteName,
		ClassName:          opts.ClassName,
		TestCaseIdentifier: opts.TestCaseIdentifier,
		Comment:            opts.Comment,
		BuildSlug:          opts.BuildID,
	}
	if len(opts.Branches) > 0 || len(opts.Targets) > 0 {
		req.Filters = &bitriseapi.QuarantinedTestFilters{Branches: opts.Branches, Targets: opts.Targets}
	}

	t, err := s.client.CreateQuarantinedTest(ctx, appSlug, req)
	var conflict *bitriseapi.AlreadyQuarantinedError
	if errors.As(err, &conflict) {
		return QuarantinedTest{}, fmt.Errorf("the test is already quarantined as %s; change its scope with 'bitrise app quarantined-tests update %s'", conflict.Existing.ID, conflict.Existing.ID)
	}
	if err != nil {
		return QuarantinedTest{}, err
	}
	return toQuarantinedTest(t), nil
}

// UpdateQuarantinedTest changes the comment or filters of a quarantined test.
func (s *Service) UpdateQuarantinedTest(ctx context.Context, appSlug, id string, opts UpdateQuarantinedTestOptions) (QuarantinedTest, error) {
	t, err := s.client.UpdateQuarantinedTest(ctx, appSlug, id, bitriseapi.UpdateQuarantinedTestRequest{
		Comment:  opts.Comment,
		Branches: opts.Branches,
		Targets:  opts.Targets,
	})
	if err != nil {
		return QuarantinedTest{}, err
	}
	return toQuarantinedTest(t), nil
}

// RemoveQuarantinedTest removes a test from quarantine.
func (s *Service) RemoveQuarantinedTest(ctx context.Context, appSlug, id string) error {
	return s.client.DeleteQuarantinedTest(ctx, appSlug, id)
}

// findTestCase returns the one test case of opts.BuildID that opts selects.
// Identical cases, like a test the report lists once per device, count once.
func (s *Service) findTestCase(ctx context.Context, appSlug string, opts AddQuarantinedTestOptions) (internalbuild.TestCase, error) {
	res, err := internalbuild.NewService(s.client).TestResults(ctx, appSlug, opts.BuildID, internalbuild.TestResultsFilter{})
	if err != nil {
		return internalbuild.TestCase{}, err
	}

	var found []internalbuild.TestCase
	seen := map[string]bool{}
	var walk func(suites []internalbuild.TestSuite)
	walk = func(suites []internalbuild.TestSuite) {
		for _, su := range suites {
			for _, tc := range su.TestCases {
				key := strings.Join([]string{tc.Name, tc.ClassName, strings.Join(tc.TestSuites, "\x00"), tc.Identifier}, "\x01")
				if !seen[key] && testCaseMatches(tc, opts) {
					seen[key] = true
					found = append(found, tc)
				}
			}
			walk(su.TestSuites)
		}
	}
	walk(res.TestSuites)

	switch {
	case len(found) == 0:
		return internalbuild.TestCase{}, fmt.Errorf("build %s has no test case %q matching the given flags; pass the name as 'bitrise build test-results %s --app %s' lists it", opts.BuildID, opts.TestCaseName, opts.BuildID, appSlug)
	case len(found) > 1:
		var b strings.Builder
		fmt.Fprintf(&b, "%d test cases of build %s match; narrow it down with --class-name, --suite or --identifier:", len(found), opts.BuildID)
		for i, tc := range found {
			if i == maxListedMatches {
				fmt.Fprintf(&b, "\n  and %d more", len(found)-i)
				break
			}
			fmt.Fprintf(&b, "\n  suites: %s  class name: %s  name: %s", strings.Join(tc.TestSuites, " › "), tc.ClassName, tc.Name)
			if tc.Identifier != "" {
				fmt.Fprintf(&b, "  identifier: %s", tc.Identifier)
			}
		}
		return internalbuild.TestCase{}, errors.New(b.String())
	case len(found[0].TestSuites) == 0:
		// The API requires the suites, and the build page cannot quarantine such a test case either.
		return internalbuild.TestCase{}, fmt.Errorf("the test report of build %s lists no suites for test case %q, so it cannot be quarantined", opts.BuildID, found[0].Name)
	}
	return found[0], nil
}

const maxListedMatches = 10

func testCaseMatches(tc internalbuild.TestCase, opts AddQuarantinedTestOptions) bool {
	if opts.TestCaseName != tc.Name && opts.TestCaseName != internalbuild.TestCaseDisplayName(tc.ClassName, tc.Name) {
		return false
	}
	if opts.ClassName != "" && opts.ClassName != tc.ClassName {
		return false
	}
	if opts.TestCaseIdentifier != "" && opts.TestCaseIdentifier != tc.Identifier {
		return false
	}
	for _, su := range opts.TestSuiteName {
		if !slices.Contains(tc.TestSuites, su) {
			return false
		}
	}
	return true
}

func toQuarantinedTest(t bitriseapi.QuarantinedTest) QuarantinedTest {
	return QuarantinedTest{
		ID:                 t.ID,
		TestCaseName:       t.TestCaseName,
		TestCaseIdentifier: t.TestCaseIdentifier,
		TestSuiteName:      nonNil(t.TestSuiteName),
		ClassName:          t.ClassName,
		Comment:            t.Comment,
		QuarantinedAt:      t.QuarantinedAt.UTC(),
		QuarantinedBy:      t.QuarantinedBy,
		Filters: QuarantineFilters{
			Branches: nonNil(t.Filters.Branches),
			Targets:  nonNil(t.Filters.Targets),
		},
		QuarantinedFrom: QuarantineSource{
			Source: t.QuarantinedFrom.Source,
			ID:     t.QuarantinedFrom.Slug,
			Number: t.QuarantinedFrom.Number,
		},
	}
}

// nonNil keeps an absent list as [] in JSON and YAML output, so a script can
// iterate the filters without a null check.
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
