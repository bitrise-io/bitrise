package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
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

type appsPage struct {
	Data   []App `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

// Apps returns one page of apps the authenticated user can access, and the
// cursor for the next page ("" when there isn't one).
// Endpoint: GET /apps.
func (c *Client) Apps(ctx context.Context, opts AppsListOptions) ([]App, string, error) {
	req, err := c.newRequest(ctx, "/apps", opts.params())
	if err != nil {
		return nil, "", err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	var page appsPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, "", fmt.Errorf("decode apps response: %w", err)
	}
	return page.Data, page.Paging.Next, nil
}

// App returns the details of a single app by slug.
// Endpoint: GET /apps/{app-slug}.
func (c *Client) App(ctx context.Context, appSlug string) (App, error) {
	req, err := c.newRequest(ctx, "/apps/"+url.PathEscape(appSlug), nil)
	if err != nil {
		return App{}, err
	}
	body, err := c.do(req)
	if err != nil {
		return App{}, err
	}
	var envelope struct {
		Data App `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return App{}, fmt.Errorf("decode app response: %w", err)
	}
	return envelope.Data, nil
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
	body, err := c.post(ctx, "/apps/register", nil, req)
	if err != nil {
		return RegisterAppResponse{}, err
	}
	var resp RegisterAppResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return RegisterAppResponse{}, fmt.Errorf("decode register response: %w", err)
	}
	return resp, nil
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
	body, err := c.post(ctx, "/apps/"+url.PathEscape(appSlug)+"/finish", nil, req)
	if err != nil {
		return FinishAppResponse{}, err
	}
	var resp FinishAppResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return FinishAppResponse{}, fmt.Errorf("decode finish response: %w", err)
	}
	return resp, nil
}
