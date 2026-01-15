package models

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/version"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const LogFormatVersion = "2"

// TODO: WorkflowRunPlan is used coordinate Steps' execution (cli/run.go) and also to log JSON events (PrintBitriseStartedEvent in log/log.go).
//  Some fields are only relevant for one of these purposes, that's why they have `json:"-"` struct tags.
//  Consider splitting this struct into two separate structs to separate these concerns.

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

	ExecutionPlan   []WorkflowExecutionPlan   `json:"execution_plan"`
	WithGroupPlans  map[string]WithGroupPlan  `json:"with_groups,omitempty"`
	StepBundlePlans map[string]StepBundlePlan `json:"step_bundles,omitempty"`
}

type WorkflowExecutionPlan struct {
	UUID                 string              `json:"uuid"`
	WorkflowID           string              `json:"workflow_id"`
	Steps                []StepExecutionPlan `json:"steps"`
	WorkflowTitle        string              `json:"-"`
	IsSteplibOfflineMode bool                `json:"-"`
}

type WithGroupPlan struct {
	Services  []ContainerPlan `json:"services,omitempty"`
	Container ContainerPlan   `json:"container,omitempty"`
}

type StepBundlePlan struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

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

	// StepBundleRunIfs stores each run_if statements of the including Step Bundles.
	// The first element is the run_if statement of the top most Step Bundle including the given Step.
	// To execute the Step, all run_if statements must be evaluated to true.
	StepBundleRunIfs []string `json:"-"`
}

type ContainerPlan struct {
	Image string `json:"image"`
}

type WorkflowRunModes struct {
	CIMode                  bool
	PRMode                  bool
	DebugMode               bool
	SecretFilteringMode     bool
	SecretEnvsFilteringMode bool
	NoOutputTimeout         time.Duration
	IsSteplibOfflineMode    bool
}

type WorkflowRunPlanBuilder struct {
	workflows    map[string]WorkflowModel
	stepBundles  map[string]StepBundleModel
	containers   map[string]Container
	services     map[string]Container
	uuidProvider func() string

	withGroupPlans  map[string]WithGroupPlan
	stepBundlePlans map[string]StepBundlePlan
}

func NewWorkflowRunPlanBuilder(workflows map[string]WorkflowModel, stepBundles map[string]StepBundleModel, containers map[string]Container, services map[string]Container, uuidProvider func() string) *WorkflowRunPlanBuilder {
	return &WorkflowRunPlanBuilder{
		workflows:       workflows,
		stepBundles:     stepBundles,
		containers:      containers,
		services:        services,
		uuidProvider:    uuidProvider,
		withGroupPlans:  map[string]WithGroupPlan{},
		stepBundlePlans: map[string]StepBundlePlan{},
	}
}

func (builder *WorkflowRunPlanBuilder) Build(modes WorkflowRunModes, workflowID string) (WorkflowRunPlan, error) {
	var executionPlan []WorkflowExecutionPlan

	workflowList := builder.walkWorkflows(workflowID, builder.workflows, nil)
	for _, workflowID := range workflowList {
		workflow := builder.workflows[workflowID]

		stepPlans, err := builder.processStepList(workflowID)
		if err != nil {
			return WorkflowRunPlan{}, err
		}

		workflowTitle := workflow.Title
		if workflowTitle == "" {
			workflowTitle = workflowID
		}

		executionPlan = append(executionPlan, WorkflowExecutionPlan{
			UUID:                 builder.uuidProvider(),
			WorkflowID:           workflowID,
			Steps:                stepPlans,
			WorkflowTitle:        workflowTitle,
			IsSteplibOfflineMode: modes.IsSteplibOfflineMode,
		})
	}

	cliVersion := version.VERSION
	if version.IsAlternativeInstallation {
		cliVersion = fmt.Sprintf("%s (%s)", cliVersion, version.Commit)
	}

	return WorkflowRunPlan{
		Version:                 cliVersion,
		LogFormatVersion:        LogFormatVersion,
		CIMode:                  modes.CIMode,
		PRMode:                  modes.PRMode,
		DebugMode:               modes.DebugMode,
		IsSteplibOfflineMode:    modes.IsSteplibOfflineMode,
		NoOutputTimeoutMode:     modes.NoOutputTimeout > 0,
		SecretFilteringMode:     modes.SecretFilteringMode,
		SecretEnvsFilteringMode: modes.SecretEnvsFilteringMode,
		WithGroupPlans:          builder.withGroupPlans,
		StepBundlePlans:         builder.stepBundlePlans,
		ExecutionPlan:           executionPlan,
	}, nil
}

