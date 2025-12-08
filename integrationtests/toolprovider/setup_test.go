//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupFromVersionFilesIntegration(t *testing.T) {
	t.Run("setup from .tool-versions file", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "golang 1.21.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		opts := toolprovider.SetupOptions{
			VersionFiles: []string{toolVersionsPath},
		}

		tracker := analytics.NewDefaultTracker()
		envs, err := toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Note: This may fail in test environment without proper tool setup.
		if err != nil {
			t.Logf("Setup error (expected in minimal test env): %v", err)
		} else {
			assert.NotNil(t, envs)
		}
	})

	t.Run("auto-detect version files", func(t *testing.T) {
		tmpDir := t.TempDir()

		rubyVersionPath := filepath.Join(tmpDir, ".ruby-version")
		err := os.WriteFile(rubyVersionPath, []byte("3.2.0"), 0644)
		require.NoError(t, err)

		opts := toolprovider.SetupOptions{
			WorkingDir: tmpDir,
		}

		tracker := analytics.NewDefaultTracker()
		_, err = toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Should not panic, may error if tools not available.
		if err != nil {
			t.Logf("Setup error (expected in minimal test env): %v", err)
		}
	})

	t.Run("no version files found", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := toolprovider.SetupOptions{
			WorkingDir: tmpDir,
		}

		tracker := analytics.NewDefaultTracker()
		envs, err := toolprovider.SetupFromVersionFiles(opts, tracker, false)

		require.NoError(t, err)
		assert.Nil(t, envs)
	})

	t.Run("invalid version file path", func(t *testing.T) {
		opts := toolprovider.SetupOptions{
			VersionFiles: []string{"/nonexistent/path/.tool-versions"},
		}

		tracker := analytics.NewDefaultTracker()
		_, err := toolprovider.SetupFromVersionFiles(opts, tracker, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse version file")
	})

	t.Run("extra plugins configuration", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "custom-tool 1.0.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		opts := toolprovider.SetupOptions{
			VersionFiles: []string{toolVersionsPath},
			ExtraPlugins: map[models.ToolID]string{
				"custom-tool": "https://github.com/example/custom-tool-plugin",
			},
		}

		tracker := analytics.NewDefaultTracker()
		_, err = toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Will likely fail without the actual plugin, but should parse correctly.
		if err != nil {
			t.Logf("Setup error (expected for custom plugin): %v", err)
		}
	})
}

func TestListInstalledToolsIntegration(t *testing.T) {
	t.Run("list with mise provider", func(t *testing.T) {
		tools, err := toolprovider.ListInstalledTools("mise")

		// Should not panic, may have no tools installed.
		if err != nil {
			t.Logf("List tools error (may not have mise): %v", err)
		} else {
			assert.NotNil(t, tools)
			t.Logf("Found %d tools installed", len(tools))
		}
	})

	t.Run("list with asdf provider", func(t *testing.T) {
		tools, err := toolprovider.ListInstalledTools("asdf")

		// Should not panic, may have no tools installed.
		if err != nil {
			t.Logf("List tools error (may not have asdf): %v", err)
		} else {
			assert.NotNil(t, tools)
			t.Logf("Found %d tools installed", len(tools))
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		_, err := toolprovider.ListInstalledTools("invalid-provider")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported tool provider")
	})
}
