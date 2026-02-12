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
	StepBundlePlans map[string]StepBundlePlan `json:"step_bundles,omitempty"`

	// ----
	// Container plans
	// step execution container id to container plan
	ExecutionContainerPlans map[string]ContainerPlan `json:"execution_container_plans,omitempty"`
	// service container id to container plan
	ServiceContainerPlans map[string]ContainerPlan `json:"service_container_plans,omitempty"`
	// ----

	// TODO: Old container plan, to be removed when containerisation using "With groups" is sunsetted
	WithGroupPlans map[string]WithGroupPlan `json:"with_groups,omitempty"`
}

type WorkflowExecutionPlan struct {
	UUID                 string              `json:"uuid"`
	WorkflowID           string              `json:"workflow_id"`
	Steps                []StepExecutionPlan `json:"steps"`
	WorkflowTitle        string              `json:"-"`
	IsSteplibOfflineMode bool                `json:"-"`
}

// WithGroupPlan ...
// TODO: Old container plan, to be removed when containerisation using "With groups" is sunsetted
type WithGroupPlan struct {
	Services  []ContainerPlan `json:"services,omitempty"`
	Container ContainerPlan   `json:"container,omitempty"`
}

type StepBundlePlan struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type StepExecutionPlan struct {
	UUID   string                  `json:"uuid"`
	StepID string                  `json:"step_id"`
	Step   stepmanModels.StepModel `json:"-"`

	// Step Bundle group
	StepBundleUUID string                              `json:"step_bundle_uuid,omitempty"`
	StepBundleEnvs []envmanModels.EnvironmentItemModel `json:"-"`
	// StepBundleRunIfs stores each run_if statements of the including Step Bundles.
	// The first element is the run_if statement of the top most Step Bundle including the given Step.
	// To execute the Step, all run_if statements must be evaluated to true.
	StepBundleRunIfs []string `json:"-"`

	// Containers
	ExecutionContainer *ContainerConfig  `json:"execution_container,omitempty"`
	ServiceContainers  []ContainerConfig `json:"service_containers,omitempty"`

	// With (container) group
	WithGroupUUID string   `json:"with_group_uuid,omitempty"`
	ContainerID   string   `json:"-"`
	ServiceIDs    []string `json:"-"`
}

type ContainerPlan struct {
	Image string `json:"image"`
}

