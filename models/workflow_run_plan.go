package models

import (
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

type WorkflowRunModes struct {
	CIMode                  bool
	PRMode                  bool
	DebugMode               bool
	SecretFilteringMode     bool
	SecretEnvsFilteringMode bool
	NoOutputTimeout         time.Duration
	IsSteplibOfflineMode    bool
}

// TODO: separate Plans from JSON event logging and actual workflow execution

type StepExecutionPlan struct {
	UUID   string `json:"uuid"`
	StepID string `json:"step_id"`

	Step stepmanModels.StepModel `json:"-"`
	// With (container) group
	WithGroupUUID string   `json:"with_group_uuid,omitempty"`
	ContainerID   string   `json:"-"`
	ServiceIDs    []string `json:"-"`
	// Step Bundle group
	StepBundleUUID string                              `json:"step_bundle_uuid,omitempty"`
	StepBundleEnvs []envmanModels.EnvironmentItemModel `json:"-"`
}

type WorkflowExecutionPlan struct {
	UUID                 string              `json:"uuid"`
	WorkflowID           string              `json:"workflow_id"`
	Steps                []StepExecutionPlan `json:"steps"`
	WorkflowTitle        string              `json:"-"`
	IsSteplibOfflineMode bool                `json:"-"`
}

type ContainerPlan struct {
	Image string `json:"image"`
}

type WithGroupPlan struct {
	Services  []ContainerPlan `json:"services,omitempty"`
	Container ContainerPlan   `json:"container,omitempty"`
}

type StepBundlePlan struct {
	ID string `json:"id"`
}

type WorkflowRunPlan struct {
	Version          string `json:"version"`
	LogFormatVersion string `json:"log_format_version"`

	CIMode                  bool `json:"ci_mode"`
	PRMode                  bool `json:"pr_mode"`
	DebugMode               bool `json:"debug_mode"`
	IsSteplibOfflineMode    bool `json:"-"`
	NoOutputTimeoutMode     bool `json:"no_output_timeout_mode"`
	SecretFilteringMode     bool `json:"secret_filtering_mode"`
	SecretEnvsFilteringMode bool `json:"secret_envs_filtering_mode"`

	WithGroupPlans  map[string]WithGroupPlan  `json:"with_groups,omitempty"`
	StepBundlePlans map[string]StepBundlePlan `json:"step_bundles,omitempty"`
	ExecutionPlan   []WorkflowExecutionPlan   `json:"execution_plan"`
}