func (builder *WorkflowRunPlanBuilder) walkWorkflows(workflowID string, workflows map[string]WorkflowModel, workflowStack []string) []string {
	workflow := workflows[workflowID]
	for _, before := range workflow.BeforeRun {
		workflowStack = builder.walkWorkflows(before, workflows, workflowStack)
	}

	workflowStack = append(workflowStack, workflowID)

	for _, after := range workflow.AfterRun {
		workflowStack = builder.walkWorkflows(after, workflows, workflowStack)
	}

	return workflowStack
}

func (builder *WorkflowRunPlanBuilder) processStepList(workflowID string) ([]StepExecutionPlan, error) {

	var stepPlans []StepExecutionPlan

	workflow := builder.workflows[workflowID]
	for _, stepListItem := range workflow.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return nil, err
		}

		if t == StepListItemTypeStep {
			plan, err := builder.processStep(key, &stepListItem)
			if err != nil {
				return nil, err
			}

			stepPlans = append(stepPlans, *plan)
		} else if t == StepListItemTypeWith {
			plans, err := builder.processWithGroup(&stepListItem)
			if err != nil {
				return nil, err
			}

			stepPlans = append(stepPlans, plans...)
		} else if t == StepListItemTypeBundle {
			plans, err := builder.processStepBundle(key, &stepListItem)
			if err != nil {
				return nil, err
			}

			stepPlans = append(stepPlans, plans...)
		}
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) processStep(stepID string, stepListItem StepListItem) (*StepExecutionPlan, error) {
	_, step, err := stepListItem.GetStep()
	if err != nil {
		return nil, err
	}

	return &StepExecutionPlan{
		UUID:   builder.uuidProvider(),
		StepID: stepID,
		Step:   *step,
	}, nil
}

