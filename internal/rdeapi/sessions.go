package rdeapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Session is the wire-format session record returned by the RDE API.
type Session struct {
	ID                          string                   `json:"id"`
	Name                        string                   `json:"name"`
	Description                 string                   `json:"description,omitempty"`
	Status                      string                   `json:"status,omitempty"`
	TemplateID                  string                   `json:"templateId,omitempty"`
	TemplateDeleted             bool                     `json:"templateDeleted,omitempty"`
	TemplateOutdated            bool                     `json:"templateOutdated,omitempty"`
	TemplateSnapshot            *SessionTemplateSnapshot `json:"templateSnapshot,omitempty"`
	AgentSessionStatus          string                   `json:"agentSessionStatus,omitempty"`
	AgentSessionStatusUpdatedAt string                   `json:"agentSessionStatusUpdatedAt,omitempty"`
	AIEnabled                   bool                     `json:"aiEnabled,omitempty"`
	AIConfigured                bool                     `json:"aiConfigured,omitempty"`
	AIPrompt                    string                   `json:"aiPrompt,omitempty"`
	AutoTerminateAt             string                   `json:"autoTerminateAt,omitempty"`
	AutoTerminateMinutes        int                      `json:"autoTerminateMinutes,omitempty"`
	SSHAddress                  string                   `json:"sshAddress,omitempty"`
	SSHPassword                 string                   `json:"sshPassword,omitempty"`
	SSHConnectionOpen           bool                     `json:"sshConnectionOpen,omitempty"`
	VNCAddress                  string                   `json:"vncAddress,omitempty"`
	VNCUsername                 string                   `json:"vncUsername,omitempty"`
	VNCPassword                 string                   `json:"vncPassword,omitempty"`
	PersistentDiskStatus        string                   `json:"persistentDiskStatus,omitempty"`
	Labels                      map[string]string        `json:"labels,omitempty"`
	// OwnerType says who owns the session: "user" (the creating user; the
	// default) or "workspace" (the workspace itself — such sessions are
	// visible to every member of the workspace). The backend may add more
	// owner kinds over time, so treat unknown values as opaque.
	OwnerType string `json:"ownerType,omitempty"`
	// OwnerID is the owner's identifier, typed by OwnerType: the owning
	// user's ID for "user", the workspace slug for "workspace".
	OwnerID   string `json:"ownerId,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// SessionTemplateSnapshot is the template config snapshotted at session
// creation time.
type SessionTemplateSnapshot struct {
	TemplateName string `json:"templateName,omitempty"`
	StackID      string `json:"stackId,omitempty"`
	// Image is the resolved image id, retained only as a read fallback for
	// snapshots taken before stackId was populated.
	Image            string          `json:"image,omitempty"`
	MachineType      string          `json:"machineType,omitempty"`
	WorkingDirectory string          `json:"workingDirectory,omitempty"`
	HasStartupScript bool            `json:"hasStartupScript,omitempty"`
	HasWarmupScript  bool            `json:"hasWarmupScript,omitempty"`
	SessionInputs    []SnapshotInput `json:"sessionInputs,omitempty"`
	FeatureFlags     []SnapshotFlag  `json:"featureFlags,omitempty"`
	WorkspaceLinks   []SnapshotLink  `json:"workspaceLinks,omitempty"`
	UpdatedAt        string          `json:"updatedAt,omitempty"`
}

// SnapshotInput is a session-input value captured at session creation.
//
// Secret values (IsSecret=true) are only returned when a request opts in
// with include_secrets=true, which this client never sets; internal/rde
// masks the value again as defense-in-depth in case the backend default
// ever changes.
type SnapshotInput struct {
	Key            string `json:"key"`
	Value          string `json:"value,omitempty"`
	IsSecret       bool   `json:"isSecret,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// SnapshotFlag is a feature-flag state captured at session creation.
type SnapshotFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled,omitempty"`
}

