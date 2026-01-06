//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestNoMatchingVersionError(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	request := provider.ToolRequest{
		ToolName:           provider.ToolID("nodejs"),
		UnparsedVersion:    "0.1.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
	}
	_, err = miseProvider.InstallTool(request)
	require.Error(t, err)

	var installErr provider.ToolInstallError
	require.ErrorAs(t, err, &installErr)
	require.Equal(t, provider.ToolID("nodejs"), installErr.ToolName)
	require.Equal(t, "0.1.0", installErr.RequestedVersion)
	require.Contains(t, installErr.Error(), "failed to install nodejs 0.1.0")
	require.Contains(t, installErr.Cause, "no match for requested version 0.1.0")
}
