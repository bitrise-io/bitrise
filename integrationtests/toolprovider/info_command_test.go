//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsInfoCommandHelp(t *testing.T) {
	output, err := command.New(testhelpers.BinPath(), "tools", "info", "--help").RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err)
	assert.Contains(t, output, "--active")
	assert.Contains(t, output, "--format")
	assert.Contains(t, output, "Show information about installed or active tools")
}

func TestToolsInfoCommandHelpInToolsList(t *testing.T) {
	output, err := command.New(testhelpers.BinPath(), "tools", "--help").RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err)
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "setup")
	assert.Contains(t, output, "install")
	assert.Contains(t, output, "latest")
}

func TestToolsInfoCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name: "default plaintext format",
			args: []string{"tools", "info"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				// Should either show installed tools or "No tools installed" message.
				assert.True(t,
					strings.Contains(output, "Installed tools:") || strings.Contains(output, "No tools installed"),
					"output should contain tools list or no tools message",
				)
			},
		},
		{
			name: "JSON format",
			args: []string{"tools", "info", "--format", "json"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				// Should be valid JSON (either empty array or array of tools).
				var tools []map[string]interface{}
				jsonErr := json.Unmarshal([]byte(output), &tools)
				require.NoError(t, jsonErr, "output should be valid JSON: %s", output)
			},
		},
		{
			name: "active tools plaintext",
			args: []string{"tools", "info", "--active"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				// Should either show active tools or "No active tools" message.
				assert.True(t,
					strings.Contains(output, "Active tools:") || strings.Contains(output, "No active tools"),
					"output should contain active tools list or no active tools message",
				)
			},
		},
		{
			name: "active tools JSON format",
			args: []string{"tools", "info", "--active", "--format", "json"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				var tools []map[string]interface{}
				jsonErr := json.Unmarshal([]byte(output), &tools)
				require.NoError(t, jsonErr, "output should be valid JSON: %s", output)

				// If there are active tools, they should have active_version and source.
				for _, tool := range tools {
					assert.Contains(t, tool, "name", "each tool should have a name")
					assert.Contains(t, tool, "active_version", "active tools should have active_version")
				}
			},
		},
		{
			name: "short flag for active",
			args: []string{"tools", "info", "-a"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.True(t,
					strings.Contains(output, "Active tools:") || strings.Contains(output, "No active tools"),
					"short flag -a should work the same as --active",
				)
			},
		},
		{
			name: "short flag for format",
			args: []string{"tools", "info", "-f", "json"},
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				var tools []map[string]interface{}
				jsonErr := json.Unmarshal([]byte(output), &tools)
				require.NoError(t, jsonErr, "short flag -f should work the same as --format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := command.New(testhelpers.BinPath(), tt.args...)
			output, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if tt.validateOutput != nil {
				tt.validateOutput(t, output, err)
			}
		})
	}
}

func TestToolsInfoJSONStructure(t *testing.T) {
	t.Run("installed tools JSON structure", func(t *testing.T) {
		cmd := command.New(testhelpers.BinPath(), "tools", "info", "--format", "json")
		output, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)

		var tools []struct {
			Name              string   `json:"name"`
			InstalledVersions []string `json:"installed_versions,omitempty"`
			ActiveVersion     string   `json:"active_version,omitempty"`
			Source            string   `json:"source,omitempty"`
		}
		jsonErr := json.Unmarshal([]byte(output), &tools)
		require.NoError(t, jsonErr, "output should be valid JSON matching InstalledTool struct")

		for _, tool := range tools {
			assert.NotEmpty(t, tool.Name, "tool should have a name")
		}
	})

	t.Run("active tools JSON structure", func(t *testing.T) {
		cmd := command.New(testhelpers.BinPath(), "tools", "info", "--active", "--format", "json")
		output, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)

		var tools []struct {
			Name              string   `json:"name"`
			InstalledVersions []string `json:"installed_versions,omitempty"`
			ActiveVersion     string   `json:"active_version,omitempty"`
			Source            string   `json:"source,omitempty"`
		}
		jsonErr := json.Unmarshal([]byte(output), &tools)
		require.NoError(t, jsonErr, "output should be valid JSON matching InstalledTool struct")

		// Active tools should have active_version set.
		for _, tool := range tools {
			assert.NotEmpty(t, tool.Name, "tool should have a name")
			assert.NotEmpty(t, tool.ActiveVersion, "active tool should have active_version")
		}
	})
}

func TestToolsInfoAfterInstall(t *testing.T) {
	tmpDir := t.TempDir()

	installCmd := command.New(testhelpers.BinPath(), "tools", "install", "node@20.10.0", "--format", "json")
	installCmd.SetDir(tmpDir)
	_, err := installCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		t.Skip("Skipping test: could not install node@20.10.0")
	}

	t.Run("installed tool appears in info", func(t *testing.T) {
		infoCmd := command.New(testhelpers.BinPath(), "tools", "info", "--format", "json")
		infoCmd.SetDir(tmpDir)
		output, err := infoCmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)

		var tools []struct {
			Name              string   `json:"name"`
			InstalledVersions []string `json:"installed_versions,omitempty"`
			ActiveVersion     string   `json:"active_version,omitempty"`
		}
		jsonErr := json.Unmarshal([]byte(output), &tools)
		require.NoError(t, jsonErr)

		// Look for node in the results.
		foundNode := false
		for _, tool := range tools {
			if tool.Name == "node" {
				foundNode = true
				// Version 20.10.0 should be in installed versions.
				found := false
				for _, v := range tool.InstalledVersions {
					if v == "20.10.0" {
						found = true
						break
					}
				}
				assert.True(t, found, "node 20.10.0 should be in installed versions")
				break
			}
		}
		assert.True(t, foundNode, "node should appear in tools info after installation")
	})
}
