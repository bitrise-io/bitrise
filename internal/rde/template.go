package rde

import (
	"context"
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// Template is the CLI-facing template record.
type Template struct {
	ID                string             `json:"id" yaml:"id"`
	Name              string             `json:"name" yaml:"name"`
	Description       string             `json:"description,omitempty" yaml:"description,omitempty"`
	StackID           string             `json:"stack_id,omitempty" yaml:"stack_id,omitempty"`
	MachineType       string             `json:"machine_type,omitempty" yaml:"machine_type,omitempty"`
	WorkingDirectory  string             `json:"working_directory,omitempty" yaml:"working_directory,omitempty"`
	StartupScript     string             `json:"startup_script,omitempty" yaml:"startup_script,omitempty"`
	WarmupScript      string             `json:"warmup_script,omitempty" yaml:"warmup_script,omitempty"`
	CreatedByEmail    string             `json:"created_by_email,omitempty" yaml:"created_by_email,omitempty"`
	WorkspaceID       string             `json:"workspace_id,omitempty" yaml:"workspace_id,omitempty"`
	TemplateVariables []TemplateVariable `json:"template_variables,omitempty" yaml:"template_variables,omitempty"`
	SessionInputs     []SessionInputDef  `json:"session_inputs,omitempty" yaml:"session_inputs,omitempty"`
	FeatureFlags      []FeatureFlag      `json:"feature_flags,omitempty" yaml:"feature_flags,omitempty"`
	WorkspaceLinks    []WorkspaceLink    `json:"workspace_links,omitempty" yaml:"workspace_links,omitempty"`
	CreatedAt         *time.Time         `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt         *time.Time         `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// TemplateVariable is a baked-in template variable.
type TemplateVariable struct {
	Key            string `json:"key" yaml:"key"`
	Value          string `json:"value,omitempty" yaml:"value,omitempty"`
	IsSecret       bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

type SessionInputDef struct {
	Key            string `json:"key" yaml:"key"`
	Description    string `json:"description,omitempty" yaml:"description,omitempty"`
	Required       bool   `json:"required,omitempty" yaml:"required,omitempty"`
	DefaultValue   string `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

type FeatureFlag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// WorkspaceLink is an IDE folder shortcut bundled with a template.
type WorkspaceLink struct {
	Label      string `json:"label,omitempty" yaml:"label,omitempty"`
	FolderPath string `json:"folder_path,omitempty" yaml:"folder_path,omitempty"`
	SortOrder  int    `json:"sort_order,omitempty" yaml:"sort_order,omitempty"`
}

// TemplateSpec is the editable shape of a template — used for create and
// update payloads. Pointer fields preserve "unset, leave alone" semantics
// on update; slices replace the existing list when non-nil. The JSON tags
// match the snake_case shape `template view --output json` emits, so the
// canonical workflow is view → edit → update.
type TemplateSpec struct {
	Name             *string `json:"name,omitempty" yaml:"name,omitempty"`
	Description      *string `json:"description,omitempty" yaml:"description,omitempty"`
	StackID          *string `json:"stack_id,omitempty" yaml:"stack_id,omitempty"`
	MachineType      *string `json:"machine_type,omitempty" yaml:"machine_type,omitempty"`
	WorkingDirectory *string `json:"working_directory,omitempty" yaml:"working_directory,omitempty"`
	StartupScript    *string `json:"startup_script,omitempty" yaml:"startup_script,omitempty"`
	WarmupScript     *string `json:"warmup_script,omitempty" yaml:"warmup_script,omitempty"`

	// Nil = don't touch on update; non-nil = replace the server's list with
	// these values (even if empty). On create, nil/empty is equivalent.
	TemplateVariables *[]TemplateVariableSpec `json:"template_variables,omitempty" yaml:"template_variables,omitempty"`
	SessionInputs     *[]SessionInputSpec     `json:"session_inputs,omitempty" yaml:"session_inputs,omitempty"`
	FeatureFlags      *[]FeatureFlagSpec      `json:"feature_flags,omitempty" yaml:"feature_flags,omitempty"`
	WorkspaceLinks    *[]WorkspaceLinkSpec    `json:"workspace_links,omitempty" yaml:"workspace_links,omitempty"`
}

type TemplateVariableSpec struct {
	Key            string `json:"key" yaml:"key"`
	Value          string `json:"value,omitempty" yaml:"value,omitempty"`
	IsSecret       bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

type SessionInputSpec struct {
	Key            string `json:"key" yaml:"key"`
	Description    string `json:"description,omitempty" yaml:"description,omitempty"`
	Required       bool   `json:"required,omitempty" yaml:"required,omitempty"`
	DefaultValue   string `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	ExposeAsEnvVar bool   `json:"expose_as_env_var,omitempty" yaml:"expose_as_env_var,omitempty"`
}

type FeatureFlagSpec struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type WorkspaceLinkSpec struct {
	Label           string `json:"label,omitempty" yaml:"label,omitempty"`
	FolderPath      string `json:"folder_path,omitempty" yaml:"folder_path,omitempty"`
	FeatureFlagName string `json:"feature_flag_name,omitempty" yaml:"feature_flag_name,omitempty"`
}

func (s *Service) CreateTemplate(ctx context.Context, workspaceID string, spec TemplateSpec) (Template, error) {
	if s.client == nil {
		return Template{}, errClient()
	}
	if spec.Name == nil || *spec.Name == "" {
		return Template{}, fmt.Errorf("name is required")
	}
	if spec.StackID == nil || *spec.StackID == "" {
		return Template{}, fmt.Errorf("stack_id is required")
	}
	if spec.MachineType == nil || *spec.MachineType == "" {
		return Template{}, fmt.Errorf("machine_type is required")
	}
	req := rdeapi.CreateTemplateRequest{
		Name:             *spec.Name,
		StackID:          *spec.StackID,
		MachineType:      *spec.MachineType,
		Description:      deref(spec.Description),
		WorkingDirectory: deref(spec.WorkingDirectory),
		StartupScript:    deref(spec.StartupScript),
		WarmupScript:     deref(spec.WarmupScript),
	}
	if spec.TemplateVariables != nil {
		req.TemplateVariables = toWireVariables(*spec.TemplateVariables)
	}
	if spec.SessionInputs != nil {
		req.SessionInputs = toWireSessionInputs(*spec.SessionInputs)
	}
	if spec.FeatureFlags != nil {
		req.FeatureFlags = toWireFeatureFlags(*spec.FeatureFlags)
	}
	if spec.WorkspaceLinks != nil {
		req.WorkspaceLinks = toWireWorkspaceLinks(*spec.WorkspaceLinks)
	}
	t, err := s.client.CreateTemplate(ctx, workspaceID, req)
	if err != nil {
		return Template{}, err
	}
	return templateFromAPI(t), nil
}

// UpdateTemplate patches an existing template. Pointer scalars are sent
// only when non-nil; arrays trigger their corresponding updateXxx flag
// when non-nil (even when empty — which clears the existing list).
func (s *Service) UpdateTemplate(ctx context.Context, workspaceID, templateID string, spec TemplateSpec) (Template, error) {
	if s.client == nil {
		return Template{}, errClient()
	}
	req := rdeapi.UpdateTemplateRequest{
		Name:             spec.Name,
		Description:      spec.Description,
		StackID:          spec.StackID,
		MachineType:      spec.MachineType,
		WorkingDirectory: spec.WorkingDirectory,
		StartupScript:    spec.StartupScript,
		WarmupScript:     spec.WarmupScript,
	}
	if spec.TemplateVariables != nil {
		req.TemplateVariables = toWireVariables(*spec.TemplateVariables)
		req.UpdateTemplateVariables = true
	}
	if spec.SessionInputs != nil {
		req.SessionInputs = toWireSessionInputs(*spec.SessionInputs)
		req.UpdateSessionInputs = true
	}
	if spec.FeatureFlags != nil {
		req.FeatureFlags = toWireFeatureFlags(*spec.FeatureFlags)
		req.UpdateFeatureFlags = true
	}
	if spec.WorkspaceLinks != nil {
		req.WorkspaceLinks = toWireWorkspaceLinks(*spec.WorkspaceLinks)
		req.UpdateWorkspaceLinks = true
	}
	t, err := s.client.UpdateTemplate(ctx, workspaceID, templateID, req)
	if err != nil {
		return Template{}, err
	}
	return templateFromAPI(t), nil
}

func (s *Service) DeleteTemplate(ctx context.Context, workspaceID, templateID string) error {
	if s.client == nil {
		return errClient()
	}
	return s.client.DeleteTemplate(ctx, workspaceID, templateID)
}

// ResolveTemplateID maps `value` to a template ID. UUID-shaped inputs
// short-circuit (no network call); names trigger a ListTemplates call
// and an exact case-insensitive match. Errors clearly when zero or
// multiple templates match the name, so callers can surface ambiguity
// to the user.
func (s *Service) ResolveTemplateID(ctx context.Context, workspaceID, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("template is required")
	}
	if looksLikeUUID(value) {
		return value, nil
	}
	templates, err := s.ListTemplates(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("list templates to resolve %q: %w", value, err)
	}
	var matches []Template
	for _, t := range templates {
		if equalFold(t.Name, value) {
			matches = append(matches, t)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no template named %q in workspace (try 'rde template list')", value)
	case 1:
		return matches[0].ID, nil
	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.ID)
		}
		return "", fmt.Errorf("template name %q is ambiguous (matches %d templates: %v) — pass a template ID instead", value, len(matches), ids)
	}
}

// ListTemplates returns every template visible to the caller in the workspace.
func (s *Service) ListTemplates(ctx context.Context, workspaceID string) ([]Template, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListTemplates(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]Template, 0, len(wire))
	for _, w := range wire {
		out = append(out, templateFromAPI(w))
	}
	return out, nil
}

