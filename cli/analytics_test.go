package cli

import (
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func Test_redactStepInputs(t *testing.T) {
	type args struct {
		environments map[string]string
		inputs       []models.EnvironmentItemModel
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
					{"secret_simulator_device": "---xxxx---"},
				},
				secretValues: []string{"secret_a_secret_b_secret_c"},
			},
			want: map[string]string{
				"secret_simulator_device": "[REDACTED]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := redactStepInputs(tt.args.environments, tt.args.inputs, tt.args.secretValues)
			require.NoError(t, err, "redactStepInputs()")
			require.Equal(t, got, tt.want)
		})
	}
}
