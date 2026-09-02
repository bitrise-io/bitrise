package rde

import (
	"context"
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// Session is the CLI-facing session record. JSON tags define the stable
// `--output json` shape. Field set kept minimal per the RDE plan — expand
// additively as users ask for more.
type Session struct {
	ID                          string                   `json:"id" yaml:"id"`
	Name                        string                   `json:"name" yaml:"name"`
	Description                 string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Status                      string                   `json:"status,omitempty" yaml:"status,omitempty"`
	TemplateID                  string                   `json:"template_id,omitempty" yaml:"template_id,omitempty"`
	TemplateName                string                   `json:"template_name,omitempty" yaml:"template_name,omitempty"`
	TemplateDeleted             bool                     `json:"template_deleted,omitempty" yaml:"template_deleted,omitempty"`
	TemplateOutdated            bool                     `json:"template_outdated,omitempty" yaml:"template_outdated,omitempty"`
	TemplateSnapshot            *SessionTemplateSnapshot `json:"template_snapshot,omitempty" yaml:"template_snapshot,omitempty"`
	AgentSessionStatus          string                   `json:"agent_session_status,omitempty" yaml:"agent_session_status,omitempty"`
	AgentSessionStatusUpdatedAt *time.Time               `json:"agent_session_status_updated_at,omitempty" yaml:"agent_session_status_updated_at,omitempty"`
	AIEnabled                   bool                     `json:"ai_enabled,omitempty" yaml:"ai_enabled,omitempty"`
	AIConfigured                bool                     `json:"ai_configured,omitempty" yaml:"ai_configured,omitempty"`
	AIPrompt                    string                   `json:"ai_prompt,omitempty" yaml:"ai_prompt,omitempty"`
	AutoTerminateMinutes        int                      `json:"auto_terminate_minutes,omitempty" yaml:"auto_terminate_minutes,omitempty"`
	AutoTerminateAt             *time.Time               `json:"auto_terminate_at,omitempty" yaml:"auto_terminate_at,omitempty"`
	SSHAddress                  string                   `json:"ssh_address,omitempty" yaml:"ssh_address,omitempty"`
	// SSHPassword is the ephemeral SSH password issued for this session.
	// Excluded from --output json/yml — secrets shouldn't leak into the
	// stable contract. The field is consumed internally by `rde session
	// exec` for the SSH dial.
	SSHPassword       string `json:"-" yaml:"-"`
	SSHConnectionOpen bool   `json:"ssh_connection_open,omitempty" yaml:"ssh_connection_open,omitempty"`
	VNCAddress        string `json:"vnc_address,omitempty" yaml:"vnc_address,omitempty"`
	VNCUsername       string `json:"vnc_username,omitempty" yaml:"vnc_username,omitempty"`
	// VNCPassword is the ephemeral VNC password issued for this session.
	// Same handling as SSHPassword: excluded from --output json/yml so the
	// stable contract doesn't leak secrets. Surfaced only through the
	// opt-in `rde session vnc` and `rde session open-vnc` commands.
	VNCPassword          string            `json:"-" yaml:"-"`
	PersistentDiskStatus string            `json:"persistent_disk_status,omitempty" yaml:"persistent_disk_status,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreatedAt            *time.Time        `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            *time.Time        `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// Resumable reports whether `rde claude` can resume this session: a running
// session can be reattached, and a terminated/stopped/failed one can be
// restored as long as its persistent disk is still available. Any other
// (transitional) state is reported resumable optimistically — it's an
// in-flight status that should settle into one of the above.
func (s Session) Resumable() bool {
	switch s.Status {
	case "terminated", "stopped", "failed":
		return s.PersistentDiskStatus != DiskStatusUnavailable
	default:
		return true
	}
}

// SessionTemplateSnapshot is the CLI shape of the template config captured
// at session creation. Mirrors the wire type but with snake_case tags and
// without the masked secret bag.
type SessionTemplateSnapshot struct {
	TemplateName     string          `json:"template_name,omitempty" yaml:"template_name,omitempty"`
	StackID          string          `json:"stack_id,omitempty" yaml:"stack_id,omitempty"`
	MachineType      string          `json:"machine_type,omitempty" yaml:"machine_type,omitempty"`
	WorkingDirectory string          `json:"working_directory,omitempty" yaml:"working_directory,omitempty"`
	HasStartupScript bool            `json:"has_startup_script,omitempty" yaml:"has_startup_script,omitempty"`
	HasWarmupScript  bool            `json:"has_warmup_script,omitempty" yaml:"has_warmup_script,omitempty"`
	SessionInputs    []SnapshotInput `json:"session_inputs,omitempty" yaml:"session_inputs,omitempty"`
	FeatureFlags     []SnapshotFlag  `json:"feature_flags,omitempty" yaml:"feature_flags,omitempty"`
	WorkspaceLinks   []SnapshotLink  `json:"workspace_links,omitempty" yaml:"workspace_links,omitempty"`
	UpdatedAt        *time.Time      `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type SnapshotInput struct {
	Key            string `json:"key" yaml:"key"`
	Value          string `json:"value,omitempty" yaml:"value,omitempty"`
	IsSecret       bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

type SnapshotFlag struct {
	Name    string `json:"name" yaml:"name"`
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

type SnapshotLink struct {
	Label      string `json:"label,omitempty" yaml:"label,omitempty"`
	FolderPath string `json:"folder_path,omitempty" yaml:"folder_path,omitempty"`
	SortOrder  int    `json:"sort_order,omitempty" yaml:"sort_order,omitempty"`
}

// SessionInputValue mirrors the wire type for create-session.
type SessionInputValue struct {
	Key          string
	Value        string
	IsSecret     bool
	SavedInputID string
}

// AutoMappedInput records keys auto-filled from saved inputs during create.
type AutoMappedInput struct {
	SessionInputKey string `json:"session_input_key" yaml:"session_input_key"`
	SavedInputID    string `json:"saved_input_id" yaml:"saved_input_id"`
}

// CreateSessionRequest is the CLI-side request shape. AutoTerminateMinutes
// is a pointer so "not provided" stays distinguishable from "0 = disable".
// Labels is arbitrary key=value metadata attached to the session; the
// backend validates it (entry count, key and value charset/length,
// reserved "bitrise.io/" key prefix).
type CreateSessionRequest struct {
	Name                    string
	Description             string
	TemplateID              string
	StackID                 string
	MachineType             string
	SessionInputs           []SessionInputValue
	EnabledFeatureFlagNames []string
	Cluster                 string
	AIPrompt                string
	AutoTerminateMinutes    *int
	MapSavedToSessionInputs bool
	Labels                  map[string]string
}

// UpdateSessionRequest carries optional patch fields. Pointer fields
// preserve unset semantics. Labels upserts into the session's existing
// labels; RemoveLabels deletes keys (a key in both is removed — the
// backend gives removal precedence).
type UpdateSessionRequest struct {
	Name                 *string
	Description          *string
	AutoTerminateMinutes *int
	Labels               map[string]string
	RemoveLabels         []string
}

// CreateSessionResult is what the create endpoint returns: the new session
// plus any inputs that were auto-filled from saved inputs.
type CreateSessionResult struct {
	Session          Session           `json:"session" yaml:"session"`
	AutoMappedInputs []AutoMappedInput `json:"auto_mapped_inputs,omitempty" yaml:"auto_mapped_inputs,omitempty"`
}

// ListSessions returns the caller's sessions in the workspace, optionally
// filtered by label selectors ("key=value" exact matches, ANDed). Pass nil
// for the full list.
func (s *Service) ListSessions(ctx context.Context, workspaceID string, labelSelectors []string) ([]Session, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListSessions(ctx, workspaceID, labelSelectors)
	if err != nil {
		return nil, err
	}
	out := make([]Session, 0, len(wire))
	for _, w := range wire {
		out = append(out, sessionFromAPI(w))
	}
	return out, nil
}

func (s *Service) GetSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	w, err := s.client.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		return Session{}, err
	}
	return sessionFromAPI(w), nil
}

// ResolveSessionID maps `value` to a session ID. UUID-shaped inputs
// short-circuit (no network call); names trigger a ListSessions call and an
// exact case-insensitive match. Errors clearly when zero or multiple sessions
// match the name, so callers can surface ambiguity to the user.
//
// Session names aren't unique (unlike a UUID), so an ambiguous match is an
// expected outcome — the error lists the candidate IDs so the user can re-run
// with the exact one. Mirrors ResolveTemplateID.
func (s *Service) ResolveSessionID(ctx context.Context, workspaceID, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("session is required")
	}
	if looksLikeUUID(value) {
		return value, nil
	}
	sessions, err := s.ListSessions(ctx, workspaceID, nil)
	if err != nil {
		return "", fmt.Errorf("list sessions to resolve %q: %w", value, err)
	}
	var matches []Session
	for _, sess := range sessions {
		if equalFold(sess.Name, value) {
			matches = append(matches, sess)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session named %q in workspace (try 'rde session list')", value)
	case 1:
		return matches[0].ID, nil
	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.ID)
		}
		return "", fmt.Errorf("session name %q is ambiguous (matches %d sessions: %v) — pass a session ID instead", value, len(matches), ids)
	}
}

