package bitriseapi

import (
	"context"
	"net/url"
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
		path = "/organizations/" + url.PathEscape(orgSlug) + "/available-stacks"
	}
	return get[map[string]StackInfo](ctx, c, path, nil)
}
