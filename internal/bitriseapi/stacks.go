package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
)

// StackInfo is the wire-format stack record returned by the available-stacks
// endpoints. Field names with hyphens match the Bitrise API's JSON keys.
type StackInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	OS          string `json:"os"`
	OSVersion   int    `json:"os_version,omitempty"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	StackReport string `json:"stack-report,omitempty"`
	RemovalDate string `json:"removal-date,omitempty"`
}

// AvailableStacks lists available stacks and their machine configurations.
// When orgSlug is non-empty, the org-scoped endpoint is used:
//
//	GET /organizations/{org-slug}/available-stacks
//
// Otherwise the global endpoint is used:
//
//	GET /available-stacks
//
// The response is a map of stack ID → StackInfo.
func (c *Client) AvailableStacks(ctx context.Context, orgSlug string) (map[string]StackInfo, error) {
	path := "/available-stacks"
	if orgSlug != "" {
		path = "/organizations/" + orgSlug + "/available-stacks"
	}

	req, err := c.newRequest(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var result map[string]StackInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode stacks response: %w", err)
	}
	return result, nil
}