// SnapshotLink is a workspace link captured at session creation.
type SnapshotLink struct {
	Label      string `json:"label,omitempty"`
	FolderPath string `json:"folderPath,omitempty"`
	SortOrder  int    `json:"sortOrder,omitempty"`
}

// SessionInputValue provides a value for a session input when creating a
// session. Either Value (with optional IsSecret) OR SavedInputID is used.
type SessionInputValue struct {
	Key          string `json:"key"`
	Value        string `json:"value,omitempty"`
	IsSecret     bool   `json:"isSecret,omitempty"`
	SavedInputID string `json:"savedInputId,omitempty"`
}

// AutoMappedInput records a template session input that was auto-filled
// from the user's saved inputs when MapSavedToSessionInputs=true.
type AutoMappedInput struct {
	SessionInputKey string `json:"sessionInputKey"`
	SavedInputID    string `json:"savedInputId"`
}

// CreateSessionRequest is the POST body for creating a session. TemplateID is
// optional: omit it (and supply StackID + MachineType) to create a session
// without a template. When a template is given, StackID / MachineType
// optionally override the template's defaults for this session.
type CreateSessionRequest struct {
	Name                    string              `json:"name"`
	Description             string              `json:"description,omitempty"`
	TemplateID              string              `json:"templateId,omitempty"`
	StackID                 string              `json:"stackId,omitempty"`
	MachineType             string              `json:"machineType,omitempty"`
	SessionInputs           []SessionInputValue `json:"sessionInputs,omitempty"`
	EnabledFeatureFlagNames []string            `json:"enabledFeatureFlagNames,omitempty"`
	Cluster                 string              `json:"cluster,omitempty"`
	AIPrompt                string              `json:"aiPrompt,omitempty"`
	AutoTerminateMinutes    *int                `json:"autoTerminateMinutes,omitempty"`
	MapSavedToSessionInputs bool                `json:"mapSavedToSessionInputs,omitempty"`
	// Labels is arbitrary key=value metadata attached to the session. Sent
	// verbatim; the backend enforces the constraints (at most 32 entries;
	// keys 1-63 chars of [a-zA-Z0-9._/-] starting and ending alphanumeric;
	// values 1-255 bytes of [a-zA-Z0-9._/:+-] with no positional rules;
	// the "bitrise.io/" key prefix is reserved for system-owned labels and
	// rejected on writes).
	Labels map[string]string `json:"labels,omitempty"`
}

// UpdateSessionRequest is the PATCH body for updating a session. Pointer
// fields let the caller distinguish "unset, leave alone" from "set to
// empty/zero". Labels are merged into the session's existing labels
// (existing keys overwritten, other keys untouched) with the same
// constraints as CreateSessionRequest.Labels; RemoveLabels lists keys to
// delete (unknown keys ignored; when a key appears in both, the removal
// wins server-side).
type UpdateSessionRequest struct {
	Name                 *string           `json:"name,omitempty"`
	Description          *string           `json:"description,omitempty"`
	AutoTerminateMinutes *int              `json:"autoTerminateMinutes,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	RemoveLabels         []string          `json:"removeLabels,omitempty"`
}

// TemplateConfig is the diff-endpoint view of a template — same fields on
// both sides of the diff (snapshot vs current). Deliberately distinct from
// Template: the diff uses its own *Config-variant sub-types, and secret
// values are always stripped.
type TemplateConfig struct {
	TemplateName string `json:"templateName,omitempty"`
	StackID      string `json:"stackId,omitempty"`
	// Image is retained only as a read fallback for configs from before
	// stackId was populated.
	Image             string                   `json:"image,omitempty"`
	MachineType       string                   `json:"machineType,omitempty"`
	WorkingDirectory  string                   `json:"workingDirectory,omitempty"`
	StartupScript     string                   `json:"startupScript,omitempty"`
	WarmupScript      string                   `json:"warmupScript,omitempty"`
	SessionInputs     []TemplateConfigInput    `json:"sessionInputs,omitempty"`
	FeatureFlags      []TemplateConfigFlag     `json:"featureFlags,omitempty"`
	TemplateVariables []TemplateConfigVariable `json:"templateVariables,omitempty"`
	WorkspaceLinks    []SnapshotLink           `json:"workspaceLinks,omitempty"`
	UpdatedAt         string                   `json:"updatedAt,omitempty"`
}

// TemplateConfigInput is a session-input definition for diff purposes.
type TemplateConfigInput struct {
	Key            string `json:"key"`
	Description    string `json:"description,omitempty"`
	Required       bool   `json:"required,omitempty"`
	DefaultValue   string `json:"defaultValue,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
	IsSecret       bool   `json:"isSecret,omitempty"`
}

