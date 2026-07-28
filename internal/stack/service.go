// Package stack holds the business-logic layer for stack operations.
package stack

import (
	"context"
	"sort"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// Stack is the CLI representation of a Bitrise stack. Field names are
// normalized to snake_case from the wire format's hyphenated keys.
type Stack struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	OS          string `json:"os"`
	OSVersion   int    `json:"os_version,omitempty"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	StackReport string `json:"stack_report,omitempty"`
	RemovalDate string `json:"removal_date,omitempty"`
}

// StacksResult holds all available stacks.
type StacksResult struct {
	Items []Stack `json:"items"`
}

// Service exposes stack operations to the cmd layer.
type Service struct {
	client *bitriseapi.Client
}

// NewService returns a Service backed by the given API client.
func NewService(client *bitriseapi.Client) *Service {
	return &Service{client: client}
}

// List returns available stacks sorted by ID.
// When workspaceSlug is non-empty, only stacks available to that workspace
// (including custom stacks) are returned.
func (s *Service) List(ctx context.Context, workspaceSlug string) (StacksResult, error) {
	raw, err := s.client.AvailableStacks(ctx, workspaceSlug)
	if err != nil {
		return StacksResult{}, err
	}
	items := make([]Stack, 0, len(raw))
	for id, info := range raw {
		items = append(items, fromAPI(id, info))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return StacksResult{Items: items}, nil
}

func fromAPI(id string, info bitriseapi.StackInfo) Stack {
	resolvedID := info.ID
	if resolvedID == "" {
		resolvedID = id
	}
	return Stack{
		ID:          resolvedID,
		Title:       info.Title,
		OS:          info.OS,
		OSVersion:   info.OSVersion,
		Status:      info.Status,
		Description: info.Description,
		StackReport: info.StackReport,
		RemovalDate: info.RemovalDate,
	}
}
