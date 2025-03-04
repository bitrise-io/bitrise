package models

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

type GraphPipelineAlwaysRunMode string

const (
	GraphPipelineAlwaysRunModeOff      GraphPipelineAlwaysRunMode = "off"
	GraphPipelineAlwaysRunModeWorkflow GraphPipelineAlwaysRunMode = "workflow"
)

const (
	FormatVersion                   = "23"
	StepListItemWithKey             = "with"
	StepListItemStepBundleKeyPrefix = "bundle::"
)

type StepBundleModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Inputs       []envmanModels.EnvironmentItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps        []StepListItemStepOrBundleModel     `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type StepListItemStepOrBundleModel map[string]any

type StepBundleListItemModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Inputs       []envmanModels.EnvironmentItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
}

type StepListStepBundleItemModel map[string]StepBundleListItemModel

type WithModel struct {
	ContainerID string                  `json:"container,omitempty" yaml:"container,omitempty"`
	ServiceIDs  []string                `json:"services,omitempty" yaml:"services,omitempty"`
	Steps       []StepListStepItemModel `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type StepListWithItemModel map[string]WithModel

type StepListStepItemModel map[string]stepmanModels.StepModel

type StepListItemModel map[string]interface{}

type PipelineModel struct {
	Title            string                             `json:"title,omitempty" yaml:"title,omitempty"`
	Summary          string                             `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description      string                             `json:"description,omitempty" yaml:"description,omitempty"`
	Triggers         Triggers                           `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	StatusReportName string                             `json:"status_report_name,omitempty" yaml:"status_report_name,omitempty"`
	Stages           []StageListItemModel               `json:"stages,omitempty" yaml:"stages,omitempty"`
	Workflows        GraphPipelineWorkflowListItemModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	Priority         *int                               `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type StageListItemModel map[string]StageModel

type StageModel struct {
	Title           string                       `json:"title,omitempty" yaml:"title,omitempty"`
	Summary         string                       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description     string                       `json:"description,omitempty" yaml:"description,omitempty"`
	ShouldAlwaysRun bool                         `json:"should_always_run,omitempty" yaml:"should_always_run,omitempty"`
	AbortOnFail     bool                         `json:"abort_on_fail,omitempty" yaml:"abort_on_fail,omitempty"`
	RunIf           string                       `json:"run_if,omitempty" yaml:"run_if,omitempty"`
	Workflows       []StageWorkflowListItemModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

type StageWorkflowListItemModel map[string]StageWorkflowModel

type StageWorkflowModel struct {
	RunIf string `json:"run_if,omitempty" yaml:"run_if,omitempty"`
}

type GraphPipelineWorkflowListItemModel map[string]GraphPipelineWorkflowModel

type GraphPipelineWorkflowModel struct {
	DependsOn       []string                          `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	AbortOnFail     bool                              `json:"abort_on_fail,omitempty" yaml:"abort_on_fail,omitempty"`
	RunIf           GraphPipelineRunIfModel           `json:"run_if,omitempty" yaml:"run_if,omitempty"`
	ShouldAlwaysRun GraphPipelineAlwaysRunMode        `json:"should_always_run,omitempty" yaml:"should_always_run,omitempty"`
	Uses            string                            `json:"uses,omitempty" yaml:"uses,omitempty"`
	Inputs          []GraphPipelineWorkflowModelInput `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Parallel        string                            `json:"parallel,omitempty" yaml:"parallel,omitempty"`
}

type GraphPipelineWorkflowModelInput map[string]interface{}

type GraphPipelineRunIfModel struct {
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
}

func (d *GraphPipelineAlwaysRunMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}

	if err := validateGraphPipelineAlwaysRunMode(value); err != nil {
		return err
	}

	*d = GraphPipelineAlwaysRunMode(value)

	return nil
}

func (d *GraphPipelineAlwaysRunMode) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if err := validateGraphPipelineAlwaysRunMode(value); err != nil {
		return err
	}

	*d = GraphPipelineAlwaysRunMode(value)

	return nil
}

func validateGraphPipelineAlwaysRunMode(value string) error {
	allowedValues := []string{string(GraphPipelineAlwaysRunModeOff), string(GraphPipelineAlwaysRunModeWorkflow)}
	if !slices.Contains(allowedValues, value) {
		return fmt.Errorf("%s is not a valid should_always_run value (%s)", value, allowedValues)
	}

	return nil
}

type WorkflowListItemModel map[string]WorkflowModel

type WorkflowModel struct {
	Title            string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary          string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description      string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Triggers         Triggers                            `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	StatusReportName string                              `json:"status_report_name,omitempty" yaml:"status_report_name,omitempty"`
	BeforeRun        []string                            `json:"before_run,omitempty" yaml:"before_run,omitempty"`
	AfterRun         []string                            `json:"after_run,omitempty" yaml:"after_run,omitempty"`
	Environments     []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps            []StepListItemModel                 `json:"steps,omitempty" yaml:"steps,omitempty"`
	Priority         *int                                `json:"priority,omitempty" yaml:"priority,omitempty"`
	Meta             map[string]interface{}              `json:"meta,omitempty" yaml:"meta,omitempty"`
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

type AppModel struct {
	Title            string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary          string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description      string                              `json:"description,omitempty" yaml:"description,omitempty"`
	StatusReportName string                              `json:"status_report_name,omitempty" yaml:"status_report_name,omitempty"`
	Environments     []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
}

type BitriseDataModel struct {
	FormatVersion        string `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string `json:"default_step_lib_source,omitempty" yaml:"default_step_lib_source,omitempty"`
	ProjectType          string `json:"project_type" yaml:"project_type"`
	//
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	//
	Services    map[string]Container       `json:"services,omitempty" yaml:"services,omitempty"`
	Containers  map[string]Container       `json:"containers,omitempty" yaml:"containers,omitempty"`
	App         AppModel                   `json:"app,omitempty" yaml:"app,omitempty"`
	Meta        map[string]interface{}     `json:"meta,omitempty" yaml:"meta,omitempty"`
	TriggerMap  TriggerMapModel            `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Pipelines   map[string]PipelineModel   `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Stages      map[string]StageModel      `json:"stages,omitempty" yaml:"stages,omitempty"`
	Workflows   map[string]WorkflowModel   `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	StepBundles map[string]StepBundleModel `json:"step_bundles,omitempty" yaml:"step_bundles,omitempty"`
}

type BuildRunStartModel struct {
	EventName   string    `json:"event_name" yaml:"event_name"`
	ProjectType string    `json:"project_type" yaml:"project_type"`
	StartTime   time.Time `json:"start_time" yaml:"start_time"`
}

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
	remaining := int(timeInterval / time.Second)
	h := int(remaining / 3600)
	remaining = remaining - h*3600
	m := int(remaining / 60)
	remaining = remaining - m*60
	s := remaining

	formattedTimeInterval := ""
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

type TestResultStepInfo struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Title   string `json:"title" yaml:"title"`
	Number  int    `json:"number" yaml:"number"`
}
