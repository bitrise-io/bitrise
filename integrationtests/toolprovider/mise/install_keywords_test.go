//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

// TestMiseInstallWithLatestKeyword tests that the "latest" keyword works for various tools
// without failing. We don't assert specific versions since "latest" changes over time.
func TestMiseInstallWithLatestKeyword(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	tests := []struct {
		name     string
		toolName provider.ToolID
	}{
		{"Install latest golang", "golang"},
		{"Install latest python", "python"},
		{"Install latest node", "node"},
		{"Install latest tuist", "tuist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    "latest",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			}
			result, err := miseProvider.InstallTool(request)
			require.NoError(t, err, "Installing %s with 'latest' keyword should not fail", tt.toolName)
			require.Equal(t, tt.toolName, result.ToolName)
			require.NotEmpty(t, result.ConcreteVersion, "Concrete version should be resolved for %s", tt.toolName)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}

// TestMiseInstallWithInstalledKeyword tests that the "installed" keyword works correctly.
// It first installs a specific version, then requests "installed" which should find it.
func TestMiseInstallWithInstalledKeyword(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	tests := []struct {
		name             string
		toolName         provider.ToolID
		versionToInstall string
	}{
		{"Tuist installed keyword", "tuist", "4.38"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, install a specific version
			installRequest := provider.ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    tt.versionToInstall,
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			}
			installResult, err := miseProvider.InstallTool(installRequest)
			require.NoError(t, err, "Installing %s version %s should not fail", tt.toolName, tt.versionToInstall)
			require.NotEmpty(t, installResult.ConcreteVersion)

			// Now request "installed" keyword - should find the previously installed version
			installedRequest := provider.ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    "installed",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			}
			installedResult, err := miseProvider.InstallTool(installedRequest)
			require.NoError(t, err, "Installing %s with 'installed' keyword should not fail", tt.toolName)
			require.Equal(t, tt.toolName, installedResult.ToolName)
			require.NotEmpty(t, installedResult.ConcreteVersion, "Concrete version should be resolved for %s", tt.toolName)
			require.True(t, installedResult.IsAlreadyInstalled, "Tool should be marked as already installed")
		})
	}
}

// TestMiseInstallWithLatestAfterInstalled tests a combination:
// Install a version, then install "latest" to ensure both keywords work together.
func TestMiseInstallWithLatestAfterInstalled(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Install a specific tuist version
	specificRequest := provider.ToolRequest{
		ToolName:           "tuist",
		UnparsedVersion:    "4.38",
		ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
	}
	specificResult, err := miseProvider.InstallTool(specificRequest)
	require.NoError(t, err)
	require.NotEmpty(t, specificResult.ConcreteVersion)

	// Now install "latest" - should work without issues
	latestRequest := provider.ToolRequest{
		ToolName:           "tuist",
		UnparsedVersion:    "latest",
		ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
	}
	latestResult, err := miseProvider.InstallTool(latestRequest)
	require.NoError(t, err, "Installing 'latest' after a specific version should not fail")
	require.NotEmpty(t, latestResult.ConcreteVersion)

	// Verify that "installed" keyword still works
	installedRequest := provider.ToolRequest{
		ToolName:           "tuist",
		UnparsedVersion:    "installed",
		ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
	}
	installedResult, err := miseProvider.InstallTool(installedRequest)
	require.NoError(t, err, "Using 'installed' keyword should not fail")
	require.NotEmpty(t, installedResult.ConcreteVersion)
	require.True(t, installedResult.IsAlreadyInstalled)
}
