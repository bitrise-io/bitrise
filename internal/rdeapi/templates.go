package rdeapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Template is the wire-format template record.
type Template struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	StackID     string `json:"stackId,omitempty"`
	// Image is retained only as a read fallback for templates created before
	// stackId was populated.
	Image             string             `json:"image,omitempty"`
	MachineType       string             `json:"machineType,omitempty"`
	WorkingDirectory  string             `json:"workingDirectory,omitempty"`
	StartupScript     string             `json:"startupScript,omitempty"`
	WarmupScript      string             `json:"warmupScript,omitempty"`
	CreatedByEmail    string             `json:"createdByEmail,omitempty"`
	TemplateVariables []TemplateVariable `json:"templateVariables,omitempty"`
	SessionInputs     []SessionInputDef  `json:"sessionInputs,omitempty"`
	FeatureFlags      []FeatureFlag      `json:"featureFlags,omitempty"`
	WorkspaceLinks    []WorkspaceLink    `json:"workspaceLinks,omitempty"`
	CreatedAt         string             `json:"createdAt,omitempty"`
	UpdatedAt         string             `json:"updatedAt,omitempty"`
}

// TemplateVariable is a baked-in template variable.
//
// Secret values (IsSecret=true) are only returned by the backend when a
// request opts in with include_secrets=true. The CLI intentionally never
// sets that flag on template reads, and internal/rde masks any secret value
// again as defense-in-depth in case the backend default ever changes.
type TemplateVariable struct {
	ID             string `json:"id,omitempty"`
	Key            string `json:"key"`
	Value          string `json:"value,omitempty"`
	IsSecret       bool   `json:"isSecret,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// SessionInputDef defines an input the template prompts for at session
// creation time. (Distinct from SessionInputValue, which is the user's
// answer in a CreateSessionRequest.)
type SessionInputDef struct {
	ID             string `json:"id,omitempty"`
	Key            string `json:"key"`
	Description    string `json:"description,omitempty"`
	Required       bool   `json:"required,omitempty"`
	DefaultValue   string `json:"defaultValue,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// FeatureFlag is a toggleable feature on a template.
type FeatureFlag struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// WorkspaceLink is an IDE folder shortcut bundled with a template.
type WorkspaceLink struct {
	Label      string `json:"label,omitempty"`
	FolderPath string `json:"folderPath,omitempty"`
	SortOrder  int    `json:"sortOrder,omitempty"`
}

// TemplateVariableCreate is the create-time shape of a template variable;
// distinct from TemplateVariable because no ID is sent — the server assigns
// one.
type TemplateVariableCreate struct {
	Key            string `json:"key"`
	Value          string `json:"value,omitempty"`
	IsSecret       bool   `json:"isSecret,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// SessionInputCreate is the create-time shape of a session input definition.
type SessionInputCreate struct {
	Key            string `json:"key"`
	Description    string `json:"description,omitempty"`
	Required       bool   `json:"required,omitempty"`
	DefaultValue   string `json:"defaultValue,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// FeatureFlagCreate is the create-time shape of a feature flag.
type FeatureFlagCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// WorkspaceLinkCreate is the create-time shape of a workspace link.
type WorkspaceLinkCreate struct {
	Label           string `json:"label,omitempty"`
	FolderPath      string `json:"folderPath,omitempty"`
	FeatureFlagName string `json:"featureFlagName,omitempty"`
}

// CreateTemplateRequest is the POST body for creating a template.
type CreateTemplateRequest struct {
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	StackID           string                   `json:"stackId"`
	MachineType       string                   `json:"machineType"`
	WorkingDirectory  string                   `json:"workingDirectory,omitempty"`
	StartupScript     string                   `json:"startupScript,omitempty"`
	WarmupScript      string                   `json:"warmupScript,omitempty"`
	TemplateVariables []TemplateVariableCreate `json:"templateVariables,omitempty"`
	SessionInputs     []SessionInputCreate     `json:"sessionInputs,omitempty"`
	FeatureFlags      []FeatureFlagCreate      `json:"featureFlags,omitempty"`
	WorkspaceLinks    []WorkspaceLinkCreate    `json:"workspaceLinks,omitempty"`
}

// UpdateTemplateRequest is the PATCH body. The four UpdateXxx booleans are
// required when an array field should replace the server's existing list —
// an array with its flag unset is treated as "no change", not "clear it".
type UpdateTemplateRequest struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	StackID          *string `json:"stackId,omitempty"`
	MachineType      *string `json:"machineType,omitempty"`
	WorkingDirectory *string `json:"workingDirectory,omitempty"`
	StartupScript    *string `json:"startupScript,omitempty"`
	WarmupScript     *string `json:"warmupScript,omitempty"`

	TemplateVariables       []TemplateVariableCreate `json:"templateVariables,omitempty"`
	UpdateTemplateVariables bool                     `json:"updateTemplateVariables,omitempty"`
	SessionInputs           []SessionInputCreate     `json:"sessionInputs,omitempty"`
	UpdateSessionInputs     bool                     `json:"updateSessionInputs,omitempty"`
	FeatureFlags            []FeatureFlagCreate      `json:"featureFlags,omitempty"`
	UpdateFeatureFlags      bool                     `json:"updateFeatureFlags,omitempty"`
	WorkspaceLinks          []WorkspaceLinkCreate    `json:"workspaceLinks,omitempty"`
	UpdateWorkspaceLinks    bool                     `json:"updateWorkspaceLinks,omitempty"`
}

// ListTemplates returns every template visible in the workspace.
// Endpoint: GET /v1/workspaces/{workspaceId}/templates.
//
// Deliberately does not pass include_secrets: consumers read metadata only,
// never secret variable values.
func (c *Client) ListTemplates(ctx context.Context, workspaceID string) ([]Template, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}
	var resp listTemplatesResp
	if err := c.getJSON(ctx, wsPath(workspaceID, "/templates"), &resp); err != nil {
		return nil, err
	}
	return resp.Templates, nil
}

