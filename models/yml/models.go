package yml

import (
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// # bitrise.yml documentation
//
// This is a **Markdown** doc here
// I wonder how to get syntax highlight.
type BitriseDataModel struct {
	// FormatVersion is very important but I never know what the correct value would be.
	FormatVersion        string `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string `json:"default_step_lib_source,omitempty" yaml:"default_step_lib_source,omitempty"`
	ProjectType          string `json:"project_type" yaml:"project_type"`
	//
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	//
	Services   map[string]Container   `json:"services,omitempty" yaml:"services,omitempty"`
	Containers map[string]Container   `json:"containers,omitempty" yaml:"containers,omitempty"`
	App        AppModel               `json:"app,omitempty" yaml:"app,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty" yaml:"meta,omitempty"`
	// This field is obsolete, use `triggers` within a workflow or pipeline instead.
	// See https://docs.bitrise.io/en/bitrise-ci/run-and-analyze-builds/build-triggers/configuring-build-triggers.html
	TriggerMap  TriggerMapModel            `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Pipelines   map[string]PipelineModel   `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Stages      map[string]StageModel      `json:"stages,omitempty" yaml:"stages,omitempty"`
	Workflows   map[string]WorkflowModel   `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	StepBundles map[string]StepBundleModel `json:"step_bundles,omitempty" yaml:"step_bundles,omitempty"`
	Tools       ToolsModel                 `json:"tools,omitempty" yaml:"tools,omitempty"`
	ToolConfig  *ToolConfigModel           `json:"tool_config,omitempty" yaml:"tool_config,omitempty"`
}

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
	Tools            ToolsModel                          `json:"tools,omitempty" yaml:"tools,omitempty"`
	Meta             map[string]interface{}              `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type StepBundleModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	RunIf        string                              `json:"run_if,omitempty" yaml:"run_if,omitempty"`
	Inputs       []envmanModels.EnvironmentItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps        []StepListItemStepOrBundleModel     `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type StageListItemModel map[string]StageModel

type StepListItemStepOrBundleModel map[string]any

type DockerCredentials struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Server   string `json:"server,omitempty" yaml:"server,omitempty"`
}

type StepListItemModel map[string]interface{}

type StepBundleListItemModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	RunIf        *string                             `json:"run_if,omitempty" yaml:"run_if,omitempty"`
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

type WorkflowListItemModel map[string]WorkflowModel
