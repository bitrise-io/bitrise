//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestNoMatchingVersionError(t *testing.T) {
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
		ToolName:           "nodejs",
		UnparsedVersion:    "22",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
	}
	_, err = asdfProvider.InstallTool(request)
	require.Error(t, err)

	var installErr provider.ToolInstallError
	require.ErrorAs(t, err, &installErr)
	require.Equal(t, provider.ToolID("nodejs"), installErr.ToolName)
	require.Equal(t, "22", installErr.RequestedVersion)
	require.Contains(t, installErr.Error(), "no match for requested version 22")
	require.Contains(t, installErr.Recommendation, "22:latest")
	require.Contains(t, installErr.Recommendation, "22:installed")
}

func TestNewToolPluginError(t *testing.T) {
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
		ToolName:           "foo",
		UnparsedVersion:    "1.0.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
	}
	_, err = asdfProvider.InstallTool(request)
	require.Error(t, err)

	var installErr provider.ToolInstallError
	require.ErrorAs(t, err, &installErr)
	require.Equal(t, provider.ToolID("foo"), installErr.ToolName)
	require.Equal(t, "1.0.0", installErr.RequestedVersion)
	require.Equal(t, installErr.Cause, "This tool integration (foo) is not tested or vetted by Bitrise.")
	require.Equal(t, installErr.Recommendation, "If you want to use this tool anyway, look up its asdf plugin and set its git clone URL in tool_config.extra_plugins. For example: `foo: https://github/url/to/asdf/plugin/repo.git`")
}
