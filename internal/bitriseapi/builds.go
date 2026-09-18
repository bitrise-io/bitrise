package bitriseapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"
)

// Build is the wire-format build record returned by the Bitrise API. Many
// fields are populated only for finished builds; treat empty values as "not
// yet set".
type Build struct {
	Slug                    string    `json:"slug"`
	BuildNumber             int       `json:"build_number"`
	Status                  int       `json:"status"`
	StatusText              string    `json:"status_text,omitempty"`
	AbortReason             string    `json:"abort_reason,omitempty"`
	Branch                  string    `json:"branch,omitempty"`
	TriggeredWorkflow       string    `json:"triggered_workflow,omitempty"`
	PipelineWorkflowID      string    `json:"pipeline_workflow_id,omitempty"`
	CommitHash              string    `json:"commit_hash,omitempty"`
	CommitMessage           string    `json:"commit_message,omitempty"`
	Tag                     string    `json:"tag,omitempty"`
	PullRequestID           int       `json:"pull_request_id,omitempty"`
	PullRequestTargetBranch string    `json:"pull_request_target_branch,omitempty"`
	PullRequestViewURL      string    `json:"pull_request_view_url,omitempty"`
	TriggeredAt             time.Time `json:"triggered_at,omitempty"`
	FinishedAt              time.Time `json:"finished_at,omitempty"`
	TriggeredBy             string    `json:"triggered_by,omitempty"`
	StackIdentifier         string    `json:"stack_identifier,omitempty"`
	MachineTypeID           string    `json:"machine_type_id,omitempty"`
	CreditCost              int       `json:"credit_cost,omitempty"`
	IsOnHold                bool      `json:"is_on_hold,omitempty"`
	Rebuildable             bool      `json:"rebuildable,omitempty"`
}

// BuildsListOptions filters and paginates GET /apps/{app-slug}/builds. All
// fields are optional. Field names match the API's query parameters.
type BuildsListOptions struct {
	SortBy           string // "created_at" (default) or "running_first"
	Branch           string
	Workflow         string
	CommitMessage    string
	TriggerEventType string // push, pull-request, tag
	PullRequestID    int
	BuildNumber      int
	After            int // unix timestamp
	Before           int // unix timestamp
	// Status is the integer status filter the API accepts: 0=in-progress,
	// 1=success, 2=failed, 3=aborted, 4=aborted-with-success. Tri-state
	// pointer because 0 is a meaningful filter value, not "unset".
	Status          *int
	IsPipelineBuild *bool // tri-state: nil=unset, true/false sent literally
	Next            string
	Limit           int
}

func (o BuildsListOptions) params() url.Values {
	p := url.Values{}
	if o.SortBy != "" {
		p.Set("sort_by", o.SortBy)
	}
	if o.Branch != "" {
		p.Set("branch", o.Branch)
	}
	if o.Workflow != "" {
		p.Set("workflow", o.Workflow)
	}
	if o.CommitMessage != "" {
		p.Set("commit_message", o.CommitMessage)
	}
	if o.TriggerEventType != "" {
		p.Set("trigger_event_type", o.TriggerEventType)
	}
	if o.PullRequestID > 0 {
		p.Set("pull_request_id", strconv.Itoa(o.PullRequestID))
	}
	if o.BuildNumber > 0 {
		p.Set("build_number", strconv.Itoa(o.BuildNumber))
	}
	if o.After > 0 {
		p.Set("after", strconv.Itoa(o.After))
	}
	if o.Before > 0 {
		p.Set("before", strconv.Itoa(o.Before))
	}
	if o.Status != nil {
		p.Set("status", strconv.Itoa(*o.Status))
	}
	if o.IsPipelineBuild != nil {
		p.Set("is_pipeline_build", strconv.FormatBool(*o.IsPipelineBuild))
	}
	if o.Next != "" {
		p.Set("next", o.Next)
	}
	if o.Limit > 0 {
		p.Set("limit", strconv.Itoa(o.Limit))
	}
	return p
}

