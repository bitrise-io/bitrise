package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type StepResponse struct {
	ID                  string                    `json:"id"`
	StepRef             string                    `json:"step_ref"`
	Title               string                    `json:"title"`
	Summary             string                    `json:"summary,omitempty"`
	Description         string                    `json:"description,omitempty"`
	Version             string                    `json:"version,omitempty"`
	LatestVersionNumber string                    `json:"latest_version_number,omitempty"`
	Maintainer          string                    `json:"maintainer,omitempty"`
	IsDeprecated        bool                      `json:"is_deprecated,omitempty"`
	IsLatest            bool                      `json:"is_latest,omitempty"`
	Inputs              []StepInputOutputResponse `json:"inputs,omitempty"`
	Outputs             []StepInputOutputResponse `json:"outputs,omitempty"`
}

type StepInputOutputResponse struct {
	Name         string   `json:"name,omitempty"`
	Title        string   `json:"title,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	Description  string   `json:"description,omitempty"`
	DefaultValue string   `json:"default_value,omitempty"`
	IsRequired   bool     `json:"is_required,omitempty"`
	IsSensitive  bool     `json:"is_sensitive,omitempty"`
	ValueOptions []string `json:"value_options,omitempty"`
}

type StepSearchOptions struct {
	Query       string
	Categories  []string
	Maintainers []string
}

func (o StepSearchOptions) params() url.Values {
	p := url.Values{}
	p.Set("query", o.Query)
	for _, c := range o.Categories {
		p.Add("categories", c)
	}
	for _, m := range o.Maintainers {
		p.Add("maintainers", m)
	}
	return p
}

// SearchSteps returns only the latest, non-deprecated version of each
// matching step.
func (c *Client) SearchSteps(ctx context.Context, opts StepSearchOptions) ([]StepResponse, error) {
	req, err := c.newRequest(ctx, "/search-steps", opts.params())
	if err != nil {
		return nil, err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var result []StepResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode steps response: %w", err)
	}
	return result, nil
}

// StepInputs' stepRef must be formatted as step_id@version (e.g. git-clone@8.3.1).
func (c *Client) StepInputs(ctx context.Context, stepRef string) ([]StepInputOutputResponse, error) {
	req, err := c.newRequest(ctx, "/step-inputs", url.Values{"step_ref": {stepRef}})
	if err != nil {
		return nil, err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var result []StepInputOutputResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode step inputs response: %w", err)
	}
	return result, nil
}
