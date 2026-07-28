package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// AppBitriseYML fetches the stored bitrise.yml for an app.
// Endpoint: GET /apps/{app-slug}/bitrise.yml. The API returns plain text YAML.
func (c *Client) AppBitriseYML(ctx context.Context, appSlug string) (string, error) {
	req, err := c.newRequest(ctx, "/apps/"+appSlug+"/bitrise.yml", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")
	body, err := c.do(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// BuildBitriseYML fetches the bitrise.yml that a specific build ran with.
// Endpoint: GET /apps/{app-slug}/builds/{build-slug}/bitrise.yml.
func (c *Client) BuildBitriseYML(ctx context.Context, appSlug, buildSlug string) (string, error) {
	req, err := c.newRequest(ctx, "/apps/"+appSlug+"/builds/"+buildSlug+"/bitrise.yml", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")
	body, err := c.do(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// AppConfigUpdateRequest is the body for POST /apps/{app-slug}/bitrise.yml.
// AppConfigDatastoreYAML must be the parsed YAML content (a Go map/slice
// structure), not a raw YAML string, because the API expects a JSON object.
type AppConfigUpdateRequest struct {
	AppConfigDatastoreYAML any `json:"app_config_datastore_yaml"`
}

// UpdateAppBitriseYML uploads a new bitrise.yml for an app.
// Endpoint: POST /apps/{app-slug}/bitrise.yml.
// content must be the YAML already parsed into a Go value (e.g. map[string]any).
func (c *Client) UpdateAppBitriseYML(ctx context.Context, appSlug string, content any) error {
	_, err := c.post(ctx, "/apps/"+appSlug+"/bitrise.yml", nil, AppConfigUpdateRequest{AppConfigDatastoreYAML: content})
	return err
}

// ValidateBitriseYMLResponse is the 200 response from POST /validate-bitrise-yml.
type ValidateBitriseYMLResponse struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// ValidateBitriseYML validates a bitrise.yml string via the API.
// Endpoint: POST /validate-bitrise-yml. appSlug is optional; when non-empty
// it enables app-specific validation (stack IDs, machine types, license pools).
func (c *Client) ValidateBitriseYML(ctx context.Context, yamlContent, appSlug string) (ValidateBitriseYMLResponse, error) {
	var params url.Values
	if appSlug != "" {
		params = url.Values{"app_slug": {appSlug}}
	}

	body, err := c.post(ctx, "/validate-bitrise-yml", params, struct {
		BitriseYML string `json:"bitrise_yml"`
	}{BitriseYML: yamlContent})
	if err != nil {
		return ValidateBitriseYMLResponse{}, err
	}

	var resp ValidateBitriseYMLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ValidateBitriseYMLResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return resp, nil
}