// Builds returns one page of builds for an app, and the cursor for the next
// page ("" when there isn't one).
// Endpoint: GET /apps/{app-slug}/builds.
func (c *Client) Builds(ctx context.Context, appSlug string, opts BuildsListOptions) ([]Build, string, error) {
	return getPage[Build](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/builds", opts.params())
}

// Build returns details for a single build.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}.
func (c *Client) Build(ctx context.Context, appSlug, buildSlug string) (Build, error) {
	return getEnvelope[Build](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/builds/"+url.PathEscape(buildSlug), nil)
}

// TriggerBuildRequest is the JSON body sent to POST /apps/{app-slug}/builds.
type TriggerBuildRequest struct {
	HookInfo    TriggerBuildHookInfo `json:"hook_info"`
	BuildParams TriggerBuildParams   `json:"build_params"`
}

// TriggerBuildHookInfo identifies the trigger source. CLI-initiated builds
// use type="bitrise".
type TriggerBuildHookInfo struct {
	Type string `json:"type"`
}

// TriggerBuildEnv is a single environment variable injected into a triggered
// build.
type TriggerBuildEnv struct {
	MappedTo string `json:"mapped_to"`
	Value    string `json:"value"`
	IsExpand bool   `json:"is_expand"`
}

// TriggerBuildParams holds the build-specific parameters of a trigger
// request. Most fields are optional; the API derives defaults from the
// app's trigger map and bitrise.yml when omitted.
type TriggerBuildParams struct {
	WorkflowID    string            `json:"workflow_id,omitempty"`
	PipelineID    string            `json:"pipeline_id,omitempty"`
	Branch        string            `json:"branch,omitempty"`
	BranchDest    string            `json:"branch_dest,omitempty"`
	Tag           string            `json:"tag,omitempty"`
	CommitHash    string            `json:"commit_hash,omitempty"`
	CommitMessage string            `json:"commit_message,omitempty"`
	PullRequestID int               `json:"pull_request_id,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	Environments  []TriggerBuildEnv `json:"environments,omitempty"`
}

// TriggerBuildResponse is the response from POST /apps/{app-slug}/builds.
// The deprecated top-level fields and the modern Results array carry
// overlapping information; for a single-result trigger, the top-level
// fields are used. Results is the source of truth when the API splits a
// pipeline trigger across multiple workflows.
type TriggerBuildResponse struct {
	BuildSlug         string                     `json:"build_slug,omitempty"`
	BuildNumber       int                        `json:"build_number,omitempty"`
	BuildURL          string                     `json:"build_url,omitempty"`
	TriggeredWorkflow string                     `json:"triggered_workflow,omitempty"`
	Service           string                     `json:"service,omitempty"`
	Status            string                     `json:"status,omitempty"`
	Message           string                     `json:"message,omitempty"`
	Slug              string                     `json:"slug,omitempty"`
	Results           []TriggerBuildResponseItem `json:"results,omitempty"`
}

// TriggerBuildResponseItem is one entry from TriggerBuildResponse.Results.
type TriggerBuildResponseItem struct {
	BuildSlug         string `json:"build_slug,omitempty"`
	BuildNumber       int    `json:"build_number,omitempty"`
	BuildURL          string `json:"build_url,omitempty"`
	TriggeredWorkflow string `json:"triggered_workflow,omitempty"`
	TriggeredPipeline string `json:"triggered_pipeline,omitempty"`
	Status            string `json:"status,omitempty"`
	Message           string `json:"message,omitempty"`
}

// TriggerBuild starts a new build.
// Endpoint: POST /apps/{app-slug}/builds.
func (c *Client) TriggerBuild(ctx context.Context, appSlug string, req TriggerBuildRequest) (TriggerBuildResponse, error) {
	return postDecode[TriggerBuildResponse](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/builds", nil, req)
}

// AbortBuildRequest is the JSON body for POST /apps/{app-slug}/builds/{build-slug}/abort.
type AbortBuildRequest struct {
	AbortReason         string `json:"abort_reason,omitempty"`
	AbortWithSuccess    bool   `json:"abort_with_success,omitempty"`
	SkipGitStatusReport bool   `json:"skip_git_status_report,omitempty"`
	SkipNotifications   bool   `json:"skip_notifications,omitempty"`
}

// AbortBuildResponse is the response from POST /apps/{app-slug}/builds/{build-slug}/abort.
type AbortBuildResponse struct {
	Status string `json:"status"`
}

// AbortBuild stops a running or queued build.
// Endpoint: POST /apps/{app-slug}/builds/{build-slug}/abort.
func (c *Client) AbortBuild(ctx context.Context, appSlug, buildSlug string, req AbortBuildRequest) (AbortBuildResponse, error) {
	return postDecode[AbortBuildResponse](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/builds/"+url.PathEscape(buildSlug)+"/abort", nil, req)
}

// BuildLogResponse is the JSON returned by GET /apps/{app-slug}/builds/{build-slug}/log.
//
// Behavior depends on whether the build is finished: for archived builds,
// ExpiringRawLogURL points at a time-limited URL (typically S3) holding the
// full log and LogChunks may be empty. For in-progress builds, LogChunks
// holds the new chunks since the requested afterTimestamp, and
// NextAfterTimestamp carries the value to pass on the next poll.
type BuildLogResponse struct {
	ExpiringRawLogURL  string          `json:"expiring_raw_log_url,omitempty"`
	IsArchived         bool            `json:"is_archived"`
	LogChunks          []BuildLogChunk `json:"log_chunks,omitempty"`
	NextAfterTimestamp string          `json:"next_after_timestamp,omitempty"`
}

// BuildLogChunk is one piece of build log output.
type BuildLogChunk struct {
	Chunk    string `json:"chunk"`
	Position int    `json:"position"`
}

// BuildLogManifest fetches the log manifest for a build. afterTimestamp is
// sent as the after_timestamp query parameter to request only log chunks
// newer than that value; pass "" for the first call or a full log fetch.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}/log.
func (c *Client) BuildLogManifest(ctx context.Context, appSlug, buildSlug, afterTimestamp string) (BuildLogResponse, error) {
	var params url.Values
	if afterTimestamp != "" {
		params = url.Values{"after_timestamp": {afterTimestamp}}
	}
	return get[BuildLogResponse](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/builds/"+url.PathEscape(buildSlug)+"/log", params)
}

// BuildLog fetches the log for a build and writes it to w.
//
// For archived (finished) builds, the full log is downloaded from
// ExpiringRawLogURL. For in-progress builds, whatever chunks are currently
// available are written — the caller can re-run to see more. Returns the
// manifest so callers can inspect IsArchived (e.g. to print a "still
// running" hint).
func (c *Client) BuildLog(ctx context.Context, appSlug, buildSlug string, w io.Writer) (BuildLogResponse, error) {
	manifest, err := c.BuildLogManifest(ctx, appSlug, buildSlug, "")
	if err != nil {
		return manifest, err
	}
	if manifest.ExpiringRawLogURL != "" {
		if err := c.streamRawLog(ctx, manifest.ExpiringRawLogURL, w); err != nil {
			return manifest, err
		}
		return manifest, nil
	}
	// The API collects chunks from parallel producers and does not guarantee
	// order even within a single manifest response (Service.Watch reorders
	// across polls for the same reason). Clone so the returned manifest
	// keeps the server's original ordering.
	chunks := slices.Clone(manifest.LogChunks)
	slices.SortFunc(chunks, func(a, b BuildLogChunk) int { return a.Position - b.Position })
	for _, chunk := range chunks {
		if _, err := io.WriteString(w, chunk.Chunk); err != nil {
			return manifest, fmt.Errorf("write log chunk: %w", err)
		}
	}
	return manifest, nil
}

// streamRawLog GETs rawURL and copies the body to w. rawURL is expected to
// be a presigned URL from BuildLogResponse.ExpiringRawLogURL — no auth
// header is sent to it.
func (c *Client) streamRawLog(ctx context.Context, rawURL string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build log URL request: %w", err)
	}
	resp, err := c.streamHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("fetch raw log: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch raw log: HTTP %d", resp.StatusCode)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("stream raw log: %w", err)
	}
	return nil
}
