//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/stretchr/testify/require"
)

func TestAsdfInstallGolangVersion(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy provider.ResolutionStrategy
		expectedVersion    string
	}{
		{"Install specific version", "1.23.4", provider.ResolutionStrategyStrict, "1.23.4"},
		{"Install partial major.minor version", "1.22", provider.ResolutionStrategyLatestInstalled, "1.22.12"},
		{"Install partial major.minor version, latest released", "1.22", provider.ResolutionStrategyLatestReleased, "1.22.12"},
	}

	for _, tt := range tests {
		testEnv, err := createTestEnv(t, asdfInstallation{
			flavor:  flavorAsdfClassic,
			version: "0.14.0",
			plugins: []string{"golang"},
		})
		require.NoError(t, err)

		asdfProvider := asdf.AsdfToolProvider{
			ExecEnv: testEnv.toExecEnv(),
		}
		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           "golang",
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, err := asdfProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, provider.ToolID("golang"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
