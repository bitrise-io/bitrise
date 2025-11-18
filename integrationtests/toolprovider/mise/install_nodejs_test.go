//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallNodeVersion(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy provider.ResolutionStrategy
		expectedVersion    string
	}{
		{"Install specific version", "18.16.0", provider.ResolutionStrategyStrict, "18.16.0"},
		{"Install partial major version", "18", provider.ResolutionStrategyLatestInstalled, "18.20.8"},
		{"Install partial major.minor version", "18.10", provider.ResolutionStrategyLatestReleased, "18.10.0"},
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
				ToolName:           provider.ToolID("nodejs"),
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, err := miseProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, provider.ToolID("nodejs"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
