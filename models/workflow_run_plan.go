package models

import "time"

type WorkflowRunModes struct {
	CIMode                  bool
	PRMode                  bool
	DebugMode               bool
	SecretFilteringMode     bool
	SecretEnvsFilteringMode bool
	NoOutputTimeout         time.Duration
	IsSteplibOfflineMode    bool
}

type StepExecutionPlan struct {
	UUID   string `json:"uuid"`
	StepID string `json:"step_id"`
}

type WorkflowExecutionPlan struct {
	UUID       string              `json:"uuid"`
	WorkflowID string              `json:"workflow_id"`
	Steps      []StepExecutionPlan `json:"steps"`

	IsSteplibOfflineMode bool
}

type WorkflowRunPlan struct {
	Version          string `json:"version"`
	LogFormatVersion string `json:"log_format_version"`

	CIMode                  bool `json:"ci_mode"`
	PRMode                  bool `json:"pr_mode"`
	DebugMode               bool `json:"debug_mode"`
	IsSteplibOfflineMode    bool
	NoOutputTimeoutMode     bool `json:"no_output_timeout_mode"`
	SecretFilteringMode     bool `json:"secret_filtering_mode"`
	SecretEnvsFilteringMode bool `json:"secret_envs_filtering_mode"`

	ExecutionPlan []WorkflowExecutionPlan `json:"execution_plan"`
}