// GetTemplate returns a single template by ID.
// Endpoint: GET /v1/workspaces/{workspaceId}/templates/{templateId}.
//
// Deliberately does not pass include_secrets — see ListTemplates.
func (c *Client) GetTemplate(ctx context.Context, workspaceID, templateID string) (Template, error) {
	if workspaceID == "" {
		return Template{}, fmt.Errorf("workspace ID is required")
	}
	if templateID == "" {
		return Template{}, fmt.Errorf("template ID is required")
	}
	var resp templateResp
	p := wsPath(workspaceID, "/templates/"+url.PathEscape(templateID))
	if err := c.getJSON(ctx, p, &resp); err != nil {
		return Template{}, err
	}
	return resp.Template, nil
}

// CreateTemplate creates a new template in the workspace.
// Endpoint: POST /v1/workspaces/{workspaceId}/templates.
func (c *Client) CreateTemplate(ctx context.Context, workspaceID string, req CreateTemplateRequest) (Template, error) {
	if workspaceID == "" {
		return Template{}, fmt.Errorf("workspace ID is required")
	}
	var resp templateResp
	if err := c.sendJSON(ctx, http.MethodPost, wsPath(workspaceID, "/templates"), req, &resp); err != nil {
		return Template{}, err
	}
	return resp.Template, nil
}

// UpdateTemplate patches an existing template. See UpdateTemplateRequest
// for the UpdateXxx-boolean semantics around the array fields.
// Endpoint: PATCH /v1/workspaces/{workspaceId}/templates/{templateId}.
func (c *Client) UpdateTemplate(ctx context.Context, workspaceID, templateID string, req UpdateTemplateRequest) (Template, error) {
	if workspaceID == "" {
		return Template{}, fmt.Errorf("workspace ID is required")
	}
	if templateID == "" {
		return Template{}, fmt.Errorf("template ID is required")
	}
	var resp templateResp
	p := wsPath(workspaceID, "/templates/"+url.PathEscape(templateID))
	if err := c.sendJSON(ctx, http.MethodPatch, p, req, &resp); err != nil {
		return Template{}, err
	}
	return resp.Template, nil
}

// DeleteTemplate removes a template (soft-delete server-side).
// Endpoint: DELETE /v1/workspaces/{workspaceId}/templates/{templateId}.
func (c *Client) DeleteTemplate(ctx context.Context, workspaceID, templateID string) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace ID is required")
	}
	if templateID == "" {
		return fmt.Errorf("template ID is required")
	}
	return c.del(ctx, wsPath(workspaceID, "/templates/"+url.PathEscape(templateID)))
}

type listTemplatesResp struct {
	Templates []Template `json:"templates"`
}

type templateResp struct {
	Template Template `json:"template"`
}
