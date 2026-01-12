package models

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/version"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const LogFormatVersion = "2"

// TODO: WorkflowRunPlan is used for both structured logging (bitrise_started event content) and for coordinating step execution.
//  The bitrise_started event has a strict structure, this is why some of the properties are annotated with `json:"-"` to avoid their serialization.
//  Models for these two usages should be separated in future.

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

type WithGroupPlan struct {
	Services  []ContainerPlan `json:"services,omitempty"`
	Container ContainerPlan   `json:"container,omitempty"`
}

type StepBundlePlan struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
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

type WorkflowRunModes struct {
	CIMode                  bool
	PRMode                  bool
	DebugMode               bool
	SecretFilteringMode     bool
	SecretEnvsFilteringMode bool
	NoOutputTimeout         time.Duration
	IsSteplibOfflineMode    bool
}

func NewWorkflowRunPlan(
	modes WorkflowRunModes, targetWorkflow string, workflows map[string]WorkflowModel,
	stepBundles map[string]StepBundleModel, containers map[string]Container, services map[string]Container,
	uuidProvider func() string,
) (WorkflowRunPlan, error) {
	builder := newWorkflowPlanBuilder(modes, workflows, stepBundles, containers, services, uuidProvider)

	workflowList := walkWorkflows(targetWorkflow, workflows, nil)
	for _, workflowID := range workflowList {
		if err := builder.processWorkflow(workflowID); err != nil {
			return WorkflowRunPlan{}, err
		}
	}

	return builder.build(), nil
}

func walkWorkflows(workflowID string, workflows map[string]WorkflowModel, workflowStack []string) []string {
	workflow := workflows[workflowID]
	for _, before := range workflow.BeforeRun {
		workflowStack = walkWorkflows(before, workflows, workflowStack)
	}

	workflowStack = append(workflowStack, workflowID)

	for _, after := range workflow.AfterRun {
		workflowStack = walkWorkflows(after, workflows, workflowStack)
	}

	return workflowStack
}

// workflowPlanBuilder encapsulates the state and logic for building a WorkflowRunPlan
type workflowPlanBuilder struct {
	modes        WorkflowRunModes
	workflows    map[string]WorkflowModel
	stepBundles  map[string]StepBundleModel
	containers   map[string]Container
	services     map[string]Container
	uuidProvider func() string

	// Collected state
	executionPlans  []WorkflowExecutionPlan
	withGroupPlans  map[string]WithGroupPlan
	stepBundlePlans map[string]StepBundlePlan
}

// newWorkflowPlanBuilder creates a new workflow plan builder
func newWorkflowPlanBuilder(
	modes WorkflowRunModes,
	workflows map[string]WorkflowModel,
	stepBundles map[string]StepBundleModel,
	containers, services map[string]Container,
	uuidProvider func() string,
) *workflowPlanBuilder {
	return &workflowPlanBuilder{
		modes:           modes,
		workflows:       workflows,
		stepBundles:     stepBundles,
		containers:      containers,
		services:        services,
		uuidProvider:    uuidProvider,
		executionPlans:  []WorkflowExecutionPlan{},
		withGroupPlans:  make(map[string]WithGroupPlan),
		stepBundlePlans: make(map[string]StepBundlePlan),
	}
}

// processWorkflow processes a single workflow and adds its execution plan
func (b *workflowPlanBuilder) processWorkflow(workflowID string) error {
	workflow := b.workflows[workflowID]
	var stepPlans []StepExecutionPlan

	for _, stepListItem := range workflow.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return err
		}

		if t == StepListItemTypeStep {
			plan, err := processRegularStep(key, stepListItem, b.uuidProvider)
			if err != nil {
				return err
			}
			stepPlans = append(stepPlans, plan)
		} else if t == StepListItemTypeWith {
			plans, groupID, err := processWithGroup(stepListItem, b.uuidProvider)
			if err != nil {
				return err
			}

			with, _ := stepListItem.GetWith()
			b.addWithGroup(groupID, with)

			stepPlans = append(stepPlans, plans...)
		} else if t == StepListItemTypeBundle {
			plans, err := processStepBundle(key, stepListItem, b.stepBundles, b.stepBundlePlans, b.uuidProvider)
			if err != nil {
				return err
			}
			stepPlans = append(stepPlans, plans...)
		}
	}

	workflowTitle := workflow.Title
	if workflowTitle == "" {
		workflowTitle = workflowID
	}

	b.executionPlans = append(b.executionPlans, WorkflowExecutionPlan{
		UUID:                 b.uuidProvider(),
		WorkflowID:           workflowID,
		Steps:                stepPlans,
		WorkflowTitle:        workflowTitle,
		IsSteplibOfflineMode: b.modes.IsSteplibOfflineMode,
	})

	return nil
}

// addWithGroup creates and stores a with group plan for the given group ID
func (b *workflowPlanBuilder) addWithGroup(groupID string, with *WithModel) {
	var containerPlan ContainerPlan
	if with.ContainerID != "" {
		containerPlan.Image = b.containers[with.ContainerID].Image
	}

	var servicePlans []ContainerPlan
	for _, serviceID := range with.ServiceIDs {
		servicePlans = append(servicePlans, ContainerPlan{Image: b.services[serviceID].Image})
	}

	b.withGroupPlans[groupID] = WithGroupPlan{
		Services:  servicePlans,
		Container: containerPlan,
	}
}

