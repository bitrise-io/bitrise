package rdeapi

import (
	"context"
	"testing"
)

func TestGetWorkspaceUsage_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{
		"totals": {"linux": {"sessionCount": 2, "vcpu": 4, "memoryGb": 16}},
		"users": [{
			"userId": "u-1", "userSlug": "alice-slug", "email": "alice@example.com",
			"totals": {"linux": {"sessionCount": 2, "vcpu": 4, "memoryGb": 16}}
		}],
		"unknownMachineTypeCount": 1
	}`)

	usage, err := rs.client().GetWorkspaceUsage(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("GetWorkspaceUsage: %v", err)
	}
	if want := "/v1/workspaces/ws-1/usage"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if usage.Totals == nil || usage.Totals.Linux == nil || usage.Totals.Linux.VCPU != 4 || usage.Totals.Linux.MemoryGB != 16 {
		t.Errorf("totals = %+v", usage.Totals)
	}
	// The backend omits empty bucket objects; absent buckets stay nil here
	// (the service layer densifies).
	if usage.Totals.Macos != nil {
		t.Errorf("macos bucket = %+v, want nil for an omitted bucket", usage.Totals.Macos)
	}
	if len(usage.Users) != 1 || usage.Users[0].UserSlug != "alice-slug" || usage.Users[0].UserID != "u-1" {
		t.Errorf("users = %+v", usage.Users)
	}
	if usage.UnknownMachineTypeCount != 1 {
		t.Errorf("unknownMachineTypeCount = %d, want 1", usage.UnknownMachineTypeCount)
	}
}

func TestGetWorkspaceUsage_ValidationGuard(t *testing.T) {
	rs := newRecordingServer(t, `{}`)

	if _, err := rs.client().GetWorkspaceUsage(context.Background(), ""); err == nil {
		t.Error("expected validation error, got nil")
	}
	if rs.hits != 0 {
		t.Errorf("validation guard made %d HTTP call(s); should short-circuit", rs.hits)
	}
}
