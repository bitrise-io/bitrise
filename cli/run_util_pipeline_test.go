package cli

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/bitrise"
)

const validStagedPipeline = `
format_version: '13'
pipelines:
  staged:
    stages:
    - s1: {}
    - s2: {}
    - s3: {}
stages:
  s1:
    workflows:
    - a: {}
  s2:
    workflows:
    - b: {}
    - c: {}
    - d: {}
  s3:
    workflows:
    - e: {}
workflows:
  a: {}
  b: {}
  c: {}
  d: {}
  e: {}
`

const validDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      a1: { uses: a, inputs: [key: value] }
      b: { depends_on: [a], parallel: 3 }
      c: { depends_on: [a] }
      d: { depends_on: [a] }
      e: { depends_on: [b, d] }
      f: { depends_on: [e] }
      g: { depends_on: [a, e, f] }
workflows:
  a: {}
  b: {}
  c: {}
  d: {}
  e: {}
  f: {}
  g: {}
`

const mixedStagedAndDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a] }
    stages:
    - stage1: {}
stages:
  stage1:
    workflows:
    - a: {}
    - b: {}
workflows:
  a: {}
  b: {}
`

const missingWorkflowInDAGPipelineDefinition = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [c] }
workflows:
  a: {}
  b: {}
  c: {}
`

const missingWorkflowInWorkflowDefinitionForDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [c] }
      c: {}
workflows:
  a: {}
  b: {}
`

const missingWorkflowInWorkflowVariantDefinitionForDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { uses: c }
workflows:
  a: {}
`

const workflowVariantHasTheSameNameAsAnExistingWorkflowForDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { uses: c }
workflows:
  a: {}
  b: {}
  c: {}
`

const generatedParallelWorkflowVariantHasTheSameNameAsAnExistingWorkflowForDAGPipeline = `
format_version: '13'
pipelines:
 dag:
    workflows:
      a: { parallel: 3}
workflows:
  a: {}
  a_3: {}
`

const parallelValueIsNonPositive = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: { parallel: 0}
workflows:
  a: {}
`

const parallelValueIsNonInteger = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: { parallel: true}
workflows:
  a: {}
`

const duplicatedDependencyDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a, a] }
workflows:
  a: {}
  b: {}
`

const utilityWorkflowDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      _a: {}
      b: { depends_on: [_a] }
workflows:
  _a: {}
  b: {}
`

const cycleInDAGPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [c] }
      c: { depends_on: [b] }
workflows:
  a: {}
  b: {}
  c: {}
`

const emptyPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows: {}
workflows:
  a: {}
`

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "Mixing stages and workflows in the same pipeline",
			config:  mixedStagedAndDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: pipeline (dag) has both stages and workflows",
		},
		{
			name:    "Workflow is missing from the DAG pipeline definition",
			config:  missingWorkflowInDAGPipelineDefinition,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (c) defined in dependencies (b) is not part of pipeline (dag)",
		},
		{
			name:    "Workflow is missing from the Workflow definition",
			config:  missingWorkflowInWorkflowDefinitionForDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (c) defined in pipeline (dag) is not found in the workflow definitions",
		},
		{
			name:    "Workflow is missing from the Workflow Variant definition",
			config:  missingWorkflowInWorkflowVariantDefinitionForDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (c) referenced in pipeline (dag) in workflow variant (b) is not found in the workflow definitions",
		},
		{
			name:    "Workflow variant has the same name as an existing workflow",
			config:  workflowVariantHasTheSameNameAsAnExistingWorkflowForDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (b) defined in pipeline (dag) is a variant of another workflow, but it is also defined as a workflow",
		},
		{
			name:    "Generated parallel workflow variant has the same name as an existing workflow",
			config:  generatedParallelWorkflowVariantHasTheSameNameAsAnExistingWorkflowForDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: parallel workflow variant (a_3) would be generated by workflow (a) defined in pipeline (dag), but it is also defined as a workflow",
		},
		{
			name:    "Parallel value is non-positive", // We let the Pipeline Service define the upper limit for the parallel value
			config:  parallelValueIsNonPositive,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (a) defined in pipeline (dag) has invalid parallel value (0), should be at least 1",
		},
		{
			name:    "Parallel value is non-integer",
			config:  parallelValueIsNonInteger,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: yaml: unmarshal errors:\n  line 6: cannot unmarshal !!bool `true` into int",
		},
		{
			name:    "Utility workflow is referenced in the DAG pipeline",
			config:  utilityWorkflowDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (_a) defined in pipeline (dag) is a utility workflow",
		},
		{
			name:    "Duplicated dependency in the DAG pipeline",
			config:  duplicatedDependencyDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: workflow (a) is duplicated in the dependency list (b)",
		},
		{
			name:    "Cycle in the DAG pipeline",
			config:  cycleInDAGPipeline,
			wantErr: "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: the dependency between workflow 'b' and workflow 'c' creates a cycle in the graph",
		},
		{
			name:   "Valid DAG pipeline",
			config: validDAGPipeline,
		},
		{
			name:   "Valid staged pipeline",
			config: validStagedPipeline,
		},
		{
			name:   "Empty pipeline",
			config: emptyPipeline,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b64Data := base64.StdEncoding.EncodeToString([]byte(tt.config))
			_, _, err := CreateBitriseConfigFromCLIParams(b64Data, "", bitrise.ValidationTypeFull)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
