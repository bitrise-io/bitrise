package models

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/version"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const LogFormatVersion = "2"

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

func NewWorkflowRunPlan(
	modes WorkflowRunModes, targetWorkflow string, workflows map[string]WorkflowModel,
	stepBundles map[string]StepBundleModel, containers map[string]Container, services map[string]Container,
	uuidProvider func() string,
) (WorkflowRunPlan, error) {
	var executionPlan []WorkflowExecutionPlan
	withGroupPlans := map[string]WithGroupPlan{}
	stepBundlePlans := map[string]StepBundlePlan{}

	workflowList := walkWorkflows(targetWorkflow, workflows, nil)
	for _, workflowID := range workflowList {
		workflow := workflows[workflowID]

		var stepPlans []StepExecutionPlan

		for _, stepListItem := range workflow.Steps {
			key, t, err := stepListItem.GetKeyAndType()
			if err != nil {
				return WorkflowRunPlan{}, err
			}

			if t == StepListItemTypeStep {
				step, err := stepListItem.GetStep()
				if err != nil {
					return WorkflowRunPlan{}, err
				}

				stepID := key
				stepPlans = append(stepPlans, StepExecutionPlan{
					UUID:   uuidProvider(),
					StepID: stepID,
					Step:   *step,
				})
			} else if t == StepListItemTypeWith {
				with, err := stepListItem.GetWith()
				if err != nil {
					return WorkflowRunPlan{}, err
				}

				groupID := uuidProvider()

				var containerPlan ContainerPlan
				if with.ContainerID != "" {
					containerPlan.Image = containers[with.ContainerID].Image
				}

				var servicePlans []ContainerPlan
				for _, serviceID := range with.ServiceIDs {
					servicePlans = append(servicePlans, ContainerPlan{Image: services[serviceID].Image})
				}

				withGroupPlans[groupID] = WithGroupPlan{
					Services:  servicePlans,
					Container: containerPlan,
				}

				for _, stepListStepItem := range with.Steps {
					stepID, step, err := stepListStepItem.GetStepIDAndStep()
					if err != nil {
						return WorkflowRunPlan{}, err
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
			} else if t == StepListItemTypeBundle {
				bundleID := key
				bundleOverride, err := stepListItem.GetBundle()
				if err != nil {
					return WorkflowRunPlan{}, err
				}

				bundleDefinition, ok := stepBundles[bundleID]
				if !ok {
					return WorkflowRunPlan{}, fmt.Errorf("referenced step bundle not defined: %s", bundleID)
				}

				bundleEnvs, err := gatherBundleEnvs(*bundleOverride, bundleDefinition)
				if err != nil {
					return WorkflowRunPlan{}, err
				}

				bundleUUID := uuidProvider()

				stepBundlePlans[bundleUUID] = StepBundlePlan{
					ID: bundleID,
				}

				plans, err := gatherBundleSteps(bundleDefinition, bundleUUID, bundleEnvs, stepBundles, stepBundlePlans, uuidProvider)
				if err != nil {
					return WorkflowRunPlan{}, err
				}

				stepPlans = append(stepPlans, plans...)
			}
		}

		workflowTitle := workflow.Title
		if workflowTitle == "" {
			workflowTitle = workflowID
		}

		executionPlan = append(executionPlan, WorkflowExecutionPlan{
			UUID:                 uuidProvider(),
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
		WithGroupPlans:          withGroupPlans,
		StepBundlePlans:         stepBundlePlans,
		ExecutionPlan:           executionPlan,
	}, nil
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

func gatherBundleSteps(bundleDefinition StepBundleModel, bundleUUID string, bundleEnvs []envmanModels.EnvironmentItemModel, stepBundles map[string]StepBundleModel, stepBundlePlans map[string]StepBundlePlan, uuidProvider func() string) ([]StepExecutionPlan, error) {
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
				UUID:           uuidProvider(),
				StepID:         stepID,
				Step:           *step,
				StepBundleUUID: bundleUUID,
			}

			if stepIDX == 0 {
				stepPlan.StepBundleEnvs = bundleEnvs
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

			stepBundlePlans[uuid] = StepBundlePlan{
				ID: bundleID,
			}

			plans, err := gatherBundleSteps(definition, uuid, envs, stepBundles, stepBundlePlans, uuidProvider)
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
