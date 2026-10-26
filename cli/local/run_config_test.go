package local

import (
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v3/configs"
	envmanModels "github.com/bitrise-io/bitrise/v3/envman/models"
	"github.com/stretchr/testify/require"
)

func TestReadNoOutputTimoutConfiguration(t *testing.T) {
	key := configs.NoOutputTimeoutEnvKey

	tests := []struct {
		name         string
		processEnv   string
		environments []envmanModels.EnvironmentItemModel
		want         time.Duration
	}{
		{
			name: "disabled when not set",
			want: -1,
		},
		{
			name:       "process env sets the timeout",
			processEnv: "60",
			want:       60 * time.Second,
		},
		{
			name:         "secret wins over process env",
			processEnv:   "60",
			environments: []envmanModels.EnvironmentItemModel{{key: "120"}},
			want:         120 * time.Second,
		},
		{
			name: "app env wins over secret",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "120"},
				{"UNRELATED": "x"},
				{key: "300"},
			},
			want: 300 * time.Second,
		},
		{
			name: "invalid app env does not discard a valid secret",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "120"},
				{key: "$OTHER_VAR"},
			},
			want: 120 * time.Second,
		},
		{
			name:         "empty app env does not override process env",
			processEnv:   "60",
			environments: []envmanModels.EnvironmentItemModel{{key: ""}},
			want:         60 * time.Second,
		},
		{
			name:         "zero app env disables a timeout set by a secret",
			environments: []envmanModels.EnvironmentItemModel{{key: "120"}, {key: "0"}},
			want:         -1,
		},
		{
			name:         "negative value disables",
			environments: []envmanModels.EnvironmentItemModel{{key: "-5"}},
			want:         -1,
		},
		{
			name:       "invalid process env is ignored",
			processEnv: "abc",
			want:       -1,
		},
		{
			name:         "later duplicate secret wins",
			environments: []envmanModels.EnvironmentItemModel{{key: "60"}, {key: "0"}},
			want:         -1,
		},
		{
			name:         "value overflowing time.Duration does not discard a valid process env",
			processEnv:   "60",
			environments: []envmanModels.EnvironmentItemModel{{key: "10000000000"}},
			want:         60 * time.Second,
		},
		{
			name:         "largest representable value is accepted",
			environments: []envmanModels.EnvironmentItemModel{{key: "9223372036"}},
			want:         9223372036 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(key, tt.processEnv)

			require.Equal(t, tt.want, readNoOutputTimoutConfiguration(tt.environments))
		})
	}
}
