//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallJavaVersion(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy provider.ResolutionStrategy
		expectedVersion    string
	}{
		{
			name:               "OpenJDK major version only",
			requestedVersion:   "21",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			expectedVersion:    "21.0.2",
		},
		{
			name:               "OpenJDK major version only, latest released",
			requestedVersion:   "17",
			resolutionStrategy: provider.ResolutionStrategyLatestReleased,
			expectedVersion:    "17.0.2",
		},
		{
			name:               "Temurin major version only",
			requestedVersion:   "temurin-22",
			resolutionStrategy: provider.ResolutionStrategyLatestReleased,
			expectedVersion:    "temurin-22.0.2+9",
		},
		{
			name:               "Temurin exact version",
			requestedVersion:   "temurin-18.0.2+9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			expectedVersion:    "temurin-18.0.2+9",
		},
	}

	for _, tt := range tests {
		miseInstallDir := t.TempDir()
		miseDataDir := t.TempDir()
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
		require.NoError(t, err)

		err = miseProvider.Bootstrap()
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           provider.ToolID("java"),
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, err := miseProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, provider.ToolID("java"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
