package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallFromAlternateBackends(t *testing.T) {
	tests := []struct {
		name             string
		toolName         provider.ToolID
		strategy         provider.ResolutionStrategy
		requestedVersion string
		expectedVersion  string
	}{
		{
			name:             "golangci-lint from ubi",
			strategy:         provider.ResolutionStrategyStrict,
			toolName:         "ubi:golangci/golangci-lint",
			requestedVersion: "2.2.2",
			expectedVersion:  "2.2.2",
		},
		{
			name:             "hadolint from aqua",
			strategy:         provider.ResolutionStrategyLatestReleased,
			toolName:         "aqua:hadolint/hadolint",
			requestedVersion: "2.11",
			expectedVersion:  "2.11.0",
		},
	}

	for _, tt := range tests {
		miseInstallDir := t.TempDir()
		miseDataDir := t.TempDir()
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir)
		require.NoError(t, err)

		err = miseProvider.Bootstrap()
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.strategy,
			}
			result, err := miseProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, tt.toolName, result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
