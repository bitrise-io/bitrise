package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
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
	req, err := c.newRequest(ctx, "/organizations", nil)
	if err != nil {
		return nil, err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []Organization `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode organizations response: %w", err)
	}
	return envelope.Data, nil
}

// Organization returns a single workspace the authenticated user belongs to.
// Endpoint: GET /organizations/{org-slug}.
func (c *Client) Organization(ctx context.Context, orgSlug string) (Organization, error) {
	return getEnvelope[Organization](ctx, c, "/organizations/"+url.PathEscape(orgSlug), nil)
}
