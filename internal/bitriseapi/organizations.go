package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
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