// build constructs the final WorkflowRunPlan
func (b *workflowPlanBuilder) build() WorkflowRunPlan {
	cliVersion := version.VERSION
	if version.IsAlternativeInstallation {
		cliVersion = fmt.Sprintf("%s (%s)", cliVersion, version.Commit)
	}

	return WorkflowRunPlan{
		Version:                 cliVersion,
		LogFormatVersion:        LogFormatVersion,
		CIMode:                  b.modes.CIMode,
		PRMode:                  b.modes.PRMode,
		DebugMode:               b.modes.DebugMode,
		IsSteplibOfflineMode:    b.modes.IsSteplibOfflineMode,
		NoOutputTimeoutMode:     b.modes.NoOutputTimeout > 0,
		SecretFilteringMode:     b.modes.SecretFilteringMode,
		SecretEnvsFilteringMode: b.modes.SecretEnvsFilteringMode,
		WithGroupPlans:          b.withGroupPlans,
		StepBundlePlans:         b.stepBundlePlans,
		ExecutionPlan:           b.executionPlans,
	}
}

// processRegularStep handles a single step (not in a with group or bundle)
func processRegularStep(stepID string, stepListItem StepListItemModel, uuidProvider func() string) (StepExecutionPlan, error) {
	step, err := stepListItem.GetStep()
	if err != nil {
		return StepExecutionPlan{}, err
	}

	return StepExecutionPlan{
		UUID:   uuidProvider(),
		StepID: stepID,
		Step:   *step,
	}, nil
}

// processWithGroup handles a 'with' group containing steps with containers/services
func processWithGroup(
	stepListItem StepListItemModel,
	uuidProvider func() string,
) ([]StepExecutionPlan, string, error) {
	with, err := stepListItem.GetWith()
	if err != nil {
		return nil, "", err
	}

	groupID := uuidProvider()
	var stepPlans []StepExecutionPlan

	for _, stepListStepItem := range with.Steps {
		stepID, step, err := stepListStepItem.GetStepIDAndStep()
		if err != nil {
			return nil, "", err
		}

		stepPlans = append(stepPlans, StepExecutionPlan{
			UUID:          uuidProvider(),
			StepID:        stepID,
			Step:          step,
			WithGroupUUID: groupID,
			ContainerID:   with.ContainerID,
			ServiceIDs:    with.ServiceIDs,
		})
	}

	return stepPlans, groupID, nil
}

// processStepBundle handles a step bundle and its nested steps
func processStepBundle(
	bundleID string,
	stepListItem StepListItemModel,
	stepBundles map[string]StepBundleModel,
	stepBundlePlans map[string]StepBundlePlan,
	uuidProvider func() string,
) ([]StepExecutionPlan, error) {
	bundleOverride, err := stepListItem.GetBundle()
	if err != nil {
		return nil, err
	}

	bundleDefinition, ok := stepBundles[bundleID]
	if !ok {
		return nil, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
	}

	bundleEnvs, err := gatherBundleEnvs(*bundleOverride, bundleDefinition)
	if err != nil {
		return nil, err
	}

	bundleUUID := uuidProvider()
	title := bundleDefinition.Title
	if bundleOverride.Title != "" {
		title = bundleOverride.Title
	}

	stepBundlePlans[bundleUUID] = StepBundlePlan{
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

	return gatherBundleSteps(bundleDefinition, bundleUUID, bundleEnvs, runIfs, stepBundles, stepBundlePlans, uuidProvider)
}

func gatherBundleSteps(
	bundleDefinition StepBundleModel,
	bundleUUID string,
	bundleEnvs []envmanModels.EnvironmentItemModel,
	runIfs []string,
	stepBundles map[string]StepBundleModel,
	stepBundlePlans map[string]StepBundlePlan,
	uuidProvider func() string,
) ([]StepExecutionPlan, error) {
	var stepPlans []StepExecutionPlan
	stepIDX := -1
	for _, stepListStepOrBundleItem := range bundleDefinition.Steps {
		key, t, err := stepListStepOrBundleItem.GetKeyAndType()
		if err != nil {
			return nil, err
		}

		if t == StepListItemTypeStep {
			stepIDX++
			step, err := stepListStepOrBundleItem.GetStep()
			if err != nil {
				return nil, err
			}

			stepID := key
			stepPlan := StepExecutionPlan{
				UUID:             uuidProvider(),
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

			definition, ok := stepBundles[bundleID]
			if !ok {
				return nil, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
			}

			envs, err := gatherBundleEnvs(*override, definition)
			if err != nil {
				return nil, err
			}
			envs = append(bundleEnvs, envs...)

			uuid := uuidProvider()
			title := definition.Title
			if override.Title != "" {
				title = override.Title
			}

			stepBundlePlans[uuid] = StepBundlePlan{
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

			plans, err := gatherBundleSteps(definition, uuid, envs, newBundleRunIfs, stepBundles, stepBundlePlans, uuidProvider)
			if err != nil {
				return nil, err
			}

			stepPlans = append(stepPlans, plans...)
		}
	}

	return stepPlans, nil
}

func gatherBundleEnvs(bundleOverride StepBundleListItemModel, bundleDefinition StepBundleModel) ([]envmanModels.EnvironmentItemModel, error) {
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
