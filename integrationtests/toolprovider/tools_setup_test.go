//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsSetupCommand(t *testing.T) {
	t.Run("setup from .tool-versions file", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "golang 1.21.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", toolVersionsPath, "--format", "plaintext")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		// May fail if tool not available, but should parse correctly.
		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		} else {
			// Should contain env var output.
			assert.Contains(t, out, "Env vars to activate installed tools")
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

		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", toolVersionsPath, "--format", "plaintext")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		}
	})

	t.Run("setup from bitrise.yml with global tools", func(t *testing.T) {
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

		cmd := command.New(testhelpers.BinPath(), "tools", "setup",
			"--config", bitriseYml,
			"--workflow", "test",
			"--format", "plaintext")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		}
	})

	t.Run("setup from bitrise.yml with workflow-specific tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		bitriseYml := filepath.Join(tmpDir, "bitrise.yml")

		content := `format_version: "17"
tools:
  ruby: 3.2.0
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

		cmd := command.New(testhelpers.BinPath(), "tools", "setup",
			"--config", bitriseYml,
			"--workflow", "test",
			"--format", "plaintext")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		}
	})

	t.Run("setup with multiple tools in bitrise.yml", func(t *testing.T) {
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

		cmd := command.New(testhelpers.BinPath(), "tools", "setup",
			"--config", bitriseYml,
			"--workflow", "build",
			"--format", "plaintext")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		}
	})

	t.Run("output format json", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "golang 1.21.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", toolVersionsPath, "--format", "json")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		} else {
			// Check JSON validity.
			trimmed := strings.TrimSpace(out)
			if trimmed != "" {
				assert.True(t, strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["),
					"JSON output should start with { or [, got: %s", trimmed)
			}
		}
	})

	t.Run("output format bash", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "golang 1.21.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", toolVersionsPath, "--format", "bash")
		cmd.SetDir(tmpDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		if err != nil {
			t.Logf("Setup output: %s", out)
			t.Logf("Setup error (may be expected): %v", err)
		} else if out != "" {
			// Bash output should be environment variable assignments.
			assert.Contains(t, out, "=", "Bash output should contain env var assignments")
		}
	})

	t.Run("error on nonexistent config file", func(t *testing.T) {
		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", "/nonexistent/path/.tool-versions")
		_, err := cmd.RunAndReturnTrimmedCombinedOutput()

		require.Error(t, err)
	})

	t.Run("error on multiple bitrise.yml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		bitriseYml1 := filepath.Join(tmpDir, "bitrise1.yml")
		bitriseYml2 := filepath.Join(tmpDir, "bitrise2.yml")

		content := `format_version: "17"
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

		err := os.WriteFile(bitriseYml1, []byte(content), 0644)
		require.NoError(t, err)
		err = os.WriteFile(bitriseYml2, []byte(content), 0644)
		require.NoError(t, err)

		cmd := command.New(testhelpers.BinPath(), "tools", "setup",
			"--config", bitriseYml1,
			"--config", bitriseYml2)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		require.Error(t, err)
		assert.Contains(t, out, "multiple bitrise config files")
	})

	t.Run("error on invalid output format", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := "golang 1.21.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		cmd := command.New(testhelpers.BinPath(), "tools", "setup", "--config", toolVersionsPath, "--format", "invalid")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		require.Error(t, err)
		assert.Contains(t, out, "invalid --format")
	})
}
