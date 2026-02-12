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
			name:           "untyped container can be used as both execution and service",
			modes:          WorkflowRunModes{},
			targetWorkflow: "workflow1",
			containers: map[string]Container{
				"alpine": {
					// No type specified - can be used as both
					Image: "alpine:latest",
				},
			},
			workflows: map[string]WorkflowModel{
				"workflow1": {
					Steps: []StepListItemModel{
						{"script1": stepmanModels.StepModel{
							ExecutionContainer: "alpine",
						}},
						{"script2": stepmanModels.StepModel{
							ServiceContainers: []stepmanModels.ContainerReference{"alpine"},
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
				ServiceContainerPlans: map[string]ContainerPlan{
					"alpine": {Image: "alpine:latest"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_3", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{
							UUID:               "uuid_1",
							StepID:             "script1",
							Step:               stepmanModels.StepModel{ExecutionContainer: "alpine"},
							ExecutionContainer: &ContainerConfig{ContainerID: "alpine", Recreate: false},
						},
						{
							UUID:   "uuid_2",
							StepID: "script2",
							Step: stepmanModels.StepModel{
								ServiceContainers: []stepmanModels.ContainerReference{"alpine"},
							},
							ServiceContainers: []ContainerConfig{
								{ContainerID: "alpine", Recreate: false},
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
