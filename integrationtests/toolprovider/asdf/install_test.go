//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestAsdfInstallClassic(t *testing.T) {
	if useSystemAsdf() {
		t.Skip("Irrelevant test when using system asdf")
	}

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
		UnparsedVersion: "18.16.0",
	}
	result, err := asdfProvider.InstallTool(request)
	require.NoError(t, err)
	require.Equal(t, provider.ToolID("nodejs"), result.ToolName)
	require.Equal(t, "18.16.0", result.ConcreteVersion)
	require.False(t, result.IsAlreadyInstalled)
}

func TestAsdfInstallRewrite(t *testing.T) {
	if useSystemAsdf() {
		t.Skip("Irrelevant test when using system asdf")
	}

	testEnv, err := createTestEnv(t, asdfInstallation{
		flavor:  flavorAsdfRewrite,
		version: "0.18.0",
		plugins: []string{"nodejs"},
	})
	require.NoError(t, err)

	asdfProvider := asdf.AsdfToolProvider{
		ExecEnv: testEnv.toExecEnv(),
	}

	request := provider.ToolRequest{
		ToolName:        "nodejs",
		UnparsedVersion: "18.16.0",
	}
	result, err := asdfProvider.InstallTool(request)
	require.NoError(t, err)
	require.Equal(t, provider.ToolID("nodejs"), result.ToolName)
	require.Equal(t, "18.16.0", result.ConcreteVersion)
	require.False(t, result.IsAlreadyInstalled)
}