func (s *Service) GetTemplate(ctx context.Context, workspaceID, templateID string) (Template, error) {
	if s.client == nil {
		return Template{}, errClient()
	}
	w, err := s.client.GetTemplate(ctx, workspaceID, templateID)
	if err != nil {
		return Template{}, err
	}
	return templateFromAPI(w), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func toWireVariables(in []TemplateVariableSpec) []rdeapi.TemplateVariableCreate {
	out := make([]rdeapi.TemplateVariableCreate, 0, len(in))
	for _, v := range in {
		out = append(out, rdeapi.TemplateVariableCreate{
			Key:            v.Key,
			Value:          v.Value,
			IsSecret:       v.IsSecret,
			ExposeAsEnvVar: v.ExposeAsEnvVar,
		})
	}
	return out
}

func toWireSessionInputs(in []SessionInputSpec) []rdeapi.SessionInputCreate {
	out := make([]rdeapi.SessionInputCreate, 0, len(in))
	for _, i := range in {
		out = append(out, rdeapi.SessionInputCreate{
			Key:            i.Key,
			Description:    i.Description,
			Required:       i.Required,
			DefaultValue:   i.DefaultValue,
			ExposeAsEnvVar: i.ExposeAsEnvVar,
		})
	}
	return out
}

func toWireFeatureFlags(in []FeatureFlagSpec) []rdeapi.FeatureFlagCreate {
	out := make([]rdeapi.FeatureFlagCreate, 0, len(in))
	for _, f := range in {
		out = append(out, rdeapi.FeatureFlagCreate{
			Name:        f.Name,
			Description: f.Description,
		})
	}
	return out
}

func toWireWorkspaceLinks(in []WorkspaceLinkSpec) []rdeapi.WorkspaceLinkCreate {
	out := make([]rdeapi.WorkspaceLinkCreate, 0, len(in))
	for _, l := range in {
		out = append(out, rdeapi.WorkspaceLinkCreate{
			Label:           l.Label,
			FolderPath:      l.FolderPath,
			FeatureFlagName: l.FeatureFlagName,
		})
	}
	return out
}

func templateFromAPI(w rdeapi.Template) Template {
	out := Template{
		ID:               w.ID,
		Name:             w.Name,
		Description:      w.Description,
		StackID:          firstNonEmpty(w.StackID, w.Image),
		MachineType:      w.MachineType,
		WorkingDirectory: w.WorkingDirectory,
		StartupScript:    w.StartupScript,
		WarmupScript:     w.WarmupScript,
		CreatedByEmail:   w.CreatedByEmail,
		WorkspaceID:      w.WorkspaceID,
		CreatedAt:        parseTime(w.CreatedAt),
		UpdatedAt:        parseTime(w.UpdatedAt),
	}
	for _, v := range w.TemplateVariables {
		// Mask secret values at the CLI boundary: passing them through
		// would leak the value into stdout, shell history, and log files.
		// Because the CLI never opts into include_secrets on template reads,
		// the backend already omits these values — this masking
		// is the belt-and-suspenders second line of defense.
		val := v.Value
		if v.IsSecret {
			val = ""
		}
		out.TemplateVariables = append(out.TemplateVariables, TemplateVariable{
			Key:            v.Key,
			Value:          val,
			IsSecret:       v.IsSecret,
			ExposeAsEnvVar: v.ExposeAsEnvVar,
		})
	}
	for _, i := range w.SessionInputs {
		out.SessionInputs = append(out.SessionInputs, SessionInputDef{
			Key:            i.Key,
			Description:    i.Description,
			Required:       i.Required,
			DefaultValue:   i.DefaultValue,
			ExposeAsEnvVar: i.ExposeAsEnvVar,
		})
	}
	for _, f := range w.FeatureFlags {
		out.FeatureFlags = append(out.FeatureFlags, FeatureFlag{
			Name:        f.Name,
			Description: f.Description,
		})
	}
	for _, l := range w.WorkspaceLinks {
		out.WorkspaceLinks = append(out.WorkspaceLinks, WorkspaceLink{
			Label:      l.Label,
			FolderPath: l.FolderPath,
			SortOrder:  l.SortOrder,
		})
	}
	return out
}
