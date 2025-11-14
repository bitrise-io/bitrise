//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallNixpkgsRuby(t *testing.T) {
	tests := []struct {
		name               string
		version            string
		resolutionStrategy provider.ResolutionStrategy
		want               string
	}{
		{
			name:               "Install specific version",
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			want:               "3.3.9",
		},
	}

	t.Setenv("BITRISE_TOOLSETUP_FAST_INSTALL", "1")

	for _, tt := range tests {
		miseInstallDir := t.TempDir()
		miseDataDir := t.TempDir()
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir)
		require.NoError(t, err)

		err = miseProvider.Bootstrap()
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    tt.version,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, installErr := miseProvider.InstallTool(request)

			require.NoError(t, installErr)
			require.Equal(t, provider.ToolID("ruby"), result.ToolName)
			require.Equal(t, tt.want, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