// TemplateConfigFlag is a feature flag with its default state.
type TemplateConfigFlag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

// TemplateConfigVariable is a template variable's metadata (no value).
type TemplateConfigVariable struct {
	Key            string `json:"key"`
	IsSecret       bool   `json:"isSecret,omitempty"`
	ExposeAsEnvVar bool   `json:"exposeAsEnvVar,omitempty"`
}

// CompareSessionTemplateResponse is the wire shape of /template-diff.
type CompareSessionTemplateResponse struct {
	Snapshot            *TemplateConfig `json:"snapshot,omitempty"`
	Current             *TemplateConfig `json:"current,omitempty"`
	ChangedVariableKeys []string        `json:"changedVariableKeys,omitempty"`
}

// CompareSessionTemplate fetches the snapshot-vs-current template diff.
// Endpoint: GET /v1/workspaces/{workspaceId}/sessions/{sessionId}/template-diff.
func (c *Client) CompareSessionTemplate(ctx context.Context, workspaceID, sessionID string) (CompareSessionTemplateResponse, error) {
	if workspaceID == "" {
		return CompareSessionTemplateResponse{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return CompareSessionTemplateResponse{}, fmt.Errorf("session ID is required")
	}
	var resp CompareSessionTemplateResponse
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/template-diff")
	if err := c.getJSON(ctx, p, &resp); err != nil {
		return CompareSessionTemplateResponse{}, err
	}
	return resp, nil
}

// ListSessions returns sessions in the workspace. Each labelSelectors entry
// is a "key=value" exact-match label filter; multiple selectors are ANDed and
// passed through verbatim as repeated query params (the backend validates
// them: key=value form, at most 8, no duplicate keys).
//
// scope selects whose sessions are listed: "mine" (sessions the caller
// created; also the backend default when empty) or "workspace" (sessions
// owned by the workspace itself, visible to every member of the workspace).
// The value is translated to the backend's enum name; anything else is
// omitted so the backend default applies — mirrors how
// ListSessionNotifications handles order.
// Endpoint: GET /v1/workspaces/{workspaceId}/sessions.
func (c *Client) ListSessions(ctx context.Context, workspaceID string, labelSelectors []string, scope string) ([]Session, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}
	p := wsPath(workspaceID, "/sessions")
	q := url.Values{}
	for _, sel := range labelSelectors {
		q.Add("labelSelectors", sel)
	}
	switch scope {
	case "mine":
		q.Set("scope", "SESSION_LIST_SCOPE_MINE")
	case "workspace":
		q.Set("scope", "SESSION_LIST_SCOPE_WORKSPACE")
	}
	if encoded := q.Encode(); encoded != "" {
		p += "?" + encoded
	}
	var resp listSessionsResp
	if err := c.getJSON(ctx, p, &resp); err != nil {
		return nil, err
	}
	return resp.Sessions, nil
}

// GetSession returns a single session by ID.
// Endpoint: GET /v1/workspaces/{workspaceId}/sessions/{sessionId}.
func (c *Client) GetSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if workspaceID == "" {
		return Session{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return Session{}, fmt.Errorf("session ID is required")
	}
	var resp sessionResp
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID))
	if err := c.getJSON(ctx, p, &resp); err != nil {
		return Session{}, err
	}
	return resp.Session, nil
}

