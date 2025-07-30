//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/stretchr/testify/require"
)

func TestAsdfInstallFlutterNoPlugin(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy toolprovider.ResolutionStrategy
		plugin             string
		expectedVersion    string
	}{
		{"Install specific version", "3.32.5-stable", toolprovider.ResolutionStrategyStrict, "flutter::https://github.com/asdf-community/asdf-flutter.git", "3.32.5-stable"},
		{"Install specific version", "3.32.1-stable", toolprovider.ResolutionStrategyStrict, "", "3.32.1-stable"},
	}

	for _, tt := range tests {
		testEnv, err := createTestEnv(t, asdfInstallation{
			flavor:  flavorAsdfClassic,
			version: "0.14.0",
		})
		require.NoError(t, err)

		asdfProvider := asdf.AsdfToolProvider{
			ExecEnv: testEnv.toExecEnv(),
		}
		t.Run(tt.name, func(t *testing.T) {
			request := toolprovider.ToolRequest{
				ToolName:           "flutter",
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
				PluginIdentifier:   &tt.plugin,
			}
			result, err := asdfProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, toolprovider.ToolID("flutter"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
