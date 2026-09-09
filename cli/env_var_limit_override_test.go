package cli

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/envman/v2/envman"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/require"
)

func TestApplyEnvVarLimitOverrides(t *testing.T) {
	runner := WorkflowRunner{logger: log.NewLogger(log.GetGlobalLoggerOpts())}
	key := envman.EnvListBytesLimitInKBEnvKey

	tests := []struct {
		name         string
		processEnv   string
		environments []envmanModels.EnvironmentItemModel
		want         string
	}{
		{
			name:         "app env sets the override",
			environments: []envmanModels.EnvironmentItemModel{{key: "1024"}},
			want:         "1024",
		},
		{
			name:       "app env wins over process env",
			processEnv: "512",
			environments: []envmanModels.EnvironmentItemModel{
				{"UNRELATED": "x"},
				{key: "1024"},
			},
			want: "1024",
		},
		{
			name:         "process env is left untouched when no build env sets it",
			processEnv:   "512",
			environments: []envmanModels.EnvironmentItemModel{{"UNRELATED": "x"}},
			want:         "512",
		},
		{
			name:         "invalid build env is skipped, process env stands",
			processEnv:   "512",
			environments: []envmanModels.EnvironmentItemModel{{key: "not-a-number"}},
			want:         "512",
		},
		{
			name:         "negative build env is skipped",
			environments: []envmanModels.EnvironmentItemModel{{key: "-1"}},
			want:         "",
		},
		{
			name:         "empty build env value does not override process env",
			processEnv:   "512",
			environments: []envmanModels.EnvironmentItemModel{{key: ""}},
			want:         "512",
		},
		{
			name: "valid app env overrides valid secret",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "512"},
				{key: "1024"},
			},
			want: "1024",
		},
		{
			name: "invalid app env does not discard a valid secret",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "512"},
				{key: "not-a-number"},
			},
			want: "512",
		},
		{
			name:       "invalid app env does not discard a valid secret, process env set",
			processEnv: "256",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "512"},
				{key: "-1"},
			},
			want: "512",
		},
		{
			name: "invalid secret is skipped in favor of a valid app env",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "not-a-number"},
				{key: "1024"},
			},
			want: "1024",
		},
		{
			name:       "all overrides invalid, process env stands",
			processEnv: "256",
			environments: []envmanModels.EnvironmentItemModel{
				{key: "not-a-number"},
				{key: "-5"},
			},
			want: "256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.processEnv != "" {
				require.NoError(t, os.Setenv(key, tt.processEnv))
			} else {
				require.NoError(t, os.Unsetenv(key))
			}
			t.Cleanup(func() { require.NoError(t, os.Unsetenv(key)) })

			runner.applyEnvVarLimitOverrides(tt.environments)

			require.Equal(t, tt.want, os.Getenv(key))
		})
	}
}

// TestApplyEnvVarLimitOverridesReachesConfigFile proves the resolved limit is
// written to envman's config file, so it reaches an envman binary that reads
// only the file and ignores the process env var overrides.
func TestApplyEnvVarLimitOverridesReachesConfigFile(t *testing.T) {
	runner := WorkflowRunner{logger: log.NewLogger(log.GetGlobalLoggerOpts())}
	key := envman.EnvListBytesLimitInKBEnvKey

	origHome := os.Getenv("HOME")
	t.Cleanup(func() { require.NoError(t, os.Setenv("HOME", origHome)) })
	require.NoError(t, os.Setenv("HOME", t.TempDir()))
	for _, k := range []string{key, envman.EnvBytesLimitInKBEnvKey} {
		require.NoError(t, os.Unsetenv(k))
		t.Cleanup(func() { require.NoError(t, os.Unsetenv(k)) })
	}

	overrides := runner.applyEnvVarLimitOverrides([]envmanModels.EnvironmentItemModel{{key: "1024"}})
	require.NotNil(t, overrides.EnvListBytesLimitInKB)
	require.Equal(t, 1024, *overrides.EnvListBytesLimitInKB)
	require.Nil(t, overrides.EnvBytesLimitInKB)

	restore, err := envman.SetConfigLimits(overrides)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, restore()) })

	// Clear the process env override so GetConfigs can only read the file.
	require.NoError(t, os.Unsetenv(key))

	configs, err := envman.GetConfigs()
	require.NoError(t, err)
	require.Equal(t, 1024, configs.EnvListBytesLimitInKB)

	require.NoError(t, restore())
	configs, err = envman.GetConfigs()
	require.NoError(t, err)
	require.Equal(t, 256, configs.EnvListBytesLimitInKB, "restore should revert the file to the default")
}
