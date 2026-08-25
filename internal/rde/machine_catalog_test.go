package rde

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

func TestMachineTypesForStack_FiltersByClusterOverlap(t *testing.T) {
	svc := catalogServer(t)

	// "linux" is provisionable in clusters a and b, so only machine types in a
	// or b are compatible — "m2" (cluster c) must be excluded.
	got, err := svc.MachineTypesForStack(context.Background(), "ws-1", "linux")
	if err != nil {
		t.Fatalf("MachineTypesForStack: %v", err)
	}
	gotNames := map[string]bool{}
	for _, mt := range got {
		gotNames[mt.Name] = true
	}
	if !gotNames["small"] || !gotNames["big"] {
		t.Errorf("expected small+big for linux, got %+v", got)
	}
	if gotNames["m2"] {
		t.Errorf("m2 (cluster c) should not be compatible with linux, got %+v", got)
	}
}

func TestMachineTypesForStack_SingleCluster(t *testing.T) {
	svc := catalogServer(t)

	// "mac" is only in cluster c, so only "m2" is compatible.
	got, err := svc.MachineTypesForStack(context.Background(), "ws-1", "mac")
	if err != nil {
		t.Fatalf("MachineTypesForStack: %v", err)
	}
	if len(got) != 1 || got[0].Name != "m2" {
		t.Errorf("expected only m2 for mac, got %+v", got)
	}
	// The backend's friendly metadata flows through the mapper.
	if got[0].Title != "M2 Pro" || got[0].CPU != "4 vCPU" || got[0].RAM != "6 GB" || got[0].OS != "macos" {
		t.Errorf("machine-type metadata not mapped: %+v", got[0])
	}
}

func TestMachineTypesForStack_UnknownStackErrors(t *testing.T) {
	svc := catalogServer(t)

	_, err := svc.MachineTypesForStack(context.Background(), "ws-1", "windows")
	if err == nil || !strings.Contains(err.Error(), "not found in this workspace") {
		t.Errorf("err = %v, want not-found error", err)
	}
}

func TestListStacks_CarriesIsDefault(t *testing.T) {
	svc := catalogServer(t)

	stacks, err := svc.ListStacks(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListStacks: %v", err)
	}
	var defaults []string
	for _, st := range stacks {
		if st.IsDefault {
			defaults = append(defaults, st.ID)
		}
	}
	if len(defaults) != 1 || defaults[0] != "mac" {
		t.Errorf("expected only mac flagged default, got %v", defaults)
	}
}

// catalogServer serves a fixed stacks / machine-types catalog, switching on the
// request path. The catalog mirrors the real shape: a stack lists the clusters
// where it can be provisioned, and machine types are compatible with a stack
// when they share a cluster.
func catalogServer(t *testing.T) *Service {
	t.Helper()
	const stacks = `{"stacks":[
		{"id":"linux","title":"Ubuntu 24.04","os":"linux","clusterNames":["a","b"]},
		{"id":"mac","title":"Xcode 16.0","os":"macos","clusterNames":["c"],"isDefault":true}
	]}`
	const machineTypes = `{"machineTypes":[
		{"id":"small-a","name":"small","clusterName":"a"},
		{"id":"big-b","name":"big","clusterName":"b"},
		{"id":"m2-c","name":"m2","clusterName":"c","title":"M2 Pro","cpu":"4 vCPU","ram":"6 GB","os":"macos"}
	]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/machine-types"):
			_, _ = w.Write([]byte(machineTypes))
		case strings.HasSuffix(r.URL.Path, "/stacks"):
			_, _ = w.Write([]byte(stacks))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	client, err := rdeapi.New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return NewService(client)
}
