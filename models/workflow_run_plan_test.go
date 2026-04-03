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
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_2": {ID: "bundle1"},
				},
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
			got, err := NewWorkflowRunPlanBuilder(tt.workflows, tt.stepBundles, tt.containers, tt.services, (&MockUUIDProvider{}).UUID).Build(tt.modes, tt.targetWorkflow)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkflowRunPlan_StepBundleInputs(t *testing.T) {
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
			name:           "nested step bundle inputs",
			modes:          WorkflowRunModes{},
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
			name:           "nested bundle input references parent bundle input",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle2": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input": "$input"},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle2-step1": stepmanModels.StepModel{}},
					},
				},
				"bundle1": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input": ""},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::bundle2": StepBundleListItemModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{
							Inputs: []envmanModels.EnvironmentItemModel{
								{"input": "test_value"},
							},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
					"uuid_2": {ID: "bundle2"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle1 definition inputs
							{"input": ""},
							// bundle1 override inputs
							{"input": "test_value"},
							// bundle2 definition inputs
							{"input": "$input"}}},
					}},
				},
			},
		},
		{
			name:           "parent bundle passes input value to child bundle on embedding",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"bundle2": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input": ""},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle2-step1": stepmanModels.StepModel{}},
					},
				},
				"bundle1": {
					Inputs: []envmanModels.EnvironmentItemModel{
						{"input": ""},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::bundle2": StepBundleListItemModel{
							Inputs: []envmanModels.EnvironmentItemModel{
								{"input": "$input"},
							},
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::bundle1": StepBundleListItemModel{
							Inputs: []envmanModels.EnvironmentItemModel{
								{"input": "test_value"},
							},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle1"},
					"uuid_2": {ID: "bundle2"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle1 definition inputs
							{"input": ""},
							// bundle1 override inputs
							{"input": "test_value"},
							// bundle2 definition inputs are skipped because "input" is already defined in parent
							// bundle2 override inputs
							{"input": "$input"}}},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlanBuilder(tt.workflows, tt.stepBundles, tt.containers, tt.services, (&MockUUIDProvider{}).UUID).Build(tt.modes, tt.targetWorkflow)
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
				Version:          cliVersion(),
				LogFormatVersion: "2",
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
			got, err := NewWorkflowRunPlanBuilder(tt.workflows, tt.stepBundles, tt.containers, tt.services, tt.uuidProvider).Build(tt.modes, tt.targetWorkflow)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewWorkflowRunPlan_Containers(t *testing.T) {
	tests := []struct {
		name           string
		modes          WorkflowRunModes
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		services       map[string]Container
		want           WorkflowRunPlan
		wantErr        string
	}{
		{
			name:           "step with execution_container - simple string format",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"alpine": {
					Type:  ContainerTypeExecution,
					Image: "alpine:latest",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ExecutionContainer: "alpine",
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				ExecutionContainerPlans: map[string]ContainerPlan{
					"alpine": {Image: "alpine:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:               "uuid_1",
							StepID:             "script",
							Step:               stepmanModels.StepModel{ExecutionContainer: "alpine"},
							ExecutionContainer: &ContainerConfig{ContainerID: "alpine", Recreate: false},
						},
					}},
				},
			},
		},
		{
			name:           "step with execution_container - recreate flag",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"golang": {
					Type:  ContainerTypeExecution,
					Image: "golang:1.22",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ExecutionContainer: map[string]any{
								"golang": map[any]any{
									"recreate": true,
								},
							},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				ExecutionContainerPlans: map[string]ContainerPlan{
					"golang": {Image: "golang:1.22"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:               "uuid_1",
							StepID:             "script",
							Step:               stepmanModels.StepModel{ExecutionContainer: map[string]any{"golang": map[any]any{"recreate": true}}},
							ExecutionContainer: &ContainerConfig{ContainerID: "golang", Recreate: true},
						},
					}},
				},
			},
		},
		{
			name:           "step with service_containers - simple string format",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"redis": {
					Type:  ContainerTypeService,
					Image: "redis:latest",
				},
				"postgres": {
					Type:  ContainerTypeService,
					Image: "postgres:13",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{"redis", "postgres"},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis":    {Image: "redis:latest"},
					"postgres": {Image: "postgres:13"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:   "uuid_1",
							StepID: "script",
							Step:   stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{"redis", "postgres"}},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis", Recreate: false},
								{ContainerID: "postgres", Recreate: false},
							},
						},
					}},
				},
			},
		},
		{
			name:           "step with service_containers - recreate flag",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"redis": {
					Type:  ContainerTypeService,
					Image: "redis:latest",
				},
				"postgres": {
					Type:  ContainerTypeService,
					Image: "postgres:13",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{
								"redis",
								map[string]any{
									"postgres": map[string]any{
										"recreate": true,
									},
								},
							},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis":    {Image: "redis:latest"},
					"postgres": {Image: "postgres:13"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:   "uuid_1",
							StepID: "script",
							Step: stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{
								"redis",
								map[string]any{
									"postgres": map[string]any{
										"recreate": true,
									},
								},
							}},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis", Recreate: false},
								{ContainerID: "postgres", Recreate: true},
							},
						},
					}},
				},
			},
		},
		{
			name:           "step with both execution_container and service_containers",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"golang": {
					Type:  ContainerTypeExecution,
					Image: "golang:1.22",
				},
				"redis": {
					Type:  ContainerTypeService,
					Image: "redis:latest",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ExecutionContainer: "golang",
							ServiceContainers:  []stepmanModels.ContainerReference{"redis"},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				ExecutionContainerPlans: map[string]ContainerPlan{
					"golang": {Image: "golang:1.22"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis": {Image: "redis:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_2", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:   "uuid_1",
							StepID: "script",
							Step: stepmanModels.StepModel{
								ExecutionContainer: "golang",
								ServiceContainers:  []stepmanModels.ContainerReference{"redis"},
							},
							ExecutionContainer: &ContainerConfig{ContainerID: "golang", Recreate: false},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis", Recreate: false},
							},
						},
					}},
				},
			},
		},
		{
			name:           "step bundle with execution_container",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"test_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle-step1": stepmanModels.StepModel{}},
						{"bundle-step2": stepmanModels.StepModel{}},
					},
				},
			},
			containers: map[string]Container{
				"node": {
					Type:  ContainerTypeExecution,
					Image: "node:18",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::test_bundle": StepBundleListItemModel{
							ExecutionContainer: "node",
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "test_bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"node": {Image: "node:18"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_4", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:               "uuid_2",
							StepID:             "bundle-step1",
							Step:               stepmanModels.StepModel{},
							StepBundleUUID:     "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "node", Recreate: false},
						},
						{
							UUID:               "uuid_3",
							StepID:             "bundle-step2",
							Step:               stepmanModels.StepModel{},
							StepBundleUUID:     "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "node", Recreate: false},
						},
					}},
				},
			},
		},
		{
			name:           "step bundle with service_containers",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			stepBundles: map[string]StepBundleModel{
				"test_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"bundle-step1": stepmanModels.StepModel{}},
					},
				},
			},
			containers: map[string]Container{
				"redis": {
					Type:  ContainerTypeService,
					Image: "redis:latest",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"bundle::test_bundle": StepBundleListItemModel{
							ServiceContainers: []stepmanModels.ContainerReference{"redis"},
						}},
					},
				},
			},
			want: WorkflowRunPlan{
				Version:          cliVersion(),
				LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "test_bundle"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis": {Image: "redis:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_3", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:           "uuid_2",
							StepID:         "bundle-step1",
							Step:           stepmanModels.StepModel{},
							StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis", Recreate: false},
							},
						},
					}},
				},
			},
		},
		{
			name:           "error - undefined execution container referenced",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers:     map[string]Container{},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ExecutionContainer: "undefined_container",
						}},
					},
				},
			},
			wantErr: "referenced execution container not defined: undefined_container",
		},
		{
			name:           "error - undefined service container referenced",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers:     map[string]Container{},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{"undefined_service"},
						}},
					},
				},
			},
			wantErr: "referenced service container not defined: undefined_service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlanBuilder(tt.workflows, tt.stepBundles, tt.containers, tt.services, (&MockUUIDProvider{}).UUID).Build(tt.modes, tt.targetWorkflow)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewWorkflowRunPlan_ContainerNesting(t *testing.T) {
	tests := []struct {
		name           string
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		want           WorkflowRunPlan
		wantErr        string
	}{
		// Override rule: usage-side replaces definition-side
		{
			name:           "bundle: definition-side execution container used when no usage override",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"golang": {Type: ContainerTypeExecution, Image: "golang:1.22"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ExecutionContainer: "golang",
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:         map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ExecutionContainerPlans: map[string]ContainerPlan{"golang": {Image: "golang:1.22"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_3", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "golang"}},
					},
				}},
			},
		},
		{
			name:           "bundle: definition-side service containers used when no usage override",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"redis": {Type: ContainerTypeService, Image: "redis:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"redis"},
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:       map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ServiceContainerPlans: map[string]ContainerPlan{"redis": {Image: "redis:latest"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_3", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "redis"}}},
					},
				}},
			},
		},
		{
			name:           "bundle: usage-side execution container completely overrides definition-side",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"golang_1": {Type: ContainerTypeExecution, Image: "golang:1.22"},
				"golang_2": {Type: ContainerTypeExecution, Image: "golang:1.23"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ExecutionContainer: "golang_1", // definition
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{
						ExecutionContainer: "golang_2", // usage overrides
					}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:         map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ExecutionContainerPlans: map[string]ContainerPlan{"golang_2": {Image: "golang:1.23"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_3", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "golang_2"}},
					},
				}},
			},
		},
		{
			name:           "bundle: usage-side service containers completely replace definition-side",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"redis":    {Type: ContainerTypeService, Image: "redis:latest"},
				"postgres": {Type: ContainerTypeService, Image: "postgres:13"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"redis"}, // definition
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{
						ServiceContainers: []stepmanModels.ContainerReference{"postgres"}, // usage replaces
					}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:       map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ServiceContainerPlans: map[string]ContainerPlan{"postgres": {Image: "postgres:13"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_3", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "postgres"}}},
					},
				}},
			},
		},

		// Step-level container rules inside a bundle
		{
			name:           "step: own execution container overrides bundle-inherited container",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"golang": {Type: ContainerTypeExecution, Image: "golang:1.22"},
				"alpine": {Type: ContainerTypeExecution, Image: "alpine:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ExecutionContainer: "golang",
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
						{"step2": stepmanModels.StepModel{ExecutionContainer: "alpine"}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"golang": {Image: "golang:1.22"},
					"alpine": {Image: "alpine:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "golang"}},
						{UUID: "uuid_3", StepID: "step2", StepBundleUUID: "uuid_1",
							Step:               stepmanModels.StepModel{ExecutionContainer: "alpine"},
							ExecutionContainer: &ContainerConfig{ContainerID: "alpine"}},
					},
				}},
			},
		},
		{
			name:           "step: own service containers accumulate with bundle service containers",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"redis":    {Type: ContainerTypeService, Image: "redis:latest"},
				"postgres": {Type: ContainerTypeService, Image: "postgres:13"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"redis"},
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
						{"step2": stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{"postgres"}}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis":    {Image: "redis:latest"},
					"postgres": {Image: "postgres:13"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "redis"}}},
						{UUID: "uuid_3", StepID: "step2", StepBundleUUID: "uuid_1",
							Step: stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{"postgres"}},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis"},
								{ContainerID: "postgres"},
							}},
					},
				}},
			},
		},

		// Nested bundle rules
		{
			name:           "nested bundles: inner inherits execution container from outer",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"golang": {Type: ContainerTypeExecution, Image: "golang:1.22"},
			},
			stepBundles: map[string]StepBundleModel{
				"inner_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"inner_step": stepmanModels.StepModel{}},
					},
				},
				"outer_bundle": {
					ExecutionContainer: "golang",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::inner_bundle": StepBundleListItemModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::outer_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "outer_bundle"},
					"uuid_2": {ID: "inner_bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{"golang": {Image: "golang:1.22"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "inner_step", StepBundleUUID: "uuid_2",
							ExecutionContainer: &ContainerConfig{ContainerID: "golang"}},
					},
				}},
			},
		},
		{
			name:           "nested bundles: inner own execution container takes priority over outer",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"golang_outer": {Type: ContainerTypeExecution, Image: "golang:1.22"},
				"golang_inner": {Type: ContainerTypeExecution, Image: "golang:1.23"},
			},
			stepBundles: map[string]StepBundleModel{
				"inner_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"inner_step": stepmanModels.StepModel{}},
					},
				},
				"outer_bundle": {
					ExecutionContainer: "golang_outer",
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::inner_bundle": StepBundleListItemModel{
							ExecutionContainer: "golang_inner",
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::outer_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "outer_bundle"},
					"uuid_2": {ID: "inner_bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"golang_outer": {Image: "golang:1.22"},
					"golang_inner": {Image: "golang:1.23"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "inner_step", StepBundleUUID: "uuid_2",
							ExecutionContainer: &ContainerConfig{ContainerID: "golang_inner"}},
					},
				}},
			},
		},
		{
			name:           "nested bundles: service containers accumulate from outer to inner",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"redis":    {Type: ContainerTypeService, Image: "redis:latest"},
				"postgres": {Type: ContainerTypeService, Image: "postgres:13"},
			},
			stepBundles: map[string]StepBundleModel{
				"inner_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"postgres"},
					Steps: []StepListItemStepOrBundleModel{
						{"inner_step": stepmanModels.StepModel{}},
					},
				},
				"outer_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"redis"},
					Steps: []StepListItemStepOrBundleModel{
						{"bundle::inner_bundle": StepBundleListItemModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::outer_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "outer_bundle"},
					"uuid_2": {ID: "inner_bundle"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"redis":    {Image: "redis:latest"},
					"postgres": {Image: "postgres:13"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_3", StepID: "inner_step", StepBundleUUID: "uuid_2",
							ServiceContainers: []ContainerConfig{
								{ContainerID: "redis"},
								{ContainerID: "postgres"},
							}},
					},
				}},
			},
		},

		// Bundle-level recreate rules
		{
			name:           "bundle: exec container recreate applies only to first step",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"exec": {Type: ContainerTypeExecution, Image: "exec:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ExecutionContainer: map[string]any{"exec": map[string]any{"recreate": true}},
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
						{"step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:         map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ExecutionContainerPlans: map[string]ContainerPlan{"exec": {Image: "exec:latest"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "exec", Recreate: true}},
						{UUID: "uuid_3", StepID: "step2", StepBundleUUID: "uuid_1",
							ExecutionContainer: &ContainerConfig{ContainerID: "exec"}},
					},
				}},
			},
		},
		{
			name:           "bundle: service container recreate applies only to first step",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"svc": {Type: ContainerTypeService, Image: "svc:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{
						map[string]any{"svc": map[string]any{"recreate": true}},
					},
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
						{"step2": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans:       map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ServiceContainerPlans: map[string]ContainerPlan{"svc": {Image: "svc:latest"}},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "svc", Recreate: true}}},
						{UUID: "uuid_3", StepID: "step2", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "svc"}}},
					},
				}},
			},
		},
		{
			name:           "step: own service container recreate overrides inherited bundle container",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"svc1": {Type: ContainerTypeService, Image: "svc1:latest"},
				"svc2": {Type: ContainerTypeService, Image: "svc2:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"my_bundle": {
					ServiceContainers: []stepmanModels.ContainerReference{"svc1", "svc2"},
					Steps: []StepListItemStepOrBundleModel{
						{"step1": stepmanModels.StepModel{}},
						{"step2": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{
								map[string]any{"svc1": map[string]any{"recreate": true}},
							},
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"bundle::my_bundle": StepBundleListItemModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{"uuid_1": {ID: "my_bundle"}},
				ServiceContainerPlans: map[string]ContainerPlan{
					"svc1": {Image: "svc1:latest"},
					"svc2": {Image: "svc2:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_4", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "step1", StepBundleUUID: "uuid_1",
							ServiceContainers: []ContainerConfig{{ContainerID: "svc1"}, {ContainerID: "svc2"}}},
						{UUID: "uuid_3", StepID: "step2", StepBundleUUID: "uuid_1",
							Step: stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{
								map[string]any{"svc1": map[string]any{"recreate": true}},
							}},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "svc1", Recreate: true},
								{ContainerID: "svc2"},
							}},
					},
				}},
			},
		},
		{
			// Complex recreate example
			name:           "bundle: complex recreate - exec and service recreate applied only at bundle boundary",
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"def_exec":  {Type: ContainerTypeExecution, Image: "inner-def:latest"},
				"def_svc_1": {Type: ContainerTypeService, Image: "inner-def-svc:latest"},
				"def_svc_2": {Type: ContainerTypeService, Image: "inner-def-svc:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"bundle": {
					ExecutionContainer: map[string]any{"def_exec": map[string]any{"recreate": true}},
					ServiceContainers: []stepmanModels.ContainerReference{
						"def_svc_1",
						map[string]any{"def_svc_2": map[string]any{"recreate": true}},
					},
					Steps: []StepListItemStepOrBundleModel{
						// step_1 has a plain def_svc_2 reference (no recreate); the bundle's recreate
						// still applies for the first step via OR semantics in mergeServiceContainers.
						{"step_1": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{"def_svc_2"},
						}},
						{"step_2": stepmanModels.StepModel{ExecutionContainer: "def_exec"}},
						{"step_3": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{
								map[string]any{"def_svc_1": map[string]any{"recreate": true}},
							},
						}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {Steps: []StepListItemModel{
					{"step_4": stepmanModels.StepModel{}},
					{"bundle::bundle": StepBundleListItemModel{}},
					{"step_5": stepmanModels.StepModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{"uuid_2": {ID: "bundle"}},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"def_exec": {Image: "inner-def:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"def_svc_1": {Image: "inner-def-svc:latest"},
					"def_svc_2": {Image: "inner-def-svc:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_7", WorkflowID: "workflow1", WorkflowTitle: "workflow1",
					Steps: []StepExecutionPlan{
						// top-level step before bundle — no containers
						{UUID: "uuid_1", StepID: "step_4"},
						// first step in bundle — exec + svc2 with recreate (bundle boundary);
						// step_1's plain def_svc_2 reference does not suppress the bundle's recreate (OR semantics)
						{UUID: "uuid_3", StepID: "step_1", StepBundleUUID: "uuid_2",
							Step: stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{
								"def_svc_2",
							}},
							ExecutionContainer: &ContainerConfig{ContainerID: "def_exec", Recreate: true},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "def_svc_1"},
								{ContainerID: "def_svc_2", Recreate: true},
							}},
						// second step — own exec (no recreate), bundle svcs without recreate
						{UUID: "uuid_4", StepID: "step_2", StepBundleUUID: "uuid_2",
							Step:               stepmanModels.StepModel{ExecutionContainer: "def_exec"},
							ExecutionContainer: &ContainerConfig{ContainerID: "def_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "def_svc_1"},
								{ContainerID: "def_svc_2"},
							}},
						// third step — inherited exec (no recreate); step's own svc1 recreate wins via OR
						{UUID: "uuid_5", StepID: "step_3", StepBundleUUID: "uuid_2",
							Step: stepmanModels.StepModel{ServiceContainers: []stepmanModels.ContainerReference{
								map[string]any{"def_svc_1": map[string]any{"recreate": true}},
							}},
							ExecutionContainer: &ContainerConfig{ContainerID: "def_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "def_svc_1", Recreate: true},
								{ContainerID: "def_svc_2"},
							}},
						// top-level step after bundle — no containers
						{UUID: "uuid_6", StepID: "step_5"},
					},
				}},
			},
		},

		// Complex nesting example
		{
			name:           "complex nesting: override and service accumulation across two bundle levels",
			targetWorkflow: "wf",
			containers: map[string]Container{
				"outer_exec":  {Type: ContainerTypeExecution, Image: "outer_exec:latest"},
				"inner_exec":  {Type: ContainerTypeExecution, Image: "inner_exec:latest"},
				"step_exec":   {Type: ContainerTypeExecution, Image: "step_exec:latest"},
				"outer_svc_1": {Type: ContainerTypeService, Image: "outer_svc_1:latest"},
				"outer_svc_2": {Type: ContainerTypeService, Image: "outer_svc_2:latest"},
				"inner_svc":   {Type: ContainerTypeService, Image: "inner_svc:latest"},
				"step_svc_1":  {Type: ContainerTypeService, Image: "step_svc_1:latest"},
				"step_svc_2":  {Type: ContainerTypeService, Image: "step_svc_2:latest"},
			},
			stepBundles: map[string]StepBundleModel{
				"inner_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"step_1": stepmanModels.StepModel{}},
						{"step_2": stepmanModels.StepModel{
							ExecutionContainer: "step_exec",
							ServiceContainers:  []stepmanModels.ContainerReference{"step_svc_1", "step_svc_2"},
						}},
						{"step_3": stepmanModels.StepModel{}},
					},
				},
				"outer_bundle": {
					Steps: []StepListItemStepOrBundleModel{
						{"step_4": stepmanModels.StepModel{}},
						{"bundle::inner_bundle": StepBundleListItemModel{
							ExecutionContainer: "inner_exec",
							ServiceContainers:  []stepmanModels.ContainerReference{"inner_svc"},
						}},
						{"step_5": stepmanModels.StepModel{}},
					},
				},
			},
			workflows: map[string]WorkflowModel{
				"wf": {Steps: []StepListItemModel{
					{"step_6": stepmanModels.StepModel{}},
					{"bundle::outer_bundle": StepBundleListItemModel{
						ExecutionContainer: "outer_exec",
						ServiceContainers:  []stepmanModels.ContainerReference{"outer_svc_1", "outer_svc_2"},
					}},
					{"step_7": stepmanModels.StepModel{}},
				}},
			},
			want: WorkflowRunPlan{
				Version: cliVersion(), LogFormatVersion: "2",
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_2": {ID: "outer_bundle"},
					"uuid_4": {ID: "inner_bundle"},
				},
				ExecutionContainerPlans: map[string]ContainerPlan{
					"outer_exec": {Image: "outer_exec:latest"},
					"inner_exec": {Image: "inner_exec:latest"},
					"step_exec":  {Image: "step_exec:latest"},
				},
				ServiceContainerPlans: map[string]ContainerPlan{
					"outer_svc_1": {Image: "outer_svc_1:latest"},
					"outer_svc_2": {Image: "outer_svc_2:latest"},
					"inner_svc":   {Image: "inner_svc:latest"},
					"step_svc_1":  {Image: "step_svc_1:latest"},
					"step_svc_2":  {Image: "step_svc_2:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{{UUID: "uuid_10", WorkflowID: "wf", WorkflowTitle: "wf",
					Steps: []StepExecutionPlan{
						// Top-level step, no bundle, no containers
						{UUID: "uuid_1", StepID: "step_6"},
						// step_4: outer bundle's exec/svcs (no own containers)
						{UUID: "uuid_3", StepID: "step_4", StepBundleUUID: "uuid_2",
							ExecutionContainer: &ContainerConfig{ContainerID: "outer_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "outer_svc_1"},
								{ContainerID: "outer_svc_2"},
							}},
						// step_1: inner bundle's exec + outer+inner svcs (no own containers)
						{UUID: "uuid_5", StepID: "step_1", StepBundleUUID: "uuid_4",
							ExecutionContainer: &ContainerConfig{ContainerID: "inner_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "outer_svc_1"},
								{ContainerID: "outer_svc_2"},
								{ContainerID: "inner_svc"},
							}},
						// step_2: own exec + outer+inner+step svcs
						{UUID: "uuid_6", StepID: "step_2", StepBundleUUID: "uuid_4",
							Step: stepmanModels.StepModel{
								ExecutionContainer: "step_exec",
								ServiceContainers:  []stepmanModels.ContainerReference{"step_svc_1", "step_svc_2"},
							},
							ExecutionContainer: &ContainerConfig{ContainerID: "step_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "outer_svc_1"},
								{ContainerID: "outer_svc_2"},
								{ContainerID: "inner_svc"},
								{ContainerID: "step_svc_1"},
								{ContainerID: "step_svc_2"},
							}},
						// step_3: inner bundle's exec + outer+inner svcs (no own containers)
						{UUID: "uuid_7", StepID: "step_3", StepBundleUUID: "uuid_4",
							ExecutionContainer: &ContainerConfig{ContainerID: "inner_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "outer_svc_1"},
								{ContainerID: "outer_svc_2"},
								{ContainerID: "inner_svc"},
							}},
						// step_5: outer bundle's exec/svcs (no own containers)
						{UUID: "uuid_8", StepID: "step_5", StepBundleUUID: "uuid_2",
							ExecutionContainer: &ContainerConfig{ContainerID: "outer_exec"},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "outer_svc_1"},
								{ContainerID: "outer_svc_2"},
							}},
						// Top-level step, no bundle, no containers
						{UUID: "uuid_9", StepID: "step_7"},
					},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlanBuilder(tt.workflows, tt.stepBundles, tt.containers, nil, (&MockUUIDProvider{}).UUID).Build(WorkflowRunModes{}, tt.targetWorkflow)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
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

func cliVersion() string {
	if version.IsAlternativeInstallation {
		return fmt.Sprintf("%s (%s)", version.VERSION, version.Commit)
	}
	return version.VERSION
}
