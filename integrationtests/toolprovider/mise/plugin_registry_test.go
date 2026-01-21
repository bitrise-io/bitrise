//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginRegistry_KnownTool(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Test with a tool in registry
	request := provider.ToolRequest{
		ToolName:           provider.ToolID("flutter"),
		UnparsedVersion:    "3.32.1-stable",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
		PluginURL:          nil,
	}

	err = miseProvider.InstallPlugin(request)
	assert.NoError(t, err, "InstallPlugin should succeed for a tool in the registry")
}

func TestPluginRegistry_CoreTool(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Test with a core tool
	request := provider.ToolRequest{
		ToolName:           provider.ToolID("nodejs"),
		UnparsedVersion:    "18.0.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
		PluginURL:          nil,
	}

	err = miseProvider.InstallPlugin(request)
	assert.NoError(t, err, "InstallPlugin should succeed for core tools without plugin installation")
}

func TestPluginRegistry_UnknownTool(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Test with a tool that's unlikely to be in the registry
	request := provider.ToolRequest{
		ToolName:           provider.ToolID("nonexistent-tool-xyz123"),
		UnparsedVersion:    "1.0.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
		PluginURL:          nil, // No custom plugin URL to force registry lookup
	}

	// This should fail with a ToolInstallError because the tool is not in the registry
	err = miseProvider.InstallPlugin(request)
	require.Error(t, err, "InstallPlugin should fail for unknown tools not in registry")

	var installErr provider.ToolInstallError
	require.ErrorAs(t, err, &installErr, "Error should be a ToolInstallError")
	assert.Equal(t, provider.ToolID("nonexistent-tool-xyz123"), installErr.ToolName)
	assert.Equal(t, "1.0.0", installErr.RequestedVersion)
	assert.Contains(t, installErr.Cause, "This tool integration (nonexistent-tool-xyz123) is not tested or vetted by Bitrise")
	assert.Contains(t, installErr.Recommendation, "look up its asdf plugin and set its git clone URL")
}

func TestPluginRegistry_CustomPluginURL(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Test with a custom plugin URL - should bypass registry check
	customPluginURL := "https://github.com/asdf-community/asdf-flutter.git"
	request := provider.ToolRequest{
		ToolName:           provider.ToolID("flutter"),
		UnparsedVersion:    "3.32.1-stable",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
		PluginURL:          &customPluginURL,
	}

	err = miseProvider.InstallPlugin(request)
	assert.NoError(t, err, "InstallPlugin should succeed with a custom plugin URL")
}

func TestPluginRegistry_EmptyToolName(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Test with empty tool name
	request := provider.ToolRequest{
		ToolName:           provider.ToolID(""),
		UnparsedVersion:    "1.0.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
		PluginURL:          nil,
	}

	err = miseProvider.InstallPlugin(request)
	require.Error(t, err, "InstallPlugin should fail for empty tool name")
	assert.Contains(t, err.Error(), "name is not defined")
}
