package cli

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

const mixedStagedAndDagPipeline = `
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

const missingWorkflowDagPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [c] }
workflows:
  a: {}
  b: {}
`

const duplicatedDependencyDagPipeline = `
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

const utilityWorkflowDagPipeline = `
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

const dagWithCyclePipeline = `
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

const validDagPipeline = `
format_version: '13'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a] }
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

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "Mixing stages and workflows in the same pipeline",
			config:  mixedStagedAndDagPipeline,
			wantErr: "Failed to get config (bitrise.yml) from base 64 data, err: Failed to parse bitrise config, error: pipeline (dag) has both stages and workflows",
		},
		{
			name:    "Workflow is missing from the DAG pipeline",
			config:  missingWorkflowDagPipeline,
			wantErr: "Failed to get config (bitrise.yml) from base 64 data, err: Failed to parse bitrise config, error: workflow (c) defined in dependencies (b), but does not exist in the pipeline (dag)",
		},
		{
			name:    "Utility workflow is referenced in the DAG pipeline",
			config:  utilityWorkflowDagPipeline,
			wantErr: "Failed to get config (bitrise.yml) from base 64 data, err: Failed to parse bitrise config, error: workflow (_a) defined in pipeline (dag), is a utility workflow",
		},
		{
			name:    "Duplicated dependency in the DAG pipeline",
			config:  duplicatedDependencyDagPipeline,
			wantErr: "Failed to get config (bitrise.yml) from base 64 data, err: Failed to parse bitrise config, error: workflow (a) is duplicated in the dependency list (b)",
		},
		{
			name:    "Cycle in the DAG pipeline",
			config:  dagWithCyclePipeline,
			wantErr: "Failed to get config (bitrise.yml) from base 64 data, err: Failed to parse bitrise config, error: the dependency between workflow 'c' and workflow 'b' creates a cycle in the graph",
		},
		{
			name:    "Valid DAG pipeline",
			config:  validDagPipeline,
			wantErr: "",
		},
		{
			name:    "Valid staged pipeline",
			config:  validStagedPipeline,
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b64Data := base64.StdEncoding.EncodeToString([]byte(tt.config))
			_, _, err := CreateBitriseConfigFromCLIParams(b64Data, "")

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
