package cli

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func Test_prepareStepEnvironment(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Prepare: failed to create temp dir: %s", err)
	}
	inputEnvStorePath := path.Join(tempDir, ".envstore")

	tests := []struct {
		name    string
		params  prepareStepInputParams
		want    []envmanModels.EnvironmentItemModel
		wantErr bool
	}{
		{
			name: "",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{
					{"A": "B", "opts": map[string]interface{}{}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"C": "$A", "opts": map[string]interface{}{}},
					{"D": "{{.IsCI}}", "opts": map[string]interface{}{"is_template": true}},
				},
				inputEnvstorePath: inputEnvStorePath,
				isCIMode:          true,
			},
			want: []envmanModels.EnvironmentItemModel{
				{"A": "B", "opts": map[string]interface{}{}},
				{"C": "$A", "opts": map[string]interface{}{}},
				{"D": "true", "opts": map[string]interface{}{"is_template": true}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareStepEnvironment(tt.params)
			if tt.wantErr {
				require.Error(t, err, "prepareStepEnvironment() expected to return error")
			} else {
				require.NoError(t, err, "prepareStepEnvironment()")
			}
			require.False(t, (err != nil) != tt.wantErr, fmt.Sprintf("prepareStepEnvironment() error = %v, wantErr %v", err, tt.wantErr))
			require.Equal(t, tt.want, got, "prepareStepEnvironment() result mismatch")
		})
	}
}
