// Package build holds the business-logic layer for build operations.
package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
)

// Build is the CLI-facing build record. JSON/YAML tags define the stable
// --format json/yml shape.
type Build struct {
	Slug                    string     `json:"id" yaml:"id"`
	AppSlug                 string     `json:"app_id" yaml:"app_id"`
	BuildNumber             int        `json:"build_number" yaml:"build_number"`
	Status                  string     `json:"status" yaml:"status"`
	StatusText              string     `json:"status_text,omitempty" yaml:"status_text,omitempty"`
	AbortReason             string     `json:"abort_reason,omitempty" yaml:"abort_reason,omitempty"`
	IsOnHold                bool       `json:"is_on_hold,omitempty" yaml:"is_on_hold,omitempty"`
	Rebuildable             bool       `json:"rebuildable,omitempty" yaml:"rebuildable,omitempty"`
	Workflow                string     `json:"workflow,omitempty" yaml:"workflow,omitempty"`
	PipelineWorkflowID      string     `json:"pipeline_workflow_id,omitempty" yaml:"pipeline_workflow_id,omitempty"`
	Branch                  string     `json:"branch,omitempty" yaml:"branch,omitempty"`
	Tag                     string     `json:"tag,omitempty" yaml:"tag,omitempty"`
	PullRequestID           int        `json:"pull_request_id,omitempty" yaml:"pull_request_id,omitempty"`
	PullRequestTargetBranch string     `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	PullRequestViewURL      string     `json:"pull_request_view_url,omitempty" yaml:"pull_request_view_url,omitempty"`
	CommitHash              string     `json:"commit_hash,omitempty" yaml:"commit_hash,omitempty"`
	CommitMessage           string     `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	TriggeredAt             time.Time  `json:"triggered_at,omitempty" yaml:"triggered_at,omitempty"`
	TriggeredBy             string     `json:"triggered_by,omitempty" yaml:"triggered_by,omitempty"`
	FinishedAt              *time.Time `json:"finished_at,omitempty" yaml:"finished_at,omitempty"`
	StackIdentifier         string     `json:"stack_identifier,omitempty" yaml:"stack_identifier,omitempty"`
	MachineTypeID           string     `json:"machine_type_id,omitempty" yaml:"machine_type_id,omitempty"`
	CreditCost              int        `json:"credit_cost,omitempty" yaml:"credit_cost,omitempty"`
	BuildURL                string     `json:"build_url,omitempty" yaml:"build_url,omitempty"`
}

// TriggerEnv is an environment variable to inject into a triggered build.
type TriggerEnv struct {
	Key      string
	Value    string
	IsExpand bool
}

// TriggerRequest describes a build to start.
type TriggerRequest struct {
	AppSlug       string
	Workflow      string
	Pipeline      string
	Branch        string
	BranchDest    string
	Tag           string
	CommitHash    string
	CommitMessage string
	PullRequestID int
	Priority      *int
	Stack         string
	MachineTypeID string
	LicensePoolID string
	Environments  []TriggerEnv
}

// ListOptions filters and paginates build lists. Status is a CLI-friendly
// string ("success", "failed", "aborted", "aborted-with-success",
// "in-progress"); the service translates it to the API's integer value.
// IsPipelineBuild is a tri-state: nil = no filter, true/false = filtered.
type ListOptions struct {
	AppSlug          string
	Branch           string
	Workflow         string
	CommitMessage    string
	TriggerEventType string
	PullRequestID    int
	BuildNumber      int
	After            *time.Time
	Before           *time.Time
	Status           string
	SortBy           string
	IsPipelineBuild  *bool
	Limit            int
	Cursor           string
}

// ListResult is one page of builds plus the cursor for the next page.
type ListResult struct {
	Items      []Build `json:"items" yaml:"items"`
	NextCursor string  `json:"next_cursor,omitempty" yaml:"next_cursor,omitempty"`
}

