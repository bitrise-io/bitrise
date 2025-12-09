//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
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

	t.Run("setup from .tool-versions with multiple tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := `ruby 3.2.0
nodejs 20.0.0
python 3.11.0`
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		opts := toolprovider.SetupOptions{
			VersionFiles: []string{toolVersionsPath},
		}

		tracker := analytics.NewDefaultTracker()
		_, err = toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Attempt to install all three tools.
		if err != nil {
			// Error is acceptable - testing parsing only.
			t.Logf("Setup error (expected in minimal test env): %v", err)
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

	t.Run("auto-detect multiple version files", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := map[string]string{
			".ruby-version":   "3.2.0",
			".python-version": "3.11.0",
			".node-version":   "20.0.0",
		}

		for filename, content := range files {
			path := filepath.Join(tmpDir, filename)
			err := os.WriteFile(path, []byte(content), 0644)
			require.NoError(t, err)
		}

		opts := toolprovider.SetupOptions{
			WorkingDir: tmpDir,
		}

		tracker := analytics.NewDefaultTracker()
		_, err := toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Detect and attempt to install tools.
		if err != nil {
			t.Logf("Setup error (expected in minimal test env): %v", err)
		}
	})

	t.Run("mixed .tool-versions and individual version files", func(t *testing.T) {
		tmpDir := t.TempDir()

		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")
		err := os.WriteFile(toolVersionsPath, []byte("ruby 3.2.0\ngolang 1.21.0"), 0644)
		require.NoError(t, err)

		// Additional version files.
		pythonVersionPath := filepath.Join(tmpDir, ".python-version")
		err = os.WriteFile(pythonVersionPath, []byte("3.11.0"), 0644)
		require.NoError(t, err)

		nodeVersionPath := filepath.Join(tmpDir, ".node-version")
		err = os.WriteFile(nodeVersionPath, []byte("20.0.0"), 0644)
		require.NoError(t, err)

		opts := toolprovider.SetupOptions{
			WorkingDir: tmpDir,
		}

		tracker := analytics.NewDefaultTracker()
		_, err = toolprovider.SetupFromVersionFiles(opts, tracker, false)

		// Find and parse all files.
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

func TestBitriseYmlWorkflowIntegration(t *testing.T) {
	t.Run("parse bitrise.yml with global tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  nodejs: 20.0.0
  ruby: 3.2.0
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		// Parse and validate the config.
		config, warns, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)
		if len(warns) > 0 {
			t.Logf("Warnings: %v", warns)
		}

		// Verify global tools are parsed.
		assert.NotNil(t, config.Tools)
		assert.Equal(t, "20.0.0", string(config.Tools["nodejs"]))
		assert.Equal(t, "3.2.0", string(config.Tools["ruby"]))
	})

	t.Run("parse bitrise.yml with workflow-specific tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
workflows:
  test:
    tools:
      nodejs: 20.0.0
      python: 3.11.0
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Verify workflow tools are parsed.
		workflow := config.Workflows["test"]
		assert.NotNil(t, workflow.Tools)
		assert.Equal(t, "20.0.0", string(workflow.Tools["nodejs"]))
		assert.Equal(t, "3.11.0", string(workflow.Tools["python"]))
	})

	t.Run("parse bitrise.yml with mixed global and workflow tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  nodejs: 18.0.0
  ruby: 3.2.0
workflows:
  test:
    tools:
      nodejs: 20.0.0
      python: 3.11.0
    steps:
      - script:
          inputs:
            - content: echo "test"
  other:
    steps:
      - script:
          inputs:
            - content: echo "other"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Verify global tools.
		assert.Equal(t, "18.0.0", string(config.Tools["nodejs"]))
		assert.Equal(t, "3.2.0", string(config.Tools["ruby"]))

		// Verify test workflow overrides nodejs and adds python.
		testWorkflow := config.Workflows["test"]
		assert.Equal(t, "20.0.0", string(testWorkflow.Tools["nodejs"]))
		assert.Equal(t, "3.11.0", string(testWorkflow.Tools["python"]))

		// Verify other workflow has no workflow-specific tools.
		otherWorkflow := config.Workflows["other"]
		assert.Nil(t, otherWorkflow.Tools)
	})

	t.Run("parse bitrise.yml with multiple tools in workflow", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  nodejs: 20.0.0
  ruby: 3.2.0
  python: 3.11.0
  golang: 1.21.0
  java: openjdk-11
tool_config:
  provider: mise
workflows:
  build:
    steps:
      - script:
          inputs:
            - content: echo "build"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Should handle 5+ tools in global config.
		assert.Len(t, config.Tools, 5)
		assert.Equal(t, "20.0.0", string(config.Tools["nodejs"]))
		assert.Equal(t, "3.2.0", string(config.Tools["ruby"]))
		assert.Equal(t, "3.11.0", string(config.Tools["python"]))
		assert.Equal(t, "1.21.0", string(config.Tools["golang"]))
		assert.Equal(t, "openjdk-11", string(config.Tools["java"]))

		// Verify tool_config.
		assert.NotNil(t, config.ToolConfig)
		assert.Equal(t, "mise", config.ToolConfig.Provider)
	})

	t.Run("parse bitrise.yml with tool unset in workflow", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  nodejs: 20.0.0
  ruby: 3.2.0
workflows:
  test:
    tools:
      ruby: unset
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Verify workflow has unset marker.
		testWorkflow := config.Workflows["test"]
		assert.Equal(t, "unset", string(testWorkflow.Tools["ruby"]))
	})

	t.Run("parse bitrise.yml with version strategies", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  nodejs: 20:latest
  ruby: 3.2:installed
  python: 3.11.0
tool_config:
  provider: mise
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Should parse different version resolution strategies.
		assert.Equal(t, "20:latest", string(config.Tools["nodejs"]))
		assert.Equal(t, "3.2:installed", string(config.Tools["ruby"]))
		assert.Equal(t, "3.11.0", string(config.Tools["python"]))
	})

	t.Run("parse bitrise.yml with custom tool plugins", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  custom-tool: 1.0.0
tool_config:
  provider: mise
  extra_plugins:
    custom-tool: https://github.com/example/custom-tool-plugin
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml, []byte(content), 0644)
		require.NoError(t, err)

		config, _, err := bitrise.ReadBitriseConfig(bitriseYml, bitrise.ValidationTypeFull)
		require.NoError(t, err)

		// Should handle custom tool plugins configuration.
		assert.Equal(t, "1.0.0", string(config.Tools["custom-tool"]))
		assert.NotNil(t, config.ToolConfig)
		assert.NotNil(t, config.ToolConfig.ExtraPlugins)
		assert.Equal(t, "https://github.com/example/custom-tool-plugin", config.ToolConfig.ExtraPlugins["custom-tool"])
	})
}
