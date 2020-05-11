package cli

import (
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

type EmptyEnvironment struct{}

func (EmptyEnvironment) GetEnvironment() map[string]string {
	return map[string]string{}
}

func Test_prepareStepEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		params  prepareStepInputParams
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Template expansion works",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{},
				inputs: []envmanModels.EnvironmentItemModel{
					{"D": "{{.IsCI}}", "opts": map[string]interface{}{"is_template": true}},
				},
				isCIMode: true,
			},
			want: map[string]string{
				"D": "true",
			},
		},
		{
			name: "Default expansion flag is applied",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{
					{"A": "B", "opts": map[string]interface{}{}},
					{"myinput": "$A", "opts": map[string]interface{}{}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"myinput": "$A", "opts": map[string]interface{}{}},
				},
			},
			want: map[string]string{
				"A":       "B",
				"myinput": "B",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareStepEnvironment(tt.params, EmptyEnvironment{})
			if tt.wantErr {
				require.Error(t, err, "prepareStepEnvironment() expected to return error")
			} else {
				require.NoError(t, err, "prepareStepEnvironment()")
			}
			require.Equal(t, tt.want, got, "prepareStepEnvironment() result mismatch")
		})
	}
}
