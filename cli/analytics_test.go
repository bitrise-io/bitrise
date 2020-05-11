package cli

import (
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func Test_expandStepInputsForAnalytics(t *testing.T) {
	type args struct {
		environments []envmanModels.EnvironmentItemModel
		inputs       []envmanModels.EnvironmentItemModel
		secretValues []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Secret filtering",
			args: args{
				environments: []envmanModels.EnvironmentItemModel{
					{"secret_simulator_device": "secret_a_secret_b_secret_c", "opts": map[string]interface{}{"is_sensitive": false}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"secret_simulator_device": ""},
				},
				secretValues: []string{"secret_a_secret_b_secret_c"},
			},
			want: map[string]string{
				"secret_simulator_device": "[REDACTED]",
			},
		},
		{
			name: "Default expansion flag is applied",
			args: args{
				environments: []envmanModels.EnvironmentItemModel{
					{"A": "B", "opts": map[string]interface{}{}},
					{"myinput": "$A", "opts": map[string]interface{}{}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"myinput": "$A", "opts": map[string]interface{}{}},
				},
				secretValues: []string{},
			},
			want: map[string]string{
				"myinput": "B",
			},
		},
		{
			name: "Input is empty, and skip_if_empty is true",
			args: args{
				environments: []envmanModels.EnvironmentItemModel{
					{"myinput": "", "opts": map[string]interface{}{"skip_if_empty": true}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"myinput": "", "opts": map[string]interface{}{"skip_if_empty": true}},
				},
				secretValues: []string{},
			},
			want: map[string]string{
				"myinput": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandStepInputsForAnalytics(tt.args.environments, tt.args.inputs, tt.args.secretValues)
			require.NoError(t, err, "expandStepInputsForAnalytics")
			require.Equal(t, tt.want, got)
		})
	}
}