// CreateSession creates a session. Provide either a TemplateID or, for a
// templateless session, a StackID + MachineType.
func (s *Service) CreateSession(ctx context.Context, workspaceID string, req CreateSessionRequest) (CreateSessionResult, error) {
	if s.client == nil {
		return CreateSessionResult{}, errClient()
	}
	wireInputs := make([]rdeapi.SessionInputValue, 0, len(req.SessionInputs))
	for _, i := range req.SessionInputs {
		wireInputs = append(wireInputs, rdeapi.SessionInputValue{
			Key:          i.Key,
			Value:        i.Value,
			IsSecret:     i.IsSecret,
			SavedInputID: i.SavedInputID,
		})
	}
	session, mapped, err := s.client.CreateSession(ctx, workspaceID, rdeapi.CreateSessionRequest{
		Name:                    req.Name,
		Description:             req.Description,
		TemplateID:              req.TemplateID,
		StackID:                 req.StackID,
		MachineType:             req.MachineType,
		SessionInputs:           wireInputs,
		EnabledFeatureFlagNames: req.EnabledFeatureFlagNames,
		Cluster:                 req.Cluster,
		AIPrompt:                req.AIPrompt,
		AutoTerminateMinutes:    req.AutoTerminateMinutes,
		MapSavedToSessionInputs: req.MapSavedToSessionInputs,
		Labels:                  req.Labels,
	})
	if err != nil {
		return CreateSessionResult{}, err
	}
	res := CreateSessionResult{Session: sessionFromAPI(session)}
	for _, m := range mapped {
		res.AutoMappedInputs = append(res.AutoMappedInputs, AutoMappedInput{
			SessionInputKey: m.SessionInputKey,
			SavedInputID:    m.SavedInputID,
		})
	}
	return res, nil
}

