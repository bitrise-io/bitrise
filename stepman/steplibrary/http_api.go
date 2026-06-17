package steplibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

// HTTPAPI fetches the V2 inventory layout (step_ids.json, latest.json,
// versions.json, step-info.json, step.json) over HTTP from a base URL.
// JSON endpoints are decoded in memory and returned as structs.
type HTTPAPI struct {
	BaseURL string
	Fetcher httpfetch.Client
}

func NewHTTPAPI(baseURL string, fetcher httpfetch.Client) *HTTPAPI {
	return &HTTPAPI{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Fetcher: fetcher,
	}
}

func (h *HTTPAPI) GetAllStepIDs(ctx context.Context) ([]string, error) {
	var out steplibindex.StepIDs
	if err := h.fetchJSON(ctx, steplibindex.StepIDsPath().URL(), &out); err != nil {
		return nil, err
	}
	return out.StepIDs, nil
}

func (h *HTTPAPI) GetLatestStepVersions(ctx context.Context, id string) (steplibindex.LatestPointer, error) {
	p, err := steplibindex.LatestPointerPath(id)
	if err != nil {
		return steplibindex.LatestPointer{}, err
	}
	var out steplibindex.LatestPointer
	if err := h.fetchJSON(ctx, p.URL(), &out); err != nil {
		return steplibindex.LatestPointer{}, err
	}
	return out, nil
}

func (h *HTTPAPI) GetAllStepVersions(ctx context.Context, id string) ([]string, error) {
	p, err := steplibindex.VersionsPath(id)
	if err != nil {
		return nil, err
	}
	var out steplibindex.Versions
	if err := h.fetchJSON(ctx, p.URL(), &out); err != nil {
		return nil, err
	}
	return out.Versions, nil
}

func (h *HTTPAPI) GetStepGroupInfo(ctx context.Context, id string) (steplibindex.StepInfo, error) {
	p, err := steplibindex.StepInfoPath(id)
	if err != nil {
		return steplibindex.StepInfo{}, err
	}
	var out steplibindex.StepInfo
	if err := h.fetchJSON(ctx, p.URL(), &out); err != nil {
		return steplibindex.StepInfo{}, err
	}
	return out, nil
}

// GetStepModel fetches the V2 step manifest (`steps/<id>/<v>/step.json`,
// which serializes models.StepModel) and returns the decoded model.
func (h *HTTPAPI) GetStepModel(ctx context.Context, step ResolvedStepVersion) (models.StepModel, error) {
	p, err := steplibindex.StepJSONPath(step.ID, step.Version)
	if err != nil {
		return models.StepModel{}, err
	}
	var out models.StepModel
	if err := h.fetchJSON(ctx, p.URL(), &out); err != nil {
		return models.StepModel{}, err
	}
	return out, nil
}

func (h *HTTPAPI) fetchJSON(ctx context.Context, path string, dst any) (err error) {
	body, err := h.Fetcher.Get(ctx, h.BaseURL+path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close response body for %s%s: %w", h.BaseURL, path, cerr)
		}
	}()
	if derr := json.NewDecoder(body).Decode(dst); derr != nil {
		return fmt.Errorf("decode %s%s: %w", h.BaseURL, path, derr)
	}
	return nil
}
