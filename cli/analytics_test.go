package cli

import (
	"testing"

	"github.com/bitrise-io/envman/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func Test_expandStepInputsForAnalytics(t *testing.T) {
	type args struct {
		environments map[string]string
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
				environments: map[string]string{"secret_simulator_device": "secret_a_secret_b_secret_c"},
				inputs: []models.EnvironmentItemModel{
					{"secret_simulator_device": "secret_a_secret_b_secret_c"},
				},
				secretValues: []string{"secret_a_secret_b_secret_c"},
			},
			want: map[string]string{
				"secret_simulator_device": "[REDACTED]",
			},
		},
		{
			name: "Input is empty, and skip_if_empty is true",
			args: args{
				environments: map[string]string{},
				inputs: []models.EnvironmentItemModel{
					{"myinput": ""},
				},
			},
			want: map[string]string{
				"myinput": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := redactStepInputs(tt.args.environments, tt.args.inputs, tt.args.secretValues)
			require.NoError(t, err, "expandStepInputsForAnalytics")
			require.Equal(t, tt.want, got)
		})
	}
}