// UpdateSession patches name, description, auto-terminate minutes, or
// labels (upserts via Labels, removals via RemoveLabels).
func (s *Service) UpdateSession(ctx context.Context, workspaceID, sessionID string, req UpdateSessionRequest) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	w, err := s.client.UpdateSession(ctx, workspaceID, sessionID, rdeapi.UpdateSessionRequest{
		Name:                 req.Name,
		Description:          req.Description,
		AutoTerminateMinutes: req.AutoTerminateMinutes,
		Labels:               req.Labels,
		RemoveLabels:         req.RemoveLabels,
	})
	if err != nil {
		return Session{}, err
	}
	return sessionFromAPI(w), nil
}

// RestoreSession restores a terminated session by re-provisioning its VM
// from the persistent disk. The session re-enters the STARTING state and
// (assuming no failures) reaches RUNNING again.
func (s *Service) RestoreSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	w, err := s.client.RestoreSession(ctx, workspaceID, sessionID)
	if err != nil {
		return Session{}, err
	}
	return sessionFromAPI(w), nil
}

// TerminateSession terminates a running session (preserves the session for
// later restart; the VM goes away).
func (s *Service) TerminateSession(ctx context.Context, workspaceID, sessionID string) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	w, err := s.client.TerminateSession(ctx, workspaceID, sessionID)
	if err != nil {
		return Session{}, err
	}
	return sessionFromAPI(w), nil
}

