package rdeapi

import (
	"context"
	"testing"
)

func TestListStacks_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"stacks":[{"id":"osx-xcode-16.0.x-edge","title":"Xcode 16.0","os":"macos","osVersion":26,"status":"edge","clusterNames":["c1"]}]}`)

	stacks, err := rs.client().ListStacks(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListStacks: %v", err)
	}
	if want := "/v1/workspaces/ws-1/stacks"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(stacks) != 1 || stacks[0].ID != "osx-xcode-16.0.x-edge" || stacks[0].Title != "Xcode 16.0" {
		t.Errorf("stacks = %+v", stacks)
	}
	if stacks[0].OSVersion != 26 || len(stacks[0].ClusterNames) != 1 {
		t.Errorf("stack metadata = %+v", stacks[0])
	}
}

func TestListMachineTypes_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"machineTypes":[{"id":"m1","name":"g2.mac","title":"M2 Pro Large","cpu":"4 vCPU","ram":"6 GB","os":"macos"}]}`)

	types, err := rs.client().ListMachineTypes(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListMachineTypes: %v", err)
	}
	if want := "/v1/workspaces/ws-1/machine-types"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(types) != 1 || types[0].Name != "g2.mac" {
		t.Errorf("machine types = %+v", types)
	}
	if types[0].Title != "M2 Pro Large" || types[0].CPU != "4 vCPU" || types[0].RAM != "6 GB" || types[0].OS != "macos" {
		t.Errorf("machine-type metadata = %+v", types[0])
	}
}

func TestMachineCatalog_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	cases := map[string]func() error{
		"Stacks/no-ws":       func() error { _, err := c.ListStacks(ctx, ""); return err },
		"MachineTypes/no-ws": func() error { _, err := c.ListMachineTypes(ctx, ""); return err },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
	if rs.hits != 0 {
		t.Errorf("validation guards made %d HTTP call(s); should short-circuit", rs.hits)
	}
}
