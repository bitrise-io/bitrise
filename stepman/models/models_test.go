package models

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func Test_serialize_StepModel(t *testing.T) {
	step := StepModel{
		Title: pointers.NewStringPtr("Test Step"),
		Toolkit: &StepToolkitModel{
			Go: &GoStepToolkitModel{
				PackageName: "go/package",
			},
			Bash: &BashStepToolkitModel{
				EntryFile: "step.sh",
			},
		},
	}

	// JSON
	{
		bytes, err := json.Marshal(step)
		require.NoError(t, err)
		require.Equal(t, `{"title":"Test Step","source":{},"toolkit":{"bash":{"entry_file":"step.sh"},"go":{"package_name":"go/package"}},"deps":{}}`, string(bytes))
	}

	// YAML
	{
		bytes, err := yaml.Marshal(step)
		require.NoError(t, err)
		t.Log("t: ", string(bytes))
		require.Equal(t, `title: Test Step
toolkit:
  bash:
    entry_file: step.sh
  go:
    package_name: go/package
`,
			string(bytes))
	}
}
