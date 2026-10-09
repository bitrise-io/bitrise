package bitriseapi

import (
	"context"
	"net/url"
)

// Organization is a workspace the authenticated user belongs to.
type Organization struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Organizations returns the workspaces the authenticated user can access.
// Endpoint: GET /organizations.
func (c *Client) Organizations(ctx context.Context) ([]Organization, error) {
	return getEnvelope[[]Organization](ctx, c, "/organizations", nil)
}

// Organization returns a single workspace the authenticated user belongs to.
// Endpoint: GET /organizations/{org-slug}.
func (c *Client) Organization(ctx context.Context, orgSlug string) (Organization, error) {
	return getEnvelope[Organization](ctx, c, "/organizations/"+url.PathEscape(orgSlug), nil)
}