// TemplateConfig is the CLI shape of the template config on either side of
// a session's template diff.
type TemplateConfig struct {
	TemplateName      string                   `json:"template_name,omitempty" yaml:"template_name,omitempty"`
	StackID           string                   `json:"stack_id,omitempty" yaml:"stack_id,omitempty"`
	MachineType       string                   `json:"machine_type,omitempty" yaml:"machine_type,omitempty"`
	WorkingDirectory  string                   `json:"working_directory,omitempty" yaml:"working_directory,omitempty"`
	StartupScript     string                   `json:"startup_script,omitempty" yaml:"startup_script,omitempty"`
	WarmupScript      string                   `json:"warmup_script,omitempty" yaml:"warmup_script,omitempty"`
	SessionInputs     []TemplateConfigInput    `json:"session_inputs,omitempty" yaml:"session_inputs,omitempty"`
	FeatureFlags      []TemplateConfigFlag     `json:"feature_flags,omitempty" yaml:"feature_flags,omitempty"`
	TemplateVariables []TemplateConfigVariable `json:"template_variables,omitempty" yaml:"template_variables,omitempty"`
	WorkspaceLinks    []SnapshotLink           `json:"workspace_links,omitempty" yaml:"workspace_links,omitempty"`
	UpdatedAt         *time.Time               `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// TemplateConfigInput mirrors the diff-endpoint session-input definition.
type TemplateConfigInput struct {
	Key            string `json:"key" yaml:"key"`
	Description    string `json:"description,omitempty" yaml:"description,omitempty"`
	Required       bool   `json:"required,omitempty" yaml:"required,omitempty"`
	DefaultValue   string `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
	IsSecret       bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
}

type TemplateConfigFlag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

