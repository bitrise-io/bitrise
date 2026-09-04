package rdeapi

import (
	"context"
	"fmt"
)

// PlatformUsage aggregates active sessions on one OS platform. The backend
// omits zero-valued fields from the JSON, so absent means 0.
type PlatformUsage struct {
	SessionCount int32 `json:"sessionCount"`
	VCPU         int32 `json:"vcpu"`
	MemoryGB     int32 `json:"memoryGb"`
}

// UsageTotals splits active-session usage by OS platform. Buckets are
// pointers because the backend may omit an empty bucket object entirely.
type UsageTotals struct {
	Linux   *PlatformUsage `json:"linux"`
	Macos   *PlatformUsage `json:"macos"`
	Unknown *PlatformUsage `json:"unknown"`
}

// UserUsage is one row of the per-user usage breakdown. The workspace-owned
// bucket row (IsWorkspace) carries no user identity.
type UserUsage struct {
	UserID      string       `json:"userId"`
	UserSlug    string       `json:"userSlug"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	IsWorkspace bool         `json:"isWorkspace"`
	Totals      *UsageTotals `json:"totals"`
}

// WorkspaceUsage is the workspace usage report: a point-in-time snapshot of
// the sessions currently consuming resources, split by OS and by user.
type WorkspaceUsage struct {
	Totals *UsageTotals `json:"totals"`
	Users  []UserUsage  `json:"users"`
	// UnknownMachineTypeCount is the number of active sessions whose machine
	// type had no resolvable vCPU/RAM spec; they contribute 0 to the sums, so
	// totals may undercount when this is non-zero.
	UnknownMachineTypeCount int32 `json:"unknownMachineTypeCount"`
}

// GetWorkspaceUsage returns the workspace's active-session usage snapshot.
// Requires the workspace's billing-view permission (workspace owners and
// billing-managing custom roles); other members get a permission error.
// Endpoint: GET /v1/workspaces/{workspaceId}/usage.
func (c *Client) GetWorkspaceUsage(ctx context.Context, workspaceID string) (WorkspaceUsage, error) {
	if workspaceID == "" {
		return WorkspaceUsage{}, fmt.Errorf("workspace ID is required")
	}
	var resp WorkspaceUsage
	if err := c.getJSON(ctx, wsPath(workspaceID, "/usage"), &resp); err != nil {
		return WorkspaceUsage{}, err
	}
	return resp, nil
}
