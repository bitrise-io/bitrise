package analytics

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/stretchr/testify/require"
)

func Test_stateChecker_Enabled(t *testing.T) {
	tests := []struct {
		name            string
		disabledEnv     string
		unprefixedEnv   string
		expectedEnabled bool
	}{
		{
			name:            "Neither env var set",
			expectedEnabled: true,
		},
		{
			name:            "BITRISE_ANALYTICS_DISABLED=true disables",
			disabledEnv:     "true",
			expectedEnabled: false,
		},
		{
			name:            "ANALYTICS_DISABLED=true disables",
			unprefixedEnv:   "true",
			expectedEnabled: false,
		},
		{
			name:            "Both set to true still disables",
			disabledEnv:     "true",
			unprefixedEnv:   "true",
			expectedEnabled: false,
		},
		{
			name:            "Non-true values don't disable",
			disabledEnv:     "false",
			unprefixedEnv:   "0",
			expectedEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(DisabledEnvKey, tt.disabledEnv)
			t.Setenv(unprefixedDisabledEnvKey, tt.unprefixedEnv)

			checker := NewStateChecker(env.NewRepository())

			require.Equal(t, tt.expectedEnabled, checker.Enabled())
		})
	}
}