func (builder *WorkflowRunPlanBuilder) processWithGroup(stepListItem StepListItem) ([]StepExecutionPlan, error) {
	with, err := stepListItem.GetWith()
	if err != nil {
		return nil, err
	}

	groupID := builder.uuidProvider()

	var containerPlan ContainerPlan
	if with.ContainerID != "" {
		containerPlan.Image = builder.containers[with.ContainerID].Image
	}

	var servicePlans []ContainerPlan
	for _, serviceID := range with.ServiceIDs {
		servicePlans = append(servicePlans, ContainerPlan{Image: builder.services[serviceID].Image})
	}

	builder.withGroupPlans[groupID] = WithGroupPlan{
		Services:  servicePlans,
		Container: containerPlan,
	}

	var stepPlans []StepExecutionPlan
	for _, stepListStepItem := range with.Steps {
		stepID, step, err := stepListStepItem.GetStep()
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, StepExecutionPlan{
			UUID:          builder.uuidProvider(),
			StepID:        stepID,
			Step:          *step,
			WithGroupUUID: groupID,
			ContainerID:   with.ContainerID,
			ServiceIDs:    with.ServiceIDs,
		})
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) processStepBundle(bundleID string, stepListItem StepListItem) ([]StepExecutionPlan, error) {
	bundleOverride, err := stepListItem.GetBundle()
	if err != nil {
		return nil, err
	}

	bundleDefinition, ok := builder.stepBundles[bundleID]
	if !ok {
		return nil, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
	}

	bundleEnvs, err := builder.gatherBundleEnvs(*bundleOverride, bundleDefinition)
	if err != nil {
		return nil, err
	}

	bundleUUID := builder.uuidProvider()
	title := bundleDefinition.Title
	if bundleOverride.Title != "" {
		title = bundleOverride.Title
	}

	builder.stepBundlePlans[bundleUUID] = StepBundlePlan{
		ID:    bundleID,
		Title: title,
	}

	runIf := bundleDefinition.RunIf
	if bundleOverride.RunIf != nil {
		runIf = *bundleOverride.RunIf
	}
	var runIfs []string
	if runIf != "" {
		runIfs = []string{runIf}
	}

	plans, err := builder.gatherBundleSteps(bundleDefinition, bundleUUID, bundleEnvs, runIfs)
	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (builder *WorkflowRunPlanBuilder) gatherBundleSteps(bundleDefinition StepBundleModel, bundleUUID string, bundleEnvs []envmanModels.EnvironmentItemModel, runIfs []string) ([]StepExecutionPlan, error) {
	var stepPlans []StepExecutionPlan
	for _, stepListStepOrBundleItem := range bundleDefinition.Steps {
		key, t, err := stepListStepOrBundleItem.GetKeyAndType()
		if err != nil {
			return nil, err
		}

		if t == StepListItemTypeStep {
			_, step, err := stepListStepOrBundleItem.GetStep()
			if err != nil {
				return nil, err
			}

			stepID := key
			stepPlan := StepExecutionPlan{
				UUID:             builder.uuidProvider(),
				StepID:           stepID,
				Step:             *step,
				StepBundleUUID:   bundleUUID,
				StepBundleRunIfs: runIfs,
				StepBundleEnvs:   bundleEnvs,
			}

			stepPlans = append(stepPlans, stepPlan)
		} else if t == StepListItemTypeBundle {
			bundleID := key
			override, err := stepListStepOrBundleItem.GetBundle()
			if err != nil {
				return nil, err
			}

			definition, ok := builder.stepBundles[bundleID]
			if !ok {
				return nil, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
			}

			envs, err := builder.gatherBundleEnvs(*override, definition)
			if err != nil {
				return nil, err
			}
			envs = append(bundleEnvs, envs...)

			uuid := builder.uuidProvider()
			title := definition.Title
			if override.Title != "" {
				title = override.Title
			}

			builder.stepBundlePlans[uuid] = StepBundlePlan{
				ID:    bundleID,
				Title: title,
			}

			runIf := definition.RunIf
			if override.RunIf != nil {
				runIf = *override.RunIf
			}

			// Create a new runIfs slice that includes the runIf of the current bundle, instead of modifying the original slice.
			// This is necessary to ensure that the runIfs of the current bundle are evaluated correctly in the context of the parent bundle.
			// The Go slice wraps a pointer to the actual data inside.
			// So passing it around and adding items to it would update all the slices internal data storage.
			var newBundleRunIfs []string
			if len(runIfs) > 0 {
				newBundleRunIfs = make([]string, len(runIfs))
				copy(newBundleRunIfs, runIfs)
			}
			if runIf != "" {
				newBundleRunIfs = append(newBundleRunIfs, runIf)
			}

			plans, err := builder.gatherBundleSteps(definition, uuid, envs, newBundleRunIfs)
			if err != nil {
				return nil, err
			}

			stepPlans = append(stepPlans, plans...)
		}
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) gatherBundleEnvs(bundleOverride StepBundleListItemModel, bundleDefinition StepBundleModel) ([]envmanModels.EnvironmentItemModel, error) {
	var bundleEnvs []envmanModels.EnvironmentItemModel

	bundleEnvs = append(bundleEnvs, bundleDefinition.Environments...)
	bundleEnvs = append(bundleEnvs, bundleOverride.Environments...)

	bundleEnvs = append(bundleEnvs, bundleDefinition.Inputs...)

	// Filter undefined bundleOverride inputs
	bundleDefinitionInputKeys := map[string]bool{}
	for _, input := range bundleDefinition.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return nil, err
		}

		bundleDefinitionInputKeys[key] = true
	}
	for _, input := range bundleOverride.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return nil, err
		}

		if _, ok := bundleDefinitionInputKeys[key]; ok {
			bundleEnvs = append(bundleEnvs, input)
		}
	}

	return bundleEnvs, nil
}
