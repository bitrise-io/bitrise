package bitriseapi

import (
	"context"
	"net/url"
	"strconv"
)

// App is a Bitrise app (project), as returned by GET /apps and
// GET /apps/{app-slug}.
type App struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Provider    string   `json:"provider"`
	RepoURL     string   `json:"repo_url"`
	ProjectType string   `json:"project_type,omitempty"`
	IsDisabled  bool     `json:"is_disabled"`
	Owner       AppOwner `json:"owner"`
}

// AppOwner is the workspace or user that owns an app.
type AppOwner struct {
	Slug string `json:"slug"`
}

// AppsListOptions filters and paginates GET /apps. All fields are optional.
type AppsListOptions struct {
	SortBy      string // "created_at" or "last_build_at"
	Next        string // pagination cursor from a previous page's response
	Limit       int    // server default (50) when 0
	Title       string
	ProjectType string
}

func (o AppsListOptions) params() url.Values {
	p := url.Values{}
	if o.SortBy != "" {
		p.Set("sort_by", o.SortBy)
	}
	if o.Next != "" {
		p.Set("next", o.Next)
	}
	if o.Limit > 0 {
		p.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Title != "" {
		p.Set("title", o.Title)
	}
	if o.ProjectType != "" {
		p.Set("project_type", o.ProjectType)
	}
	return p
}

// Apps returns one page of apps the authenticated user can access, and the
// cursor for the next page ("" when there isn't one).
// Endpoint: GET /apps.
func (c *Client) Apps(ctx context.Context, opts AppsListOptions) ([]App, string, error) {
	return getPage[App](ctx, c, "/apps", opts.params())
}

// App returns the details of a single app by slug.
// Endpoint: GET /apps/{app-slug}.
func (c *Client) App(ctx context.Context, appSlug string) (App, error) {
	return getEnvelope[App](ctx, c, "/apps/"+url.PathEscape(appSlug), nil)
}

// RegisterAppRequest is the body of POST /apps/register.
//
// IsPublic has no omitempty so the JSON always carries is_public:false when
// not opted in, matching the website's "Add new app" flow. FlowType is the
// analytics attribution string the server reads out of params[:flow_type].
type RegisterAppRequest struct {
	RepoURL           string `json:"repo_url"`
	OrganizationSlug  string `json:"organization_slug"`
	Provider          string `json:"provider"`
	IsPublic          bool   `json:"is_public"`
	Title             string `json:"title,omitempty"`
	DefaultBranchName string `json:"default_branch_name,omitempty"`
	FlowType          string `json:"flow_type,omitempty"`
}

// RegisterAppResponse is the response from POST /apps/register. The app is
// in status=-1 (setup-incomplete) until FinishApp is called.
type RegisterAppResponse struct {
	Slug string `json:"slug"`
}

// RegisterApp creates a new app on Bitrise.
// Endpoint: POST /apps/register.
func (c *Client) RegisterApp(ctx context.Context, req RegisterAppRequest) (RegisterAppResponse, error) {
	return postDecode[RegisterAppResponse](ctx, c, "/apps/register", nil, req)
}

// FinishAppRequest is the body of POST /apps/{app-slug}/finish.
type FinishAppRequest struct {
	StackID     string `json:"stack_id"`
	Mode        string `json:"mode"`
	ProjectType string `json:"project_type,omitempty"`
	Config      string `json:"config,omitempty"`
	FlowType    string `json:"flow_type,omitempty"`
}

// FinishAppResponse is the response from POST /apps/{app-slug}/finish.
type FinishAppResponse struct {
	BuildTriggerToken string `json:"build_trigger_token"`
	BranchName        string `json:"branch_name"`
}

// FinishApp activates a registered app, returning its build trigger token.
// Endpoint: POST /apps/{app-slug}/finish.
func (c *Client) FinishApp(ctx context.Context, appSlug string, req FinishAppRequest) (FinishAppResponse, error) {
	return postDecode[FinishAppResponse](ctx, c, "/apps/"+url.PathEscape(appSlug)+"/finish", nil, req)
}
