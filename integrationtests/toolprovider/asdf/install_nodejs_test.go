//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/stretchr/testify/require"
)

func TestAsdfInstallNodeVersion(t *testing.T) {
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
		testEnv, err := createTestEnv(t, asdfInstallation{
			flavor:  flavorAsdfClassic,
			version: "0.14.0",
			plugins: []string{"nodejs"},
		})
		require.NoError(t, err)

		asdfProvider := asdf.AsdfToolProvider{
			ExecEnv: testEnv.toExecEnv(),
		}
		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           "nodejs",
				UnparsedVersion:    tt.requestedVersion,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, err := asdfProvider.InstallTool(request)
			require.NoError(t, err)
			require.Equal(t, provider.ToolID("nodejs"), result.ToolName)
			require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}

func TestCorepackWithNewNodeInstall(t *testing.T) {
	testEnv, err := createTestEnv(t, asdfInstallation{
		flavor:  flavorAsdfClassic,
		version: "0.14.0",
		plugins: []string{"nodejs"},
	})
	require.NoError(t, err)

	asdfProvider := asdf.AsdfToolProvider{
		ExecEnv: testEnv.toExecEnv(),
	}
	request := provider.ToolRequest{
		ToolName:        "nodejs",
		UnparsedVersion: "22.17.0",
	}
	result, err := asdfProvider.InstallTool(request)
	require.NoError(t, err)
	require.Equal(t, provider.ToolID("nodejs"), result.ToolName)
	require.Equal(t, "22.17.0", result.ConcreteVersion)
	require.False(t, result.IsAlreadyInstalled)

	extraEnvs := map[string]string{
		// Simulate the activated environment
		"ASDF_NODEJS_VERSION": "22.17.0",
	}
	out, err := testEnv.runCommand(extraEnvs, "pnpm", "--help")
	require.NoError(t, err)
	require.Contains(t, out, "Usage: pnpm [command] [flags]")
}