// TemplateConfigVariable is variable metadata (values stripped server-side).
type TemplateConfigVariable struct {
	Key            string `json:"key" yaml:"key"`
	IsSecret       bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

// SessionTemplateDiff is the CLI shape of /template-diff.
type SessionTemplateDiff struct {
	Snapshot            *TemplateConfig `json:"snapshot,omitempty" yaml:"snapshot,omitempty"`
	Current             *TemplateConfig `json:"current,omitempty" yaml:"current,omitempty"`
	ChangedVariableKeys []string        `json:"changed_variable_keys,omitempty" yaml:"changed_variable_keys,omitempty"`
}

// DiffSessionTemplate returns the snapshot-vs-current template diff for
// a session. Current is nil when the template was deleted.
func (s *Service) DiffSessionTemplate(ctx context.Context, workspaceID, sessionID string) (SessionTemplateDiff, error) {
	if s.client == nil {
		return SessionTemplateDiff{}, errClient()
	}
	w, err := s.client.CompareSessionTemplate(ctx, workspaceID, sessionID)
	if err != nil {
		return SessionTemplateDiff{}, err
	}
	out := SessionTemplateDiff{ChangedVariableKeys: w.ChangedVariableKeys}
	if w.Snapshot != nil {
		c := templateConfigFromAPI(*w.Snapshot)
		out.Snapshot = &c
	}
	if w.Current != nil {
		c := templateConfigFromAPI(*w.Current)
		out.Current = &c
	}
	return out, nil
}

func templateConfigFromAPI(w rdeapi.TemplateConfig) TemplateConfig {
	out := TemplateConfig{
		TemplateName:     w.TemplateName,
		StackID:          firstNonEmpty(w.StackID, w.Image),
		MachineType:      w.MachineType,
		WorkingDirectory: w.WorkingDirectory,
		StartupScript:    w.StartupScript,
		WarmupScript:     w.WarmupScript,
		UpdatedAt:        parseTime(w.UpdatedAt),
	}
	for _, i := range w.SessionInputs {
		// Mask secret default values at the CLI boundary, same as
		// snapshotFromAPI — the backend may return them in cleartext and
		// the diff is rendered verbatim in --output json.
		def := i.DefaultValue
		if i.IsSecret {
			def = ""
		}
		out.SessionInputs = append(out.SessionInputs, TemplateConfigInput{
			Key:            i.Key,
			Description:    i.Description,
			Required:       i.Required,
			DefaultValue:   def,
			ExposeAsEnvVar: i.ExposeAsEnvVar,
			IsSecret:       i.IsSecret,
		})
	}
	for _, f := range w.FeatureFlags {
		out.FeatureFlags = append(out.FeatureFlags, TemplateConfigFlag{
			Name:        f.Name,
			Description: f.Description,
			Enabled:     f.Enabled,
		})
	}
	for _, v := range w.TemplateVariables {
		out.TemplateVariables = append(out.TemplateVariables, TemplateConfigVariable{
			Key:            v.Key,
			IsSecret:       v.IsSecret,
			ExposeAsEnvVar: v.ExposeAsEnvVar,
		})
	}
	for _, l := range w.WorkspaceLinks {
		out.WorkspaceLinks = append(out.WorkspaceLinks, SnapshotLink{
			Label:      l.Label,
			FolderPath: l.FolderPath,
			SortOrder:  l.SortOrder,
		})
	}
	return out
}

// WaitForReady polls GetSession until the session leaves the provisioning
// states ("" / "pending" / "starting" / "unknown") and returns the resulting
// Session. "unknown" means the backend can't currently determine the machine
// state; it's expected to settle, so it's treated as still provisioning.
// The caller decides whether the returned status counts as success.
// Returns context.Canceled when ctx is cancelled.
//
// onPoll, when non-nil, is called with the session's status on every poll, so a
// caller can surface the live provisioning state (e.g. in a progress spinner)
// without polling separately. It must not block.
func (s *Service) WaitForReady(ctx context.Context, workspaceID, sessionID string, interval time.Duration, onPoll func(status string)) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for {
		sess, err := s.GetSession(ctx, workspaceID, sessionID)
		if err != nil {
			return Session{}, err
		}
		if onPoll != nil {
			onPoll(sess.Status)
		}
		switch sess.Status {
		case "", "pending", "starting", "unknown":
			// still provisioning (or state not yet determinable) — keep polling
		default:
			return sess, nil
		}
		select {
		case <-ctx.Done():
			return Session{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// WaitForSSHReady polls GetSession until the session's SSH endpoint is usable
// — connection open and address + password populated — and returns the
// resulting Session. A "running" status (what WaitForReady waits for) does not
// guarantee SSH is up: the backend issues credentials a few seconds later, so
// callers that want to dial in must wait on this too.
//
// If the session leaves the "running" state while waiting (e.g. it fails or is
// terminated), it returns an error rather than spinning forever. Returns
// context.Canceled when ctx is cancelled.
func (s *Service) WaitForSSHReady(ctx context.Context, workspaceID, sessionID string, interval time.Duration) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for {
		sess, err := s.GetSession(ctx, workspaceID, sessionID)
		if err != nil {
			return Session{}, err
		}
		if sess.Status != "running" {
			return Session{}, fmt.Errorf("session is no longer running (status: %q) while waiting for SSH", sess.Status)
		}
		if sess.SSHConnectionOpen && sess.SSHAddress != "" && sess.SSHPassword != "" {
			return sess, nil
		}
		select {
		case <-ctx.Done():
			return Session{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// WaitForTerminated polls GetSession until the session leaves the
// transitional teardown states ("terminating" / "draining") and returns the
// resulting Session — normally "terminated" (or "failed"). The caller decides
// whether the final status is acceptable. Returns context.Canceled when ctx
// is cancelled.
//
// This is the teardown companion to WaitForReady: a bare TerminateSession
// returns while the session is still "terminating", so a
// 'terminate && delete' pipeline races the backend — delete rejects any
// session that isn't yet "terminated" or "failed". Waiting here closes
// that gap.
func (s *Service) WaitForTerminated(ctx context.Context, workspaceID, sessionID string, interval time.Duration) (Session, error) {
	if s.client == nil {
		return Session{}, errClient()
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for {
		sess, err := s.GetSession(ctx, workspaceID, sessionID)
		if err != nil {
			return Session{}, err
		}
		switch sess.Status {
		case "terminating", "draining":
			// still tearing down — keep polling
		default:
			return sess, nil
		}
		select {
		case <-ctx.Done():
			return Session{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// DeleteSession permanently removes a session.
func (s *Service) DeleteSession(ctx context.Context, workspaceID, sessionID string) error {
	if s.client == nil {
		return errClient()
	}
	return s.client.DeleteSession(ctx, workspaceID, sessionID)
}

// DeleteTerminatedSessions removes every terminated session and returns
// the count of sessions actually deleted.
func (s *Service) DeleteTerminatedSessions(ctx context.Context, workspaceID string) (int, error) {
	if s.client == nil {
		return 0, errClient()
	}
	return s.client.DeleteTerminatedSessions(ctx, workspaceID)
}

func sessionFromAPI(w rdeapi.Session) Session {
	out := Session{
		ID:                   w.ID,
		Name:                 w.Name,
		Description:          w.Description,
		Status:               statusFromAPI(w.Status),
		TemplateID:           w.TemplateID,
		TemplateDeleted:      w.TemplateDeleted,
		TemplateOutdated:     w.TemplateOutdated,
		AgentSessionStatus:   agentStatusFromAPI(w.AgentSessionStatus),
		AIEnabled:            w.AIEnabled,
		AIConfigured:         w.AIConfigured,
		AIPrompt:             w.AIPrompt,
		AutoTerminateMinutes: w.AutoTerminateMinutes,
		SSHAddress:           w.SSHAddress,
		SSHPassword:          w.SSHPassword,
		SSHConnectionOpen:    w.SSHConnectionOpen,
		VNCAddress:           w.VNCAddress,
		VNCUsername:          w.VNCUsername,
		VNCPassword:          w.VNCPassword,
		PersistentDiskStatus: diskStatusFromAPI(w.PersistentDiskStatus),
		Labels:               w.Labels,
	}
	out.AgentSessionStatusUpdatedAt = parseTime(w.AgentSessionStatusUpdatedAt)
	out.AutoTerminateAt = parseTime(w.AutoTerminateAt)
	out.CreatedAt = parseTime(w.CreatedAt)
	out.UpdatedAt = parseTime(w.UpdatedAt)
	if w.TemplateSnapshot != nil {
		snap := snapshotFromAPI(*w.TemplateSnapshot)
		out.TemplateSnapshot = &snap
		out.TemplateName = snap.TemplateName
	}
	return out
}

func snapshotFromAPI(w rdeapi.SessionTemplateSnapshot) SessionTemplateSnapshot {
	out := SessionTemplateSnapshot{
		TemplateName:     w.TemplateName,
		StackID:          firstNonEmpty(w.StackID, w.Image),
		MachineType:      w.MachineType,
		WorkingDirectory: w.WorkingDirectory,
		HasStartupScript: w.HasStartupScript,
		HasWarmupScript:  w.HasWarmupScript,
		UpdatedAt:        parseTime(w.UpdatedAt),
	}
	for _, i := range w.SessionInputs {
		// Mask secret values at the CLI boundary: passing them through would
		// leak the value into stdout, shell history, and log files. Keep the
		// key + is_secret marker so callers can still see what was set.
		// Because the CLI never opts into include_secrets on session reads,
		// the backend already omits these values — this masking is
		// the belt-and-suspenders second line of defense in case the backend
		// default ever changes.
		val := i.Value
		if i.IsSecret {
			val = ""
		}
		out.SessionInputs = append(out.SessionInputs, SnapshotInput{
			Key:            i.Key,
			Value:          val,
			IsSecret:       i.IsSecret,
			ExposeAsEnvVar: i.ExposeAsEnvVar,
		})
	}
	for _, f := range w.FeatureFlags {
		out.FeatureFlags = append(out.FeatureFlags, SnapshotFlag{
			Name:    f.Name,
			Enabled: f.Enabled,
		})
	}
	for _, l := range w.WorkspaceLinks {
		out.WorkspaceLinks = append(out.WorkspaceLinks, SnapshotLink{
			Label:      l.Label,
			FolderPath: l.FolderPath,
			SortOrder:  l.SortOrder,
		})
	}
	return out
}
