//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallFlutter(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy provider.ResolutionStrategy
		expectedVersion    string
	}{
		{"Install specific version", "3.32.1", provider.ResolutionStrategyStrict, "3.32.1"},
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
				ToolName:           "flutter",
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, err := miseProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, provider.ToolID("flutter"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
