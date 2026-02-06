package bitrise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/stretchr/testify/require"
)

func Test_validateInstalledPlugins(t *testing.T) {
	t.Run("validates existing plugins successfully", func(t *testing.T) {
		err := plugins.InitPaths()
		require.NoError(t, err)

		err = validateInstalledPlugins()
		// Should not return an error even if no plugins are installed.
		require.NoError(t, err)
	})

	t.Run("handles empty routing file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "bitrise-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Force init paths to use tmp directory.
		plugins.ForceInitPaths(tmpDir)
		defer func() {
			if err := plugins.InitPaths(); err != nil {
				t.Logf("Failed to re-initialize plugin paths: %v", err)
			}
		}()

		pluginsDir := filepath.Join(tmpDir, "plugins")
		err = os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		// Should handle empty routing gracefully.
		err = validateInstalledPlugins()
		require.NoError(t, err)
	})

	t.Run("cleans up broken plugin directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "bitrise-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Force init paths to use tmp directory.
		plugins.ForceInitPaths(tmpDir)
		defer func() {
			if err := plugins.InitPaths(); err != nil {
				t.Logf("Failed to re-initialize plugin paths: %v", err)
			}
		}()

		pluginsDir := filepath.Join(tmpDir, "plugins")
		err = os.MkdirAll(pluginsDir, 0755)
		require.NoError(t, err)

		// Broken plugin directory (without proper definition file).
		brokenPluginDir := filepath.Join(pluginsDir, "broken-plugin", "src")
		err = os.MkdirAll(brokenPluginDir, 0755)
		require.NoError(t, err)

		// Routing file referencing the broken plugin.
		routingContent := `route_map:
  broken-plugin:
    name: broken-plugin
    source: "local"
    version: "1.0.0"
`
		routingFile := filepath.Join(pluginsDir, "spec.yml")
		err = os.WriteFile(routingFile, []byte(routingContent), 0644)
		require.NoError(t, err)

		// Should clean up the broken plugin.
		err = validateInstalledPlugins()
		require.NoError(t, err)

		// Verify the broken plugin directory was removed.
		_, err = os.Stat(filepath.Join(pluginsDir, "broken-plugin"))
		require.True(t, os.IsNotExist(err), "Broken plugin directory should be removed")

		// Verify the routing file no longer contains the broken plugin.
		routing, err := plugins.ReadPluginRouting()
		require.NoError(t, err)
		_, exists := routing.RouteMap["broken-plugin"]
		require.False(t, exists, "Broken plugin should be removed from routing")
	})
}
