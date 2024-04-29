package models

import (
	"fmt"
	"strings"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// FormatVersion ...
	FormatVersion = "14"
)

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

// PipelineModel ...
type PipelineModel struct {
	Title       string               `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Stages      []StageListItemModel `json:"stages,omitempty" yaml:"stages,omitempty"`
}

// StageListItemModel ...
type StageListItemModel map[string]StageModel

// StageModel ...
type StageModel struct {
	Title           string                       `json:"title,omitempty" yaml:"title,omitempty"`
	Summary         string                       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description     string                       `json:"description,omitempty" yaml:"description,omitempty"`
	ShouldAlwaysRun bool                         `json:"should_always_run,omitempty" yaml:"should_always_run,omitempty"`
	AbortOnFail     bool                         `json:"abort_on_fail,omitempty" yaml:"abort_on_fail,omitempty"`
	RunIf           string                       `json:"run_if,omitempty" yaml:"run_if,omitempty"`
	Workflows       []StageWorkflowListItemModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// StageWorkflowListItemModel ...
type StageWorkflowListItemModel map[string]StageWorkflowModel

// StageWorkflowModel ...
type StageWorkflowModel struct {
	RunIf string `json:"run_if,omitempty" yaml:"run_if,omitempty"`
}

// WorkflowListItemModel ...
type WorkflowListItemModel map[string]WorkflowModel

// WorkflowModel ...
type WorkflowModel struct {
	ContainerID  string                              `json:"container,omitempty" yaml:"container,omitempty"`
	ServiceIDs   []string                            `json:"services,omitempty" yaml:"services,omitempty"`
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	BeforeRun    []string                            `json:"before_run,omitempty" yaml:"before_run,omitempty"`
	AfterRun     []string                            `json:"after_run,omitempty" yaml:"after_run,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps        []StepListItemModel                 `json:"steps,omitempty" yaml:"steps,omitempty"`
	Meta         map[string]interface{}              `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type DockerCredentials struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Server   string `json:"server,omitempty" yaml:"server,omitempty"`
}

type Container struct {
	Image       string                              `json:"image,omitempty" yaml:"image,omitempty"`
	Credentials DockerCredentials                   `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Ports       []string                            `json:"ports,omitempty" yaml:"ports,omitempty"`
	Envs        []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Options     string                              `json:"options,omitempty" yaml:"options,omitempty"`
}

