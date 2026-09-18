package rde

import (
	"context"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// SavedInput is the CLI-facing saved-input record. `Value` is masked by
// savedInputFromAPI when IsSecret=true — the backend omits secret values on
// reads unless include_secrets=true is requested, which the CLI
// never does, and the CLI blanks any value the backend does return before any
// renderer sees it.
type SavedInput struct {
	ID        string     `json:"id" yaml:"id"`
	Key       string     `json:"key" yaml:"key"`
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	IsSecret  bool       `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type CreateSavedInputRequest struct {
	Key      string
	Value    string
	IsSecret bool
}

// UpdateSavedInputRequest's pointer fields preserve "unset, leave alone"
// semantics.
type UpdateSavedInputRequest struct {
	Value    *string
	IsSecret *bool
}

// ListSavedInputs returns every saved input for the caller. Saved inputs
// are user-scoped, not workspace-scoped — no workspace ID needed.
func (s *Service) ListSavedInputs(ctx context.Context) ([]SavedInput, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListSavedInputs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]SavedInput, 0, len(wire))
	for _, w := range wire {
		out = append(out, savedInputFromAPI(w))
	}
	return out, nil
}

func (s *Service) GetSavedInput(ctx context.Context, id string) (SavedInput, error) {
	if s.client == nil {
		return SavedInput{}, errClient()
	}
	w, err := s.client.GetSavedInput(ctx, id)
	if err != nil {
		return SavedInput{}, err
	}
	return savedInputFromAPI(w), nil
}

func (s *Service) CreateSavedInput(ctx context.Context, req CreateSavedInputRequest) (SavedInput, error) {
	if s.client == nil {
		return SavedInput{}, errClient()
	}
	w, err := s.client.CreateSavedInput(ctx, rdeapi.CreateSavedInputRequest{
		Key:      req.Key,
		Value:    req.Value,
		IsSecret: req.IsSecret,
	})
	if err != nil {
		return SavedInput{}, err
	}
	return savedInputFromAPI(w), nil
}

func (s *Service) UpdateSavedInput(ctx context.Context, id string, req UpdateSavedInputRequest) (SavedInput, error) {
	if s.client == nil {
		return SavedInput{}, errClient()
	}
	w, err := s.client.UpdateSavedInput(ctx, id, rdeapi.UpdateSavedInputRequest{
		Value:    req.Value,
		IsSecret: req.IsSecret,
	})
	if err != nil {
		return SavedInput{}, err
	}
	return savedInputFromAPI(w), nil
}

func (s *Service) DeleteSavedInput(ctx context.Context, id string) error {
	if s.client == nil {
		return errClient()
	}
	return s.client.DeleteSavedInput(ctx, id)
}

func savedInputFromAPI(w rdeapi.SavedInput) SavedInput {
	// Mask secret values at the CLI boundary: passing them through would
	// leak the value into --output json, shell history, and log files. Keep
	// the key + is_secret marker so callers can still see what was set.
	// Because the CLI never opts into include_secrets on the read endpoints,
	// the backend already omits these values on List/Get — this
	// masking is the belt-and-suspenders second line of defense (and also
	// covers the value create/update may echo back). Mirrors templateFromAPI
	// / snapshotFromAPI.
	val := w.Value
	if w.IsSecret {
		val = ""
	}
	return SavedInput{
		ID:        w.ID,
		Key:       w.Key,
		Value:     val,
		IsSecret:  w.IsSecret,
		CreatedAt: parseTime(w.CreatedAt),
		UpdatedAt: parseTime(w.UpdatedAt),
	}
}
