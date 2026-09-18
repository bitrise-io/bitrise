package rde

import (
	"context"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// PlatformUsage aggregates active sessions on one OS platform. Always dense:
// fields the backend omitted are 0.
type PlatformUsage struct {
	SessionCount int32 `json:"session_count" yaml:"session_count"`
	VCPU         int32 `json:"vcpu" yaml:"vcpu"`
	MemoryGB     int32 `json:"memory_gb" yaml:"memory_gb"`
}

// UsageTotals splits active-session usage by OS platform. Unknown holds
// sessions whose OS could not be determined.
type UsageTotals struct {
	Linux   PlatformUsage `json:"linux" yaml:"linux"`
	Macos   PlatformUsage `json:"macos" yaml:"macos"`
	Unknown PlatformUsage `json:"unknown" yaml:"unknown"`
}

// UserUsage is one row of the per-user usage breakdown. The single
// workspace-owned bucket row (IsWorkspace true) carries no user identity.
//
// UserID is the user's Bitrise ID (the wire's slug — the identifier other
// Bitrise surfaces use; the backend's internal user UUID is not exposed).
// Best-effort: may be empty, so detect the workspace bucket via IsWorkspace,
// never via an empty UserID.
type UserUsage struct {
	UserID      string      `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Email       string      `json:"email,omitempty" yaml:"email,omitempty"`
	Username    string      `json:"username,omitempty" yaml:"username,omitempty"`
	IsWorkspace bool        `json:"is_workspace,omitempty" yaml:"is_workspace,omitempty"`
	Totals      UsageTotals `json:"totals" yaml:"totals"`
}

// WorkspaceUsage is a point-in-time snapshot of the sessions currently
// consuming resources in the workspace, split by OS and by user. It is not a
// historical or billing-period report.
type WorkspaceUsage struct {
	Totals UsageTotals `json:"totals" yaml:"totals"`
	Users  []UserUsage `json:"users" yaml:"users"`
	// UnknownMachineTypeCount counts active sessions whose machine type had
	// no resolvable vCPU/RAM spec; they contribute 0 to the sums, so totals
	// may undercount when this is non-zero.
	UnknownMachineTypeCount int32 `json:"unknown_machine_type_count" yaml:"unknown_machine_type_count"`
}

// GetWorkspaceUsage returns the workspace's active-session usage snapshot.
// Requires the workspace's billing-view permission (workspace owners and
// billing-managing custom roles).
func (s *Service) GetWorkspaceUsage(ctx context.Context, workspaceID string) (WorkspaceUsage, error) {
	if s.client == nil {
		return WorkspaceUsage{}, errClient()
	}
	wire, err := s.client.GetWorkspaceUsage(ctx, workspaceID)
	if err != nil {
		return WorkspaceUsage{}, err
	}
	return usageFromAPI(wire), nil
}

// usageFromAPI densifies the wire report: the backend omits zero-valued
// fields and empty bucket objects, and the stable shape promises every
// bucket present with explicit zeros.
func usageFromAPI(w rdeapi.WorkspaceUsage) WorkspaceUsage {
	out := WorkspaceUsage{
		Totals:                  usageTotalsFromAPI(w.Totals),
		Users:                   make([]UserUsage, 0, len(w.Users)),
		UnknownMachineTypeCount: w.UnknownMachineTypeCount,
	}
	for _, u := range w.Users {
		out.Users = append(out.Users, UserUsage{
			UserID:      u.UserSlug,
			Email:       u.Email,
			Username:    u.Username,
			IsWorkspace: u.IsWorkspace,
			Totals:      usageTotalsFromAPI(u.Totals),
		})
	}
	return out
}

func usageTotalsFromAPI(w *rdeapi.UsageTotals) UsageTotals {
	if w == nil {
		return UsageTotals{}
	}
	return UsageTotals{
		Linux:   platformUsageFromAPI(w.Linux),
		Macos:   platformUsageFromAPI(w.Macos),
		Unknown: platformUsageFromAPI(w.Unknown),
	}
}

func platformUsageFromAPI(w *rdeapi.PlatformUsage) PlatformUsage {
	if w == nil {
		return PlatformUsage{}
	}
	return PlatformUsage{
		SessionCount: w.SessionCount,
		VCPU:         w.VCPU,
		MemoryGB:     w.MemoryGB,
	}
}