// CreateSession creates a session from a template. Returns the new session
// and any session inputs auto-filled from saved inputs (empty unless the
// request set MapSavedToSessionInputs=true).
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions.
func (c *Client) CreateSession(ctx context.Context, workspaceID string, req CreateSessionRequest) (Session, []AutoMappedInput, error) {
	if workspaceID == "" {
		return Session{}, nil, fmt.Errorf("workspace ID is required")
	}
	var resp createSessionResp
	if err := c.sendJSON(ctx, http.MethodPost, wsPath(workspaceID, "/sessions"), req, &resp); err != nil {
		return Session{}, nil, err
	}
	return resp.Session, resp.AutoMappedInputs, nil
}

// UpdateSession patches name, description, or auto-terminate minutes.
// Endpoint: PATCH /v1/workspaces/{workspaceId}/sessions/{sessionId}.
func (c *Client) UpdateSession(ctx context.Context, workspaceID, sessionID string, req UpdateSessionRequest) (Session, error) {
	if workspaceID == "" {
		return Session{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return Session{}, fmt.Errorf("session ID is required")
	}
	var resp sessionResp
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID))
	if err := c.sendJSON(ctx, http.MethodPatch, p, req, &resp); err != nil {
		return Session{}, err
	}
	return resp.Session, nil
}

// RestoreSession restores a terminated session — the VM is re-created from
// the persistent disk and the session moves back through STARTING to
// RUNNING. The legacy /start endpoint is still served as a deprecated alias
// on the backend but /restore is the canonical name.
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions/{sessionId}/restore.
func (c *Client) RestoreSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if workspaceID == "" {
		return Session{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return Session{}, fmt.Errorf("session ID is required")
	}
	var resp sessionResp
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/restore")
	if err := c.sendJSON(ctx, http.MethodPost, p, struct{}{}, &resp); err != nil {
		return Session{}, err
	}
	return resp.Session, nil
}

// TerminateSession terminates a running session (stops the VM, preserving
// the session for later restart). The legacy /stop endpoint is still served
// as a deprecated alias on the backend but the canonical name is /terminate.
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions/{sessionId}/terminate.
func (c *Client) TerminateSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if workspaceID == "" {
		return Session{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return Session{}, fmt.Errorf("session ID is required")
	}
	var resp sessionResp
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/terminate")
	if err := c.sendJSON(ctx, http.MethodPost, p, struct{}{}, &resp); err != nil {
		return Session{}, err
	}
	return resp.Session, nil
}

// DeleteSession permanently deletes a session.
// Endpoint: DELETE /v1/workspaces/{workspaceId}/sessions/{sessionId}.
func (c *Client) DeleteSession(ctx context.Context, workspaceID, sessionID string) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	return c.del(ctx, wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)))
}

// DeleteTerminatedSessions removes every terminated (stopped) session in the
// workspace for the caller and returns the count of deleted sessions.
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions:delete-terminated.
func (c *Client) DeleteTerminatedSessions(ctx context.Context, workspaceID string) (int, error) {
	if workspaceID == "" {
		return 0, fmt.Errorf("workspace ID is required")
	}
	var resp deleteTerminatedResp
	p := wsPath(workspaceID, "/sessions:delete-terminated")
	if err := c.sendJSON(ctx, http.MethodPost, p, struct{}{}, &resp); err != nil {
		return 0, err
	}
	return resp.DeletedCount, nil
}

type listSessionsResp struct {
	Sessions []Session `json:"sessions"`
}

type sessionResp struct {
	Session Session `json:"session"`
}

type createSessionResp struct {
	Session          Session           `json:"session"`
	AutoMappedInputs []AutoMappedInput `json:"autoMappedInputs,omitempty"`
}

type deleteTerminatedResp struct {
	DeletedCount int `json:"deletedCount,omitempty"`
}