// AppModel ...
type AppModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
}

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion        string `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string `json:"default_step_lib_source,omitempty" yaml:"default_step_lib_source,omitempty"`
	ProjectType          string `json:"project_type" yaml:"project_type"`
	//
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	//
	Services   map[string]Container     `json:"services" yaml:"services"`
	Containers map[string]Container     `json:"containers" yaml:"containers"`
	App        AppModel                 `json:"app,omitempty" yaml:"app,omitempty"`
	Meta       map[string]interface{}   `json:"meta,omitempty" yaml:"meta,omitempty"`
	TriggerMap TriggerMapModel          `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Pipelines  map[string]PipelineModel `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Stages     map[string]StageModel    `json:"stages,omitempty" yaml:"stages,omitempty"`
	Workflows  map[string]WorkflowModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// StepIDData ...
// structured representation of a composite-step-id
//
//	a composite step id is: step-lib-source::step-id@1.0.0
type StepIDData struct {
	// SteplibSource : steplib source uri, or in case of local path just "path", and in case of direct git url just "git"
	SteplibSource string
	// IDOrURI : ID if steplib is provided, URI if local step or in case a direct git url provided
	IDorURI string
	// Version : version in the steplib, or in case of a direct git step the tag-or-branch to use
	Version string
}

// BuildRunStartModel ...
type BuildRunStartModel struct {
	EventName   string    `json:"event_name" yaml:"event_name"`
	ProjectType string    `json:"project_type" yaml:"project_type"`
	StartTime   time.Time `json:"start_time" yaml:"start_time"`
}

// BuildRunResultsModel ...
type BuildRunResultsModel struct {
	WorkflowID           string                `json:"workflow_id" yaml:"workflow_id"`
	EventName            string                `json:"event_name" yaml:"event_name"`
	ProjectType          string                `json:"project_type" yaml:"project_type"`
	StartTime            time.Time             `json:"start_time" yaml:"start_time"`
	StepmanUpdates       map[string]int        `json:"stepman_updates" yaml:"stepman_updates"`
	SuccessSteps         []StepRunResultsModel `json:"success_steps" yaml:"success_steps"`
	FailedSteps          []StepRunResultsModel `json:"failed_steps" yaml:"failed_steps"`
	FailedSkippableSteps []StepRunResultsModel `json:"failed_skippable_steps" yaml:"failed_skippable_steps"`
	SkippedSteps         []StepRunResultsModel `json:"skipped_steps" yaml:"skipped_steps"`
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	StepInfo   stepmanModels.StepInfoModel `json:"step_info" yaml:"step_info"`
	StepInputs map[string]string           `json:"step_inputs" yaml:"step_inputs"`
	Status     StepRunStatus               `json:"status" yaml:"status"`
	Idx        int                         `json:"idx" yaml:"idx"`
	RunTime    time.Duration               `json:"run_time" yaml:"run_time"`
	StartTime  time.Time                   `json:"start_time" yaml:"start_time"`
	ErrorStr   string                      `json:"error_str" yaml:"error_str"`
	ExitCode   int                         `json:"exit_code" yaml:"exit_code"`

	Timeout         time.Duration `json:"-"`
	NoOutputTimeout time.Duration `json:"-"`
}

// StepError ...
type StepError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s StepRunResultsModel) StatusReasonAndErrors() (string, []StepError) {
	switch s.Status {
	case StepRunStatusCodeSuccess:
		return "", nil
	case StepRunStatusCodeSkipped:
		return s.statusReason(), nil
	case StepRunStatusCodeSkippedWithRunIf:
		return s.statusReason(), nil
	case StepRunStatusCodeFailedSkippable:
		return s.statusReason(), s.error()
	case StepRunStatusCodeFailed:
		return "", s.error()
	case StepRunStatusCodePreparationFailed:
		return "", s.error()
	case StepRunStatusAbortedWithCustomTimeout:
		return "", s.error()
	case StepRunStatusAbortedWithNoOutputTimeout:
		return "", s.error()
	default:
		return "", nil
	}
}

func (s StepRunResultsModel) statusReason() string {
	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed,
		StepRunStatusAbortedWithCustomTimeout,
		StepRunStatusAbortedWithNoOutputTimeout:
		return ""
	case StepRunStatusCodeFailedSkippable:
		return `This Step failed, but it was marked as "is_skippable", so the build continued.`
	case StepRunStatusCodeSkipped:
		return `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`
	case StepRunStatusCodeSkippedWithRunIf:
		return fmt.Sprintf(`This Step was skipped, because its "run_if" expression evaluated to false.
The "run_if" expression was: %s`, *s.StepInfo.Step.RunIf)
	}

	return ""
}

func (s StepRunResultsModel) error() []StepError {
	message := ""

	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeSkipped,
		StepRunStatusCodeSkippedWithRunIf:
		return nil
	case StepRunStatusCodeFailedSkippable,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed:
		message = s.ErrorStr
	case StepRunStatusAbortedWithCustomTimeout:
		message = fmt.Sprintf("This Step timed out after %s.", formatStatusReasonTimeInterval(s.Timeout))
	case StepRunStatusAbortedWithNoOutputTimeout:
		message = fmt.Sprintf("This Step failed, because it has not sent any output for %s.", formatStatusReasonTimeInterval(s.NoOutputTimeout))
	}

	return []StepError{{
		Code:    s.ExitCode,
		Message: message,
	}}
}

func formatStatusReasonTimeInterval(timeInterval time.Duration) string {
	var remaining = int(timeInterval / time.Second)
	h := int(remaining / 3600)
	remaining = remaining - h*3600
	m := int(remaining / 60)
	remaining = remaining - m*60
	s := remaining

	var formattedTimeInterval = ""
	if h > 0 {
		formattedTimeInterval += fmt.Sprintf("%dh ", h)
	}

	if m > 0 {
		formattedTimeInterval += fmt.Sprintf("%dm ", m)
	}

	if s > 0 {
		formattedTimeInterval += fmt.Sprintf("%ds", s)
	}

	formattedTimeInterval = strings.TrimSpace(formattedTimeInterval)

	return formattedTimeInterval
}

// TestResultStepInfo ...
type TestResultStepInfo struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Title   string `json:"title" yaml:"title"`
	Number  int    `json:"number" yaml:"number"`
}
