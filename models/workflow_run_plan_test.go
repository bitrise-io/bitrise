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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
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
			name:           "order of steps -  nested step bundles",
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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
				StepBundlePlans: map[string]StepBundlePlan{
					"uuid_1": {ID: "bundle3"},
					"uuid_3": {ID: "bundle2"},
					"uuid_4": {ID: "bundle1"},
				},
				ExecutionPlan: []WorkflowExecutionPlan{
					{UUID: "uuid_10", WorkflowID: "workflow1", WorkflowTitle: "workflow1", Steps: []StepExecutionPlan{
						{UUID: "uuid_2", StepID: "bundle3-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
						{UUID: "uuid_5", StepID: "bundle1-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_4", StepBundleRunIfs: []string{}},
						{UUID: "uuid_6", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_4", StepBundleRunIfs: []string{}},
						{UUID: "uuid_7", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3", StepBundleRunIfs: []string{}},
						{UUID: "uuid_8", StepID: "bundle2-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_3", StepBundleRunIfs: []string{}},
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
				Version:          version.VERSION,
				LogFormatVersion: "2",
				WithGroupPlans:   map[string]WithGroupPlan{},
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
							{"input2": "value2"}}, StepBundleRunIfs: []string{}},
						{UUID: "uuid_4", StepID: "bundle1-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_2", StepBundleRunIfs: []string{}},
						{UUID: "uuid_5", StepID: "bundle2-step1", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1", StepBundleEnvs: []envmanModels.EnvironmentItemModel{
							// bundle2 definition inputs
							{"input1": "value3"}, {"input3": ""},
							// bundle2 override inputs
							{"input3": "value3"}}},
						{UUID: "uuid_6", StepID: "bundle2-step2", Step: stepmanModels.StepModel{}, StepBundleUUID: "uuid_1"},
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

type MockUUIDProvider struct {
	i int
}

func (m *MockUUIDProvider) UUID() string {
	m.i++
	return fmt.Sprintf("uuid_%d", m.i)
}