type ContainerConfig struct {
	ContainerID string `json:"container_id,omitempty"`
	Recreate    bool   `json:"_"`
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

type BundleContext struct {
	UUID   string
	Envs   []envmanModels.EnvironmentItemModel
	RunIfs []string
}

type WithGroupContext struct {
	UUID        string
	ContainerID string
	ServiceIDs  []string
}

type WorkflowRunPlanBuilder struct {
	workflows    map[string]WorkflowModel
	stepBundles  map[string]StepBundleModel
	containers   map[string]Container
	services     map[string]Container
	uuidProvider func() string

	stepBundlePlans         map[string]StepBundlePlan
	executionContainerPlans map[string]ContainerPlan
	serviceContainerPlans   map[string]ContainerPlan
	withGroupPlans          map[string]WithGroupPlan
}

func NewWorkflowRunPlanBuilder(workflows map[string]WorkflowModel, stepBundles map[string]StepBundleModel, containers map[string]Container, services map[string]Container, uuidProvider func() string) *WorkflowRunPlanBuilder {
	return &WorkflowRunPlanBuilder{
		workflows:               workflows,
		stepBundles:             stepBundles,
		containers:              containers,
		services:                services,
		uuidProvider:            uuidProvider,
		stepBundlePlans:         map[string]StepBundlePlan{},
		executionContainerPlans: map[string]ContainerPlan{},
		serviceContainerPlans:   map[string]ContainerPlan{},
		withGroupPlans:          map[string]WithGroupPlan{},
	}
}

func (builder *WorkflowRunPlanBuilder) Build(modes WorkflowRunModes, targetWorkflowID string) (WorkflowRunPlan, error) {
	var executionPlan []WorkflowExecutionPlan

	workflowList := builder.walkWorkflows(targetWorkflowID, builder.workflows, nil)
	for _, workflowID := range workflowList {
		workflow := builder.workflows[workflowID]

		var stepPlans []StepExecutionPlan
		for _, stepListItem := range workflow.Steps {
			genericStep, err := NewStepListItemFromWorkflowStep(stepListItem)
			if err != nil {
				return WorkflowRunPlan{}, err
			}

			plans, err := builder.processStepListItem(*genericStep, nil, nil, true)
			if err != nil {
				return WorkflowRunPlan{}, err
			}

			stepPlans = append(stepPlans, plans...)
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

	plan := WorkflowRunPlan{
		Version:                 cliVersion,
		LogFormatVersion:        LogFormatVersion,
		CIMode:                  modes.CIMode,
		PRMode:                  modes.PRMode,
		DebugMode:               modes.DebugMode,
		IsSteplibOfflineMode:    modes.IsSteplibOfflineMode,
		NoOutputTimeoutMode:     modes.NoOutputTimeout > 0,
		SecretFilteringMode:     modes.SecretFilteringMode,
		SecretEnvsFilteringMode: modes.SecretEnvsFilteringMode,
		ExecutionPlan:           executionPlan,
	}
	if len(builder.withGroupPlans) > 0 {
		plan.WithGroupPlans = builder.withGroupPlans
	}
	if len(builder.stepBundlePlans) > 0 {
		plan.StepBundlePlans = builder.stepBundlePlans
	}
	if len(builder.executionContainerPlans) > 0 {
		plan.ExecutionContainerPlans = builder.executionContainerPlans
	}
	if len(builder.serviceContainerPlans) > 0 {
		plan.ServiceContainerPlans = builder.serviceContainerPlans
	}

	return plan, nil
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

func (builder *WorkflowRunPlanBuilder) processStepListItem(stepListItem StepListItem, stepBundleContext *BundleContext, withGroupContext *WithGroupContext, allowContainers bool) ([]StepExecutionPlan, error) {
	var stepPlans []StepExecutionPlan

	key, t := stepListItem.GetKeyAndType()

	switch t {
	case StepListItemTypeStep:
		plan, err := builder.processStep(key, stepListItem, stepBundleContext, withGroupContext, allowContainers)
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, *plan)
	case StepListItemTypeWith:
		plans, err := builder.processWithGroup(stepListItem)
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, plans...)
	case StepListItemTypeBundle:
		plans, err := builder.processStepBundle(key, stepListItem, stepBundleContext, allowContainers)
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, plans...)
	default:
		return nil, fmt.Errorf("unknown step list item type")
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) processStep(stepID string, stepListItem StepListItem, bundleContext *BundleContext, withGroupContext *WithGroupContext, allowContainerDefinition bool) (*StepExecutionPlan, error) {
	step := stepListItem.GetStep()

	plan := StepExecutionPlan{
		UUID:   builder.uuidProvider(),
		StepID: stepID,
		Step:   *step,
	}
	if bundleContext != nil {
		plan.StepBundleUUID = bundleContext.UUID
		plan.StepBundleEnvs = bundleContext.Envs
		plan.StepBundleRunIfs = bundleContext.RunIfs
	}
	if withGroupContext != nil {
		plan.WithGroupUUID = withGroupContext.UUID
		plan.ContainerID = withGroupContext.ContainerID
		plan.ServiceIDs = withGroupContext.ServiceIDs
	}
	if allowContainerDefinition {
		executionContainerCfg, serviceContainerCfgs, err := builder.processContainerConfigs(newContainerisableFromStep(*step))
		if err != nil {
			return nil, err
		}

		plan.ExecutionContainer = executionContainerCfg
		plan.ServiceContainers = serviceContainerCfgs

	}
	return &plan, nil
}

func (builder *WorkflowRunPlanBuilder) processStepBundle(bundleID string, stepListItem StepListItem, bundleContext *BundleContext, allowContainerDefinition bool) ([]StepExecutionPlan, error) {
	bundleOverride := stepListItem.GetBundle()

	bundleDefinition, ok := builder.stepBundles[bundleID]
	if !ok {
		return nil, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
	}

	// Collect parent input keys to avoid overriding them with child's default values
	var parentInputKeys map[string]bool
	if bundleContext != nil {
		parentInputKeys = make(map[string]bool)
		for _, env := range bundleContext.Envs {
			key, _, err := env.GetKeyValuePair()
			if err != nil {
				return nil, err
			}
			parentInputKeys[key] = true
		}
	}

	// Collect Bundle Envs
	bundleEnvs, err := builder.gatherBundleEnvs(*bundleOverride, bundleDefinition, parentInputKeys)
	if err != nil {
		return nil, err
	}
	if bundleContext != nil {
		bundleEnvs = append(bundleContext.Envs, bundleEnvs...)
	}

	// Collect Bundle runIfs
	runIf := bundleDefinition.RunIf
	if bundleOverride.RunIf != nil {
		runIf = *bundleOverride.RunIf
	}

	// Create a new runIfs slice that includes the runIf of the current bundle, instead of modifying the original slice.
	// This is necessary to ensure that the runIfs of the current bundle are evaluated correctly in the context of the parent bundle.
	// The Go slice wraps a pointer to the actual data inside.
	// So passing it around and adding items to it would update all the slices internal data storage.
	var runIfs []string
	if bundleContext != nil && len(bundleContext.RunIfs) > 0 {
		runIfs = make([]string, len(bundleContext.RunIfs))
		copy(runIfs, bundleContext.RunIfs)
	}
	if runIf != "" {
		runIfs = append(runIfs, runIf)
	}

	// Register Bundle Plan
	bundleUUID := builder.uuidProvider()
	title := bundleDefinition.Title
	if bundleOverride.Title != "" {
		title = bundleOverride.Title
	}

	builder.stepBundlePlans[bundleUUID] = StepBundlePlan{
		ID:    bundleID,
		Title: title,
	}

	// Process Bundle Steps
	newBundleContext := BundleContext{
		UUID:   bundleUUID,
		Envs:   bundleEnvs,
		RunIfs: runIfs,
	}
	plans, err := builder.gatherBundleSteps(bundleDefinition, newBundleContext)
	if err != nil {
		return nil, err
	}

	if allowContainerDefinition {
		executionContainerCfg, serviceContainerCfgs, err := builder.processContainerConfigs(newContainerisableFromStepBundle(*bundleOverride))
		if err != nil {
			return nil, err
		}

		for i := range plans {
			plans[i].ExecutionContainer = executionContainerCfg
			plans[i].ServiceContainers = serviceContainerCfgs
		}
	}

	return plans, nil
}

func (builder *WorkflowRunPlanBuilder) processWithGroup(stepListItem StepListItem) ([]StepExecutionPlan, error) {
	with := stepListItem.GetWithGroup()

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

	withGroupContext := &WithGroupContext{
		UUID:        groupID,
		ContainerID: with.ContainerID,
		ServiceIDs:  with.ServiceIDs,
	}

	var stepPlans []StepExecutionPlan
	for _, stepListStepItem := range with.Steps {
		genericStep, err := NewStepListItemFromWithStep(stepListStepItem)
		if err != nil {
			return nil, err
		}

		plans, err := builder.processStepListItem(*genericStep, nil, withGroupContext, false)
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, plans...)
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) gatherBundleSteps(bundleDefinition StepBundleModel, bundleContext BundleContext) ([]StepExecutionPlan, error) {
	var stepPlans []StepExecutionPlan
	for _, stepListStepOrBundleItem := range bundleDefinition.Steps {
		genericStep, err := NewStepListItemFromBundleStep(stepListStepOrBundleItem)
		if err != nil {
			return nil, err
		}

		plans, err := builder.processStepListItem(*genericStep, &bundleContext, nil, false)
		if err != nil {
			return nil, err
		}

		stepPlans = append(stepPlans, plans...)
	}

	return stepPlans, nil
}

func (builder *WorkflowRunPlanBuilder) gatherBundleEnvs(bundleOverride StepBundleListItemModel, bundleDefinition StepBundleModel, parentInputKeys map[string]bool) ([]envmanModels.EnvironmentItemModel, error) {
	var bundleEnvs []envmanModels.EnvironmentItemModel

	bundleEnvs = append(bundleEnvs, bundleDefinition.Environments...)
	bundleEnvs = append(bundleEnvs, bundleOverride.Environments...)

	// Collect override input keys
	bundleOverrideInputKeys := map[string]bool{}
	for _, input := range bundleOverride.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return nil, err
		}
		bundleOverrideInputKeys[key] = true
	}

	// Add definition inputs, but skip keys that are already defined in parent context
	// AND will be overridden by this bundle's override inputs
	for _, input := range bundleDefinition.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return nil, err
		}

		// Skip this definition input if:
		// 1. It's already defined in parent context, AND
		// 2. This bundle has an override input for the same key
		// This prevents the definition's default value from overwriting the parent's value
		// before the override input is applied
		if parentInputKeys != nil && parentInputKeys[key] && bundleOverrideInputKeys[key] {
			continue
		}
		bundleEnvs = append(bundleEnvs, input)
	}

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