// Step is one step run of a build. Workflow is the workflow it ran in, so a
// flat list keeps the grouping. ID is the step run's UUID, not the step's
// library ID, which is StepID.
type Step struct {
	ID           string      `json:"id" yaml:"id"`
	Workflow     string      `json:"workflow" yaml:"workflow"`
	StepID       string      `json:"step_id" yaml:"step_id"`
	Title        string      `json:"title,omitempty" yaml:"title,omitempty"`
	Version      string      `json:"version,omitempty" yaml:"version,omitempty"`
	Status       string      `json:"status" yaml:"status"`
	StatusReason string      `json:"status_reason,omitempty" yaml:"status_reason,omitempty"`
	StartedAt    *time.Time  `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	FinishedAt   *time.Time  `json:"finished_at,omitempty" yaml:"finished_at,omitempty"`
	Errors       []StepError `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// StepError is one error a failed step reported.
type StepError struct {
	Code    int    `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
}

// StepStatuses lists every status a step can have in the summary, as
// build-log-service defines them (domain.StepStatus). Besides the results the
// Bitrise CLI reports, a step is started while it runs, and when the build is
// aborted, the running step becomes aborted and the steps after it not_started.
var StepStatuses = []string{
	"not_started", "started", "success", "failed", "failed_skippable", "skipped", "skipped_with_run_if",
	"preparation_failed", "aborted", "aborted_with_custom_timeout", "aborted_with_no_output",
}

// FailedStepStatuses are the statuses the "failed" filter matches: every way
// a step can fail the build. failed_skippable is not one, the build went on,
// and neither is aborted, which means the build itself was aborted.
var FailedStepStatuses = []string{"failed", "preparation_failed", "aborted_with_custom_timeout", "aborted_with_no_output"}

// StepsResult is the steps of a build, in execution order. Complete is false
// until the build's log is archived, which happens a little after the build
// finishes: until then Items lists only the steps that have started so far.
type StepsResult struct {
	Items    []Step `json:"items" yaml:"items"`
	Complete bool   `json:"complete" yaml:"complete"`
}

// StepLog is a step run of a build with its log.
type StepLog struct {
	Step `yaml:",inline"`
	Log  string `json:"log" yaml:"log"`
}

// StepLogsResult is the steps of a build with their logs, in execution order.
// Complete is false until the build's log is archived; until then the logs
// are empty, and Items lists only the steps that have started so far.
type StepLogsResult struct {
	Items    []StepLog `json:"items" yaml:"items"`
	Complete bool      `json:"complete" yaml:"complete"`
}

// LogResult is a build's full log. Complete is false until the log is
// archived; until then Log holds only what has been logged so far.
type LogResult struct {
	Log      string `json:"log" yaml:"log"`
	Complete bool   `json:"complete" yaml:"complete"`
}

// Service exposes build operations to the cmd layer.
type Service struct {
	client *bitriseapi.Client
}

// NewService returns a Service backed by the given API client. The client
// must be non-nil — every method makes a network call.
func NewService(client *bitriseapi.Client) *Service {
	return &Service{client: client}
}

// Trigger starts a new build for the given app + workflow.
// Endpoint: POST /apps/{app-slug}/builds.
func (s *Service) Trigger(ctx context.Context, req TriggerRequest) (Build, error) {
	if s.client == nil {
		return Build{}, fmt.Errorf("API client not configured")
	}
	if req.AppSlug == "" {
		return Build{}, fmt.Errorf("app ID is required")
	}
	envs := make([]bitriseapi.TriggerBuildEnv, 0, len(req.Environments))
	for _, e := range req.Environments {
		envs = append(envs, bitriseapi.TriggerBuildEnv{MappedTo: e.Key, Value: e.Value, IsExpand: e.IsExpand})
	}
	resp, err := s.client.TriggerBuild(ctx, req.AppSlug, bitriseapi.TriggerBuildRequest{
		HookInfo: bitriseapi.TriggerBuildHookInfo{Type: "bitrise"},
		BuildParams: bitriseapi.TriggerBuildParams{
			WorkflowID:    req.Workflow,
			PipelineID:    req.Pipeline,
			Branch:        req.Branch,
			BranchDest:    req.BranchDest,
			Tag:           req.Tag,
			CommitHash:    req.CommitHash,
			CommitMessage: req.CommitMessage,
			PullRequestID: req.PullRequestID,
			Priority:      req.Priority,
			Stack:         req.Stack,
			MachineTypeID: req.MachineTypeID,
			LicensePoolID: req.LicensePoolID,
			Environments:  envs,
		},
	})
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return Build{}, fmt.Errorf("app %q not found", req.AppSlug)
		}
		return Build{}, err
	}
	// The response carries none of the source fields, so echo what was sent.
	b := triggerRespToBuild(resp, req.AppSlug)
	b.Branch = req.Branch
	b.Tag = req.Tag
	b.CommitHash = req.CommitHash
	b.CommitMessage = req.CommitMessage
	if b.Workflow == "" {
		b.Workflow = req.Workflow
	}
	return b, nil
}

// List returns one page of builds for an app.
// Endpoint: GET /apps/{app-slug}/builds.
func (s *Service) List(ctx context.Context, opts ListOptions) (ListResult, error) {
	if s.client == nil {
		return ListResult{}, fmt.Errorf("API client not configured")
	}
	if opts.AppSlug == "" {
		return ListResult{}, fmt.Errorf("app is required")
	}
	statusInt, err := parseStatusFilter(opts.Status)
	if err != nil {
		return ListResult{}, err
	}
	apiOpts := bitriseapi.BuildsListOptions{
		SortBy:           opts.SortBy,
		Branch:           opts.Branch,
		Workflow:         opts.Workflow,
		CommitMessage:    opts.CommitMessage,
		TriggerEventType: opts.TriggerEventType,
		PullRequestID:    opts.PullRequestID,
		BuildNumber:      opts.BuildNumber,
		Status:           statusInt,
		IsPipelineBuild:  opts.IsPipelineBuild,
		Limit:            opts.Limit,
		Next:             opts.Cursor,
	}
	if opts.After != nil {
		apiOpts.After = int(opts.After.Unix())
	}
	if opts.Before != nil {
		apiOpts.Before = int(opts.Before.Unix())
	}
	builds, nextCursor, err := s.client.Builds(ctx, opts.AppSlug, apiOpts)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return ListResult{}, fmt.Errorf("app %q not found", opts.AppSlug)
		}
		return ListResult{}, err
	}
	items := make([]Build, 0, len(builds))
	for _, b := range builds {
		items = append(items, fromAPI(b, opts.AppSlug))
	}
	return ListResult{Items: items, NextCursor: nextCursor}, nil
}

// View returns details for a single build.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}.
func (s *Service) View(ctx context.Context, appSlug, buildSlug string) (Build, error) {
	if s.client == nil {
		return Build{}, fmt.Errorf("API client not configured")
	}
	if appSlug == "" {
		return Build{}, fmt.Errorf("app ID is required")
	}
	if buildSlug == "" {
		return Build{}, fmt.Errorf("build ID is required")
	}
	b, err := s.client.Build(ctx, appSlug, buildSlug)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return Build{}, fmt.Errorf("build %q not found", buildSlug)
		}
		return Build{}, err
	}
	return fromAPI(b, appSlug), nil
}

// Log streams the build log for the given build to w. For finished builds
// the full archived log is streamed; for in-progress builds the chunks
// available so far are written. It reports whether the log is complete, which
// it is once archived, a little after the build finishes.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}/log.
func (s *Service) Log(ctx context.Context, appSlug, buildSlug string, w io.Writer) (bool, error) {
	if s.client == nil {
		return false, fmt.Errorf("API client not configured")
	}
	if appSlug == "" {
		return false, fmt.Errorf("app ID is required")
	}
	if buildSlug == "" {
		return false, fmt.Errorf("build ID is required")
	}
	manifest, err := s.client.BuildLog(ctx, appSlug, buildSlug, w)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return false, fmt.Errorf("build %q not found", buildSlug)
		}
		return false, err
	}
	return manifest.IsArchived || manifest.ExpiringRawLogURL != "", nil
}

// LogText returns the build log Log would stream, keeping only its last tail
// lines when tail > 0.
func (s *Service) LogText(ctx context.Context, appSlug, buildSlug string, tail int) (LogResult, error) {
	if tail < 0 {
		return LogResult{}, fmt.Errorf("tail must not be negative")
	}
	buf := newLogBuffer(tail)
	complete, err := s.Log(ctx, appSlug, buildSlug, buf)
	if err != nil {
		return LogResult{}, err
	}
	return LogResult{Log: buf.String(), Complete: complete}, nil
}

// Steps returns the steps of a build across all its workflows, in execution
// order. A non-empty status keeps only the steps with that status, except
// "failed", which keeps every step that failed the build (FailedStepStatuses).
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}/log/summary.
func (s *Service) Steps(ctx context.Context, appSlug, buildSlug, status string) (StepsResult, error) {
	return s.steps(ctx, appSlug, buildSlug, status, nil)
}

// StepLogs returns the steps of a build Steps would, narrowed to stepIDs (step
// run UUIDs) when non-empty, each with its log. tail > 0 keeps only the last
// tail lines of each log. Until the build's log is archived the step logs are
// not served, so every log is empty and Complete is false.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}/log/steps/{step-uuid}.
func (s *Service) StepLogs(ctx context.Context, appSlug, buildSlug, status string, stepIDs []string, tail int) (StepLogsResult, error) {
	if tail < 0 {
		return StepLogsResult{}, fmt.Errorf("tail must not be negative")
	}
	steps, err := s.steps(ctx, appSlug, buildSlug, status, stepIDs)
	if err != nil {
		return StepLogsResult{}, err
	}
	res := StepLogsResult{Items: make([]StepLog, 0, len(steps.Items)), Complete: steps.Complete}
	for _, st := range steps.Items {
		buf := newLogBuffer(tail)
		if steps.Complete {
			if err := s.client.BuildStepLog(ctx, appSlug, buildSlug, st.ID, buf); err != nil {
				return StepLogsResult{}, fmt.Errorf("fetch log of step %q: %w", st.ID, err)
			}
		}
		res.Items = append(res.Items, StepLog{Step: st, Log: buf.String()})
	}
	return res, nil
}

// logArchiveTimeout bounds WaitForLogArchive. Archiving normally takes a few
// seconds after the build finishes; a log still not archived after this is
// unlikely to be soon.
var logArchiveTimeout = time.Minute

// WaitForLogArchive polls the build's log summary every interval until the
// log is archived, which is when StepLogs can return the step logs, calling
// waiting once if the first poll finds it not archived yet. It gives up
// without an error after logArchiveTimeout, leaving StepLogs to report the
// log as incomplete, and so does a build without a summary, leaving the error
// to StepLogs.
func (s *Service) WaitForLogArchive(ctx context.Context, appSlug, buildSlug string, interval time.Duration, waiting func()) error {
	if s.client == nil {
		return fmt.Errorf("API client not configured")
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	// The timeout also bounds an in-flight request, so a slow response can't
	// stretch the wait or turn giving up into an error.
	dctx, cancel := context.WithTimeout(ctx, logArchiveTimeout)
	defer cancel()
	for first := true; ; first = false {
		summary, err := s.client.BuildLogSummary(dctx, appSlug, buildSlug)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if dctx.Err() != nil {
				return nil
			}
			var apiErr *bitriseapi.APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		if summary.IsLogArchived {
			return nil
		}
		if first && waiting != nil {
			waiting()
		}
		select {
		case <-dctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (s *Service) steps(ctx context.Context, appSlug, buildSlug, status string, stepIDs []string) (StepsResult, error) {
	if s.client == nil {
		return StepsResult{}, fmt.Errorf("API client not configured")
	}
	if appSlug == "" {
		return StepsResult{}, fmt.Errorf("app ID is required")
	}
	if buildSlug == "" {
		return StepsResult{}, fmt.Errorf("build ID is required")
	}
	if err := ValidateStepStatus(status); err != nil {
		return StepsResult{}, err
	}
	summary, err := s.client.BuildLogSummary(ctx, appSlug, buildSlug)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			return StepsResult{}, err
		}
		// A build that has not started running yet has no summary either, so
		// the 404 alone doesn't say whether the build exists.
		b, buildErr := s.client.Build(ctx, appSlug, buildSlug)
		if buildErr != nil {
			if errors.As(buildErr, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				return StepsResult{}, fmt.Errorf("build %q not found", buildSlug)
			}
			return StepsResult{}, fmt.Errorf("no step summary for build %q, and looking up the build failed: %w", buildSlug, buildErr)
		}
		if statusString(b.Status) == "in-progress" {
			return StepsResult{Items: []Step{}}, nil
		}
		return StepsResult{}, fmt.Errorf("build %q has no step summary", buildSlug)
	}
	items := []Step{}
	found := map[string]bool{}
	for _, wf := range summary.Execution.Workflows {
		for _, st := range wf.Steps {
			if len(stepIDs) > 0 {
				if !slices.Contains(stepIDs, st.UUID) {
					continue
				}
				found[st.UUID] = true
			}
			matches := status == "" || st.Status == status ||
				(status == "failed" && slices.Contains(FailedStepStatuses, st.Status))
			if !matches {
				continue
			}
			items = append(items, stepFromAPI(st, wf.WorkflowID))
		}
	}
	// A step that has not started yet is missing from an unarchived summary,
	// so only a final summary proves the ID wrong.
	if summary.IsLogArchived {
		for _, id := range stepIDs {
			if !found[id] {
				return StepsResult{}, fmt.Errorf("step %q not found in build %q", id, buildSlug)
			}
		}
	}
	return StepsResult{Items: items, Complete: summary.IsLogArchived}, nil
}

// ValidateStepStatus checks a step status filter: empty, or one of
// StepStatuses.
func ValidateStepStatus(status string) error {
	if status != "" && !slices.Contains(StepStatuses, status) {
		return fmt.Errorf("unknown step status %q (expected: %s)", status, strings.Join(StepStatuses, ", "))
	}
	return nil
}

// AbortRequest describes a build-abort operation.
type AbortRequest struct {
	AppSlug             string
	BuildSlug           string
	Reason              string
	AbortWithSuccess    bool
	SkipGitStatusReport bool
	SkipNotifications   bool
}

// AbortResult is the CLI-facing result of aborting a build.
type AbortResult struct {
	AppSlug   string `json:"app_id" yaml:"app_id"`
	BuildSlug string `json:"build_id" yaml:"build_id"`
	Status    string `json:"status" yaml:"status"`
}

// Abort stops a running or queued build.
// Endpoint: POST /apps/{app-slug}/builds/{build-slug}/abort.
func (s *Service) Abort(ctx context.Context, req AbortRequest) (AbortResult, error) {
	if s.client == nil {
		return AbortResult{}, fmt.Errorf("API client not configured")
	}
	if req.AppSlug == "" {
		return AbortResult{}, fmt.Errorf("app ID is required")
	}
	if req.BuildSlug == "" {
		return AbortResult{}, fmt.Errorf("build ID is required")
	}
	resp, err := s.client.AbortBuild(ctx, req.AppSlug, req.BuildSlug, bitriseapi.AbortBuildRequest{
		AbortReason:         req.Reason,
		AbortWithSuccess:    req.AbortWithSuccess,
		SkipGitStatusReport: req.SkipGitStatusReport,
		SkipNotifications:   req.SkipNotifications,
	})
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return AbortResult{}, fmt.Errorf("build %q not found", req.BuildSlug)
		}
		return AbortResult{}, err
	}
	return AbortResult{
		AppSlug:   req.AppSlug,
		BuildSlug: req.BuildSlug,
		Status:    resp.Status,
	}, nil
}

// RebuildRequest describes a build-rebuild operation.
type RebuildRequest struct {
	AppSlug      string
	BuildSlug    string
	RemoteAccess bool
}

// Rebuild starts a new build with the parameters and workflow of a finished
// build. The response has no branch, tag or commit, so the returned Build
// leaves them empty.
// Endpoint: POST /apps/{app-slug}/builds/{build-slug}/rebuild.
func (s *Service) Rebuild(ctx context.Context, req RebuildRequest) (Build, error) {
	if s.client == nil {
		return Build{}, fmt.Errorf("API client not configured")
	}
	if req.AppSlug == "" {
		return Build{}, fmt.Errorf("app ID is required")
	}
	if req.BuildSlug == "" {
		return Build{}, fmt.Errorf("build ID is required")
	}
	resp, err := s.client.RebuildBuild(ctx, req.AppSlug, req.BuildSlug, bitriseapi.RebuildBuildRequest{IsRemote: req.RemoteAccess})
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return Build{}, fmt.Errorf("build %q not found", req.BuildSlug)
		}
		return Build{}, err
	}
	return triggerRespToBuild(resp, req.AppSlug), nil
}

// fromAPI maps a bitriseapi.Build (wire shape) into the CLI's Build type.
// appSlug comes from the request, not the response, since the API doesn't
// echo it back.
func fromAPI(b bitriseapi.Build, appSlug string) Build {
	out := Build{
		Slug:                    b.Slug,
		AppSlug:                 appSlug,
		BuildNumber:             b.BuildNumber,
		Status:                  statusString(b.Status),
		StatusText:              b.StatusText,
		AbortReason:             b.AbortReason,
		IsOnHold:                b.IsOnHold,
		Rebuildable:             b.Rebuildable,
		Workflow:                b.TriggeredWorkflow,
		PipelineWorkflowID:      b.PipelineWorkflowID,
		Branch:                  b.Branch,
		Tag:                     b.Tag,
		PullRequestID:           b.PullRequestID,
		PullRequestTargetBranch: b.PullRequestTargetBranch,
		PullRequestViewURL:      b.PullRequestViewURL,
		CommitHash:              b.CommitHash,
		CommitMessage:           b.CommitMessage,
		TriggeredBy:             b.TriggeredBy,
		StackIdentifier:         b.StackIdentifier,
		MachineTypeID:           b.MachineTypeID,
		CreditCost:              b.CreditCost,
	}
	if !b.TriggeredAt.IsZero() {
		out.TriggeredAt = b.TriggeredAt.UTC()
	}
	if !b.FinishedAt.IsZero() {
		t := b.FinishedAt.UTC()
		out.FinishedAt = &t
	}
	return out
}

func stepFromAPI(st bitriseapi.BuildExecutionStep, workflow string) Step {
	out := Step{
		ID:           st.UUID,
		Workflow:     workflow,
		StepID:       st.StepID,
		Title:        st.Title,
		Version:      st.Version,
		Status:       st.Status,
		StatusReason: st.StatusReason,
	}
	if !st.StartedAt.IsZero() {
		t := st.StartedAt.UTC()
		out.StartedAt = &t
	}
	if !st.FinishedAt.IsZero() {
		t := st.FinishedAt.UTC()
		out.FinishedAt = &t
	}
	for _, e := range st.Errors {
		out.Errors = append(out.Errors, StepError{Code: e.Code, Message: e.Message})
	}
	return out
}

// triggerRespToBuild collapses the trigger (or rebuild) response into our
// Build shape, preferring the modern Results[0] over the deprecated top-level
// fields.
func triggerRespToBuild(resp bitriseapi.TriggerBuildResponse, appSlug string) Build {
	slug := resp.BuildSlug
	number := resp.BuildNumber
	url := resp.BuildURL
	workflow := resp.TriggeredWorkflow
	if len(resp.Results) > 0 {
		r := resp.Results[0]
		if r.BuildSlug != "" {
			slug = r.BuildSlug
		}
		if r.BuildNumber != 0 {
			number = r.BuildNumber
		}
		if r.BuildURL != "" {
			url = r.BuildURL
		}
		if r.TriggeredWorkflow != "" {
			workflow = r.TriggeredWorkflow
		}
	}
	return Build{
		Slug:        slug,
		AppSlug:     appSlug,
		BuildNumber: number,
		Status:      "in-progress",
		StatusText:  resp.Message,
		Workflow:    workflow,
		TriggeredAt: time.Now().UTC(),
		BuildURL:    url,
	}
}

// statusString translates the API's integer status into a stable string for
// --format json/yml. Unknown values fall through as the integer for
// forward-compat with new statuses.
func statusString(n int) string {
	switch n {
	case 0:
		return "in-progress"
	case 1:
		return "success"
	case 2:
		return "failed"
	case 3:
		return "aborted"
	case 4:
		return "aborted-with-success"
	default:
		return strconv.Itoa(n)
	}
}

// parseStatusFilter is the inverse of statusString. Returns a nil pointer
// (not an error) when s is empty, signaling "no filter requested".
func parseStatusFilter(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	var n int
	switch s {
	case "in-progress":
		n = 0
	case "success":
		n = 1
	case "failed":
		n = 2
	case "aborted":
		n = 3
	case "aborted-with-success":
		n = 4
	default:
		return nil, fmt.Errorf("unknown build status %q (expected: in-progress, success, failed, aborted, aborted-with-success)", s)
	}
	return &n, nil
}

// newLogBuffer returns a buffer that keeps everything written to it, or only
// the last tail lines when tail > 0.
func newLogBuffer(tail int) interface {
	io.Writer
	fmt.Stringer
} {
	if tail > 0 {
		return &tailBuffer{n: tail}
	}
	return &strings.Builder{}
}

// tailBuffer keeps the last n lines written to it, so a long log is never held
// whole. A trailing newline ends the last line rather than starting one.
type tailBuffer struct {
	n int
	// lines holds the last complete lines, newline included, as a ring
	// starting at next once it has n lines.
	lines []string
	next  int
	cur   []byte
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	written := len(p)
	for {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			t.cur = append(t.cur, p...)
			return written, nil
		}
		line := string(append(t.cur, p[:i+1]...))
		t.cur = t.cur[:0]
		p = p[i+1:]
		if len(t.lines) < t.n {
			t.lines = append(t.lines, line)
			continue
		}
		t.lines[t.next] = line
		t.next = (t.next + 1) % t.n
	}
}

func (t *tailBuffer) String() string {
	lines := append(slices.Clone(t.lines[t.next:]), t.lines[:t.next]...)
	if len(t.cur) > 0 && len(lines) == t.n {
		lines = lines[1:]
	}
	return strings.Join(lines, "") + string(t.cur)
}
