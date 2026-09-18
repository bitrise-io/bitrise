package rde

import (
	"context"
	"fmt"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// Stack is a machine stack available in the workspace. The ID is the stable
// contract stored on a template/session; the rest is human-friendly metadata.
type Stack struct {
	ID           string `json:"id" yaml:"id"`
	Title        string `json:"title,omitempty" yaml:"title,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	OS           string `json:"os,omitempty" yaml:"os,omitempty"`
	OSVersion    int32  `json:"os_version,omitempty" yaml:"os_version,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
	XcodeVersion string `json:"xcode_version,omitempty" yaml:"xcode_version,omitempty"`
	// IsDefault is set by the backend on the deployment's default stack.
	IsDefault bool `json:"is_default,omitempty" yaml:"is_default,omitempty"`
	// ClusterNames are the clusters where this stack can be provisioned.
	ClusterNames []string `json:"cluster_names,omitempty" yaml:"cluster_names,omitempty"`
	// DescriptionLink points at the stack's pre-installed tools / system report.
	DescriptionLink string `json:"description_link,omitempty" yaml:"description_link,omitempty"`
}

// MachineType is a machine size available in the workspace. Name is the
// contract (what templates/sessions store); Title/CPU/RAM are human-friendly
// display metadata and may be empty when the backend has none.
type MachineType struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	ClusterName string `json:"cluster_name,omitempty" yaml:"cluster_name,omitempty"`
	// IsDefault is set by the backend on the deployment's default machine type.
	IsDefault bool   `json:"is_default,omitempty" yaml:"is_default,omitempty"`
	Title     string `json:"title,omitempty" yaml:"title,omitempty"`
	CPU       string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	RAM       string `json:"ram,omitempty" yaml:"ram,omitempty"`
	OS        string `json:"os,omitempty" yaml:"os,omitempty"`
}

func (s *Service) ListStacks(ctx context.Context, workspaceID string) ([]Stack, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListStacks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]Stack, 0, len(wire))
	for _, w := range wire {
		out = append(out, Stack{
			ID:              w.ID,
			Title:           w.Title,
			Description:     w.Description,
			OS:              w.OS,
			OSVersion:       w.OSVersion,
			Status:          w.Status,
			XcodeVersion:    w.XcodeVersion,
			IsDefault:       w.IsDefault,
			ClusterNames:    w.ClusterNames,
			DescriptionLink: w.DescriptionLink,
		})
	}
	return out, nil
}

func (s *Service) ListMachineTypes(ctx context.Context, workspaceID string) ([]MachineType, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListMachineTypes(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]MachineType, 0, len(wire))
	for _, w := range wire {
		out = append(out, machineTypeFromAPI(w))
	}
	return out, nil
}

func machineTypeFromAPI(w rdeapi.MachineType) MachineType {
	return MachineType{
		ID:          w.ID,
		Name:        w.Name,
		ClusterName: w.ClusterName,
		IsDefault:   w.IsDefault,
		Title:       w.Title,
		CPU:         w.CPU,
		RAM:         w.RAM,
		OS:          w.OS,
	}
}

// MachineTypesForStack returns the machine types whose cluster overlaps with
// the clusters offering the stack with the given ID. A stack is provisionable
// in one or more clusters; a machine type is compatible when it's offered by at
// least one of those same clusters. Mirrors the FE's client-side join.
//
// It errors if stackID isn't available in the workspace.
func (s *Service) MachineTypesForStack(ctx context.Context, workspaceID, stackID string) ([]MachineType, error) {
	stacks, err := s.ListStacks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	stackClusters := make(map[string]struct{})
	found := false
	for _, st := range stacks {
		if st.ID == stackID {
			found = true
			for _, c := range st.ClusterNames {
				stackClusters[c] = struct{}{}
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("stack %q not found in this workspace", stackID)
	}
	machineTypes, err := s.ListMachineTypes(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]MachineType, 0, len(machineTypes))
	for _, mt := range machineTypes {
		if _, ok := stackClusters[mt.ClusterName]; ok {
			out = append(out, mt)
		}
	}
	return out, nil
}
