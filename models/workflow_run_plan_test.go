package models

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/version"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cliVersion() string {
	if version.IsAlternativeInstallation {
		return fmt.Sprintf("%s (%s)", version.VERSION, version.Commit)
	}
	return version.VERSION
}

func TestNewWorkflowRunPlan_StepBundleRunIf(t *testing.T) {
	tests := []struct {
		name           string
		modes          WorkflowRunModes
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		services       map[string]Container
		want           WorkflowRunPlan
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name:           "steps within a bundle inherit the bundle's run_if condition",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					RunIf: "{{.IsCI}}",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{}},
						{"bundle::bundle1": StepBundleListItemModel{}},
						{"step2": stepmanModels.StepModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_2": {ID: "bundle1"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_6", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{}},
						{UUID: "uuid_3", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleRunIfs: []string{"{{.IsCI}}"}},
						{UUID: "uuid_4", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleRunIfs: []string{"{{.IsCI}}"}},
						{UUID: "uuid_5", StepID: "step2", Step: stepmanModels.StepModel{}},
					}},
				},
			},
		},
		{
			name:           "steps within embedded bundles inherit the all the parent bundles' run_if condition",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					RunIf: `{{enveq "RUN_IF_1" "true"}}`,
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle::bundle2": StepBundleListItemModel{}},
					},
				},
				"bundle2": {
					RunIf: `{{enveq "RUN_IF_2" "true"}}`,
					Steps: []StepListItemStepOrBundleModel{
						{"bundle2-step1": stepmanModels.StepModel{}},
						{"bundle::bundle3": StepBundleListItemModel{}},
					},
				},
				"bundle3": {
					RunIf: `{{enveq "RUN_IF_3" "true"}}`,
					Steps: []StepListItemStepOrBundleModel{
						{"bundle3-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
					"uuid_3": {ID: "bundle2"},
					"uuid_5": {ID: "bundle3"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_7", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", StepBundleRunIfs: []string{`{{enveq "RUN_IF_1" "true"}}`}},
						{UUID: "uuid_4", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3", StepBundleRunIfs: []string{`{{enveq "RUN_IF_1" "true"}}`, `{{enveq "RUN_IF_2" "true"}}`}},
						{UUID: "uuid_6", StepID: "bundle3-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_5", StepBundleRunIfs: []string{`{{enveq "RUN_IF_1" "true"}}`, `{{enveq "RUN_IF_2" "true"}}`, `{{enveq "RUN_IF_3" "true"}}`}},
					}},
				},
			},
		},
		{
			name:           "steps mixed with embedded bundles inside bundle",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::bundle2": StepBundleListItemModel{}},
						{"bundle1-step1": stepmanModels.StepModel{}},
					},
				},
				"bundle2": {
					RunIf: `{{enveq "RUN_IF_1" "true"}}`,
					Steps: []StepListItemStepOrBundleModel{
						{"bundle2-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
					"uuid_2": {ID: "bundle2"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_5", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleRunIfs: []string{`{{enveq "RUN_IF_1" "true"}}`}},
						{UUID: "uuid_4", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlan(tt.modes, tt.targetWorkflow, tt.workflows, tt.stepBundles, tt.containers, tt.services, (&MockUUIDProvider{}).UUID)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkflowRunPlan(t *testing.T) {
	tests := []struct {
		name           string
		modes          WorkflowRunModes
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		services       map[string]Container
		uuidProvider   func() string
		want           WorkflowRunPlan
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name:           "order of steps - step bundle used multiple times",
			modes:          WorkflowRunModes{},
			uuidProvider:   (&MockUUIDProvider{}).UUID,
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{}},
						{"bundle::bundle1": StepBundleListItemModel{}},
						{"step2": stepmanModels.StepModel{}},
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_2": {ID: "bundle1"},
					"uuid_6": {ID: "bundle1"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_9", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{}},
						{UUID: "uuid_3", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2"},
						{UUID: "uuid_4", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2"},
						{UUID: "uuid_5", StepID: "step2", Step: stepmanModels.StepModel{}},
						{UUID: "uuid_7", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_6"},
						{UUID: "uuid_8", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_6"},
					}},
				},
			},
		},
		{
			name:           "order of steps - nested step bundles",
			modes:          WorkflowRunModes{},
			uuidProvider:   (&MockUUIDProvider{}).UUID,
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{}},
					},
				},
				"bundle2": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
						{"bundle2-step1": stepmanModels.StepModel{}},
						{"bundle2-step2": stepmanModels.StepModel{}},
					},
				},
				"bundle3": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle3-step1": stepmanModels.StepModel{}},
						{"bundle::bundle2": StepBundleListItemModel{}},
						{"bundle3-step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle3": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle3"},
					"uuid_3": {ID: "bundle2"},
					"uuid_4": {ID: "bundle1"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_10", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle3-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
						{UUID: "uuid_5", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_4"},
						{UUID: "uuid_6", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_4"},
						{UUID: "uuid_7", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3"},
						{UUID: "uuid_8", StepID: "bundle2-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3"},
						{UUID: "uuid_9", StepID: "bundle3-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
					}},
				},
			},
		},
		{
			name:           "nested step bundle inputs",
			modes:          WorkflowRunModes{},
			uuidProvider:   (&MockUUIDProvider{}).UUID,
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input1": "value1"},
						{"input2": ""},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{}},
					},
				},
				"bundle2": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input1": "value3"},
						{"input3": ""},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::bundle1": StepBundleListItemModel{
							Inputs: []envmanModels.EnvironmentItemModel{
								{"input2": "value2"},
							},
						}},
						{"bundle2-step1": stepmanModels.StepModel{}},
						{"bundle2-step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle2": StepBundleListItemModel{
							Inputs: []envmanModels.EnvironmentItemModel{
								{"input3": "value3"},
							},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle2"},
					"uuid_2": {ID: "bundle1"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_7", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle2 definition inputs
							{"input1": "value3"}, {"input3": ""},
							// bundle2 override inputs
							{"input3": "value3"},
							// bundle1 definition inputs
							{"input1": "value1"}, {"input2": ""},
							// bundle1 override inputs
							{"input2": "value2"}}},
						{UUID: "uuid_4", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle2 definition inputs
							{"input1": "value3"}, {"input3": ""},
							// bundle2 override inputs
							{"input3": "value3"},
							// bundle1 definition inputs
							{"input1": "value1"}, {"input2": ""},
							// bundle1 override inputs
							{"input2": "value2"}}},
						{UUID: "uuid_5", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle2 definition inputs
							{"input1": "value3"}, {"input3": ""},
							// bundle2 override inputs
							{"input3": "value3"}}},
						{UUID: "uuid_6", StepID: "bundle2-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle2 definition inputs
							{"input1": "value3"}, {"input3": ""},
							// bundle2 override inputs
							{"input3": "value3"}}},
					}},
				},
			},
		},
		{
			name:           "step bundles title added to the plan",
			modes:          WorkflowRunModes{},
			uuidProvider:   (&MockUUIDProvider{}).UUID,
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Title: "My Bundle 1",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1", Title: "My Bundle 1"},
					"uuid_3": {ID: "bundle1", Title: "My Bundle 1"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_5", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
						{UUID: "uuid_4", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3"},
					}},
				},
			},
		},
		{
			name:           "step bundles title overrides definition title",
			modes:          WorkflowRunModes{},
			uuidProvider:   (&MockUUIDProvider{}).UUID,
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					Title: "My Bundle 1",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
						{"bundle::bundle1": StepBundleListItemModel{Title: "My Bundle Override 1"}},
						{"bundle::bundle1": StepBundleListItemModel{Title: "My Bundle Override 2"}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans:   map[string]ContainerPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1", Title: "My Bundle 1"},
					"uuid_3": {ID: "bundle1", Title: "My Bundle Override 1"},
					"uuid_5": {ID: "bundle1", Title: "My Bundle Override 2"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_7", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
						{UUID: "uuid_4", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3"},
						{UUID: "uuid_6", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_5"},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlan(tt.modes, tt.targetWorkflow, tt.workflows, tt.stepBundles, tt.containers, tt.services, tt.uuidProvider)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkflowRunPlan_StepAndStepBundleContainers(t *testing.T) {
	tests := []struct {
		name           string
		modes          WorkflowRunModes
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		services       map[string]Container
		want           WorkflowRunPlan
	}{
		{
			name:           "step with container stores container in execution plan",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"step-container": {Image: "alpine:latest"},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{
							ContainerID: "step-container",
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans:  map[string]StepBundlePlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"step-container": {Image: "alpine:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{ContainerID: "step-container"}, ContainerID: "step-container"},
					}},
				},
			},
		},
		{
			name:           "step with services stores services in execution plan",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			services: map[string]Container{
				"service1": {Image: "postgres:latest"},
				"service2": {Image: "redis:latest"},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{
							ServiceIDs: []string{"service1", "service2"},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:                 cliVersion(),
				LogFormatVersion:        "2",
				WithGroupPlans:          map[string]WithGroupPlan{},
				StepBundlePlans:         map[string]StepBundlePlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans: map[string]ContainerPlan{
					"service1": {Image: "postgres:latest"},
					"service2": {Image: "redis:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{ServiceIDs: []string{"service1", "service2"}}, ServiceIDs: []string{"service1", "service2"}},
					}},
				},
			},
		},
		{
			name:           "step with both container and services",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"step-container": {Image: "alpine:latest"},
			},
			services: map[string]Container{
				"service1": {Image: "postgres:latest"},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{
							ContainerID: "step-container",
							ServiceIDs:  []string{"service1"},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans:  map[string]StepBundlePlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"step-container": {Image: "alpine:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"service1": {Image: "postgres:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{ContainerID: "step-container", ServiceIDs: []string{"service1"}}, ContainerID: "step-container", ServiceIDs: []string{"service1"}},
					}},
				},
			},
		},
		{
			name:           "step bundle with container propagates to bundle steps",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"bundle-container": {Image: "alpine:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					ContainerID: "bundle-container",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"bundle-container": {Image: "alpine:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ContainerID: "bundle-container"},
						{UUID: "uuid_3", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ContainerID: "bundle-container"},
					}},
				},
			},
		},
		{
			name:           "step bundle with services propagates to bundle steps",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			services: map[string]Container{
				"service1": {Image: "postgres:latest"},
				"service2": {Image: "redis:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					ServiceIDs: []string{"service1", "service2"},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans: map[string]ContainerPlan{
					"service1": {Image: "postgres:latest"},
					"service2": {Image: "redis:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_3", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ServiceIDs: []string{"service1", "service2"}},
					}},
				},
			},
		},
		{
			name:           "step in bundle with own container overrides bundle container",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"bundle-container": {Image: "alpine:latest"},
				"step-container":   {Image: "ubuntu:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					ContainerID: "bundle-container",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{
							ContainerID: "step-container",
						}},
						{"bundle1-step3": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"bundle-container": {Image: "alpine:latest"},
					"step-container":   {Image: "ubuntu:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_5", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ContainerID: "bundle-container"},
						{UUID: "uuid_3", StepID: "bundle1-step2", Step: stepmanModels.StepModel{ContainerID: "step-container"}, StepBundleUUID: "uuid_1", ContainerID: "step-container"},
						{UUID: "uuid_4", StepID: "bundle1-step3", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ContainerID: "bundle-container"},
					}},
				},
			},
		},
		{
			name:           "step in bundle with own services overrides bundle services",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			services: map[string]Container{
				"bundle-service": {Image: "postgres:latest"},
				"step-service":   {Image: "redis:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"bundle1": {
					ServiceIDs: []string{"bundle-service"},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle1-step1": stepmanModels.StepModel{}},
						{"bundle1-step2": stepmanModels.StepModel{
							ServiceIDs: []string{"step-service"},
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{},
				ServiceContainerPlans: map[string]ContainerPlan{
					"bundle-service": {Image: "postgres:latest"},
					"step-service":   {Image: "redis:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", ServiceIDs: []string{"bundle-service"}},
						{UUID: "uuid_3", StepID: "bundle1-step2", Step: stepmanModels.StepModel{ServiceIDs: []string{"step-service"}}, StepBundleUUID: "uuid_1", ServiceIDs: []string{"step-service"}},
					}},
				},
			},
		},
		{
			name:           "nested bundle inherits parent bundle container when not defined",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"parent-container": {Image: "alpine:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"parent-bundle": {
					ContainerID: "parent-container",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::child-bundle": StepBundleListItemModel{}},
					},
				},
				"child-bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"child-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::parent-bundle": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "parent-bundle"},
					"uuid_2": {ID: "child-bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"parent-container": {Image: "alpine:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "child-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", ContainerID: "parent-container"},
					}},
				},
			},
		},
		{
			name:           "nested bundle with own container overrides parent bundle container",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"parent-container": {Image: "alpine:latest"},
				"child-container":  {Image: "ubuntu:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"parent-bundle": {
					ContainerID: "parent-container",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::child-bundle": StepBundleListItemModel{}},
					},
				},
				"child-bundle": {
					ContainerID: "child-container",
					Steps: []StepListItemStepOrBundleModel{
						{"child-step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::parent-bundle": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "parent-bundle"},
					"uuid_2": {ID: "child-bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"parent-container": {Image: "alpine:latest"},
					"child-container":  {Image: "ubuntu:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "child-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", ContainerID: "child-container"},
					}},
				},
			},
		},
		{
			name:           "step in nested bundle with own container overrides all parent containers",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"parent-container": {Image: "alpine:latest"},
				"child-container":  {Image: "ubuntu:latest"},
				"step-container":   {Image: "debian:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"parent-bundle": {
					ContainerID: "parent-container",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::child-bundle": StepBundleListItemModel{}},
					},
				},
				"child-bundle": {
					ContainerID: "child-container",
					Steps: []StepListItemStepOrBundleModel{
						{"child-step1": stepmanModels.StepModel{}},
						{"child-step2": stepmanModels.StepModel{
							ContainerID: "step-container",
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::parent-bundle": StepBundleListItemModel{}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "parent-bundle"},
					"uuid_2": {ID: "child-bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"parent-container": {Image: "alpine:latest"},
					"child-container":  {Image: "ubuntu:latest"},
					"step-container":   {Image: "debian:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_5", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "child-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", ContainerID: "child-container"},
						{UUID: "uuid_4", StepID: "child-step2", Step: stepmanModels.StepModel{ContainerID: "step-container"}, StepBundleUUID: "uuid_2", ContainerID: "step-container"},
					}},
				},
			},
		},
		{
			name:           "mixed steps with and without containers",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"step1-container": {Image: "alpine:latest"},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"step1": stepmanModels.StepModel{
							ContainerID: "step1-container",
						}},
						{"step2": stepmanModels.StepModel{}},
						{"step3": stepmanModels.StepModel{
							ContainerID: "step1-container",
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans:  map[string]StepBundlePlan{},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"step1-container": {Image: "alpine:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_1", StepID: "step1", Step: stepmanModels.StepModel{ContainerID: "step1-container"}, ContainerID: "step1-container"},
						{UUID: "uuid_2", StepID: "step2", Step: stepmanModels.StepModel{}},
						{UUID: "uuid_3", StepID: "step3", Step: stepmanModels.StepModel{ContainerID: "step1-container"}, ContainerID: "step1-container"},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlan(tt.modes, tt.targetWorkflow, tt.workflows, tt.stepBundles, tt.containers, tt.services, (&MockUUIDProvider{}).UUID)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

type MockUUIDProvider struct {
	i int
}

func (m *MockUUIDProvider) UUID() string {
	m.i++
	return fmt.Sprintf("uuid_%d", m.i)
}
