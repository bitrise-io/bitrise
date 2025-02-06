package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkflowRunPlan(t *testing.T) {
	type args struct {
		modes          WorkflowRunModes
		targetWorkflow string
		workflows      map[string]WorkflowModel
		stepBundles    map[string]StepBundleModel
		containers     map[string]Container
		services       map[string]Container
		uuidProvider   func() string
	}
	tests := []struct {
		name    string
		args    args
		want    WorkflowRunPlan
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkflowRunPlan(tt.args.modes, tt.args.targetWorkflow, tt.args.workflows, tt.args.stepBundles, tt.args.containers, tt.args.services, tt.args.uuidProvider)
			if !tt.wantErr(t, err, fmt.Sprintf("NewWorkflowRunPlan(%v, %v, %v, %v, %v, %v, %v)", tt.args.modes, tt.args.targetWorkflow, tt.args.workflows, tt.args.stepBundles, tt.args.containers, tt.args.services, tt.args.uuidProvider)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewWorkflowRunPlan(%v, %v, %v, %v, %v, %v, %v)", tt.args.modes, tt.args.targetWorkflow, tt.args.workflows, tt.args.stepBundles, tt.args.containers, tt.args.services, tt.args.uuidProvider)
		})
	}
}
