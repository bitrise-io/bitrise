// Package app holds the business-logic layer for app operations.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// App is the CLI representation of a Bitrise app (project).
type App struct {
	Slug        string `json:"id" yaml:"id"`
	Title       string `json:"title" yaml:"title"`
	Provider    string `json:"provider" yaml:"provider"`
	RepoURL     string `json:"repo_url" yaml:"repo_url"`
	OwnerSlug   string `json:"workspace_id,omitempty" yaml:"workspace_id,omitempty"`
	ProjectType string `json:"project_type,omitempty" yaml:"project_type,omitempty"`
	IsDisabled  bool   `json:"is_disabled,omitempty" yaml:"is_disabled,omitempty"`
}

// ListOptions paginates and filters app lists. Filter fields map to the
// query parameters of GET /apps.
type ListOptions struct {
	Limit       int
	Cursor      string
	SortBy      string
	Title       string
	ProjectType string
}

// AppsResult is one page of apps.
type AppsResult struct {
	Items      []App  `json:"items" yaml:"items"`
	NextCursor string `json:"next_cursor,omitempty" yaml:"next_cursor,omitempty"`
}

// Service exposes app operations to the cmd layer.
type Service struct {
	client *bitriseapi.Client
}

// NewService returns a Service backed by the given API client.
func NewService(client *bitriseapi.Client) *Service {
	return &Service{client: client}
}

// List returns one page of apps the authenticated user can access.
func (s *Service) List(ctx context.Context, opts ListOptions) (AppsResult, error) {
	apps, next, err := s.client.Apps(ctx, bitriseapi.AppsListOptions{
		SortBy:      opts.SortBy,
		Next:        opts.Cursor,
		Limit:       opts.Limit,
		Title:       opts.Title,
		ProjectType: opts.ProjectType,
	})
	if err != nil {
		return AppsResult{}, err
	}
	items := make([]App, 0, len(apps))
	for _, a := range apps {
		items = append(items, fromAPI(a))
	}
	return AppsResult{Items: items, NextCursor: next}, nil
}

// View returns details of a single app by slug.
func (s *Service) View(ctx context.Context, appSlug string) (App, error) {
	a, err := s.client.App(ctx, appSlug)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return App{}, fmt.Errorf("app %q not found", appSlug)
		}
		return App{}, err
	}
	return fromAPI(a), nil
}

func fromAPI(a bitriseapi.App) App {
	return App{
		Slug:        a.Slug,
		Title:       a.Title,
		Provider:    a.Provider,
		RepoURL:     a.RepoURL,
		OwnerSlug:   a.Owner.Slug,
		ProjectType: a.ProjectType,
		IsDisabled:  a.IsDisabled,
	}
}