func (builder *WorkflowRunPlanBuilder) processContainerConfigs(containerisable Containerisable) (*ContainerConfig, []ContainerConfig, error) {
	executionContainers := map[string]Container{}
	serviceContainers := map[string]Container{}
	for id, container := range builder.containers {
		switch container.Type {
		case ContainerTypeExecution:
			executionContainers[id] = container
		case ContainerTypeService:
			serviceContainers[id] = container
		default:
			executionContainers[id] = container
			serviceContainers[id] = container
		}
	}

	executionContainerCfg, err := containerisable.GetExecutionContainerConfig()
	if err != nil {
		return nil, nil, err
	}

	if executionContainerCfg != nil {
		container, ok := executionContainers[executionContainerCfg.ContainerID]
		if !ok {
			return nil, nil, fmt.Errorf("referenced execution container not defined: %s", executionContainerCfg.ContainerID)
		}

		if _, ok := builder.executionContainerPlans[executionContainerCfg.ContainerID]; !ok {
			builder.executionContainerPlans[executionContainerCfg.ContainerID] = ContainerPlan{
				Image: container.Image,
			}
		}
	}

	serviceContainerCfgs, err := containerisable.GetServiceContainerConfigs()
	if err != nil {
		return nil, nil, err
	}

	for _, containerCfg := range serviceContainerCfgs {
		container, ok := serviceContainers[containerCfg.ContainerID]
		if !ok {
			return nil, nil, fmt.Errorf("referenced service container not defined: %s", containerCfg.ContainerID)
		}

		if _, ok := builder.serviceContainerPlans[containerCfg.ContainerID]; !ok {
			builder.serviceContainerPlans[containerCfg.ContainerID] = ContainerPlan{
				Image: container.Image,
			}
		}
	}

	return executionContainerCfg, serviceContainerCfgs, nil
}
