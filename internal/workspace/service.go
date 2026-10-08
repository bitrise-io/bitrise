package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
)

// Workspace is the CLI representation of a Bitrise workspace.
type Workspace struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

// WorkspacesResult holds the workspaces the authenticated user belongs to.
type WorkspacesResult struct {
	Items []Workspace `json:"items" yaml:"items"`
}

// Service exposes workspace operations to the cmd layer.
type Service struct {
	client *bitriseapi.Client
}

// NewService returns a Service backed by the given API client.
func NewService(client *bitriseapi.Client) *Service {
	return &Service{client: client}
}

// List returns the user's workspaces in the same order as the
// multiple-workspaces error and the interactive picker (see Sort).
func (s *Service) List(ctx context.Context) (WorkspacesResult, error) {
	orgs, err := s.client.Organizations(ctx)
	if err != nil {
		return WorkspacesResult{}, err
	}
	sorted := Sort(orgs)
	items := make([]Workspace, len(sorted))
	for i, o := range sorted {
		items[i] = fromAPI(o)
	}
	return WorkspacesResult{Items: items}, nil
}

// View returns a single workspace. The API answers 404 both for a workspace
// that doesn't exist and for one the user isn't a member of.
func (s *Service) View(ctx context.Context, workspaceSlug string) (Workspace, error) {
	o, err := s.client.Organization(ctx, workspaceSlug)
	if err != nil {
		var apiErr *bitriseapi.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return Workspace{}, fmt.Errorf("workspace %q not found", workspaceSlug)
		}
		return Workspace{}, err
	}
	return fromAPI(o), nil
}

func fromAPI(o bitriseapi.Organization) Workspace {
	return Workspace{ID: o.Slug, Name: o.Name}
}
