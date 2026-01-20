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

func TestToolsCommandHelp(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "tools help shows all subcommands",
			args: []string{"tools", "--help"},
			contains: []string{
				"setup",
				"install",
				"latest",
			},
		},
		{
			name: "install help shows usage",
			args: []string{"tools", "install", "--help"},
			contains: []string{
				"TOOL@VERSION",
				"--provider",
				"--format",
				"node@20.10.0",
			},
		},
		{
			name: "latest help shows usage",
			args: []string{"tools", "latest", "--help"},
			contains: []string{
				"TOOL[@VERSION]",
				"--installed",
				"--provider",
				"--format",
				"node@20",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := command.New(testhelpers.BinPath(), tt.args...).RunAndReturnTrimmedCombinedOutput()

			require.NoError(t, err)
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "output should contain: %s", expected)
			}
		})
	}
}

func TestToolsInstallArgumentValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errContains string
	}{
		{
			name:        "no arguments",
			args:        []string{"tools", "install"},
			expectError: true,
			errContains: "requires exactly 1 argument",
		},
		{
			name:        "too many arguments",
			args:        []string{"tools", "install", "node@20.10.0", "extra"},
			expectError: true,
			errContains: "requires exactly 1 argument",
		},
		{
			name:        "latest no arguments",
			args:        []string{"tools", "latest"},
			expectError: true,
			errContains: "requires exactly 1 argument",
		},
		{
			name:        "latest too many arguments",
			args:        []string{"tools", "latest", "node@20", "extra"},
			expectError: true,
			errContains: "requires exactly 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := command.New(testhelpers.BinPath(), tt.args...).RunAndReturnTrimmedCombinedOutput()

			if tt.expectError {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, output, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestToolsInstallCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		toolSpec       string
		outputFormat   string
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name:         "install exact node version",
			toolSpec:     "node@20.10.0",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "20.10.0")
				assert.Contains(t, output, "Tool setup complete")
			},
		},
		{
			name:         "install with JSON format",
			toolSpec:     "go@1.21.5",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				var result map[string]string
				jsonErr := json.Unmarshal([]byte(output), &result)
				require.NoError(t, jsonErr)
				assert.Contains(t, result, "PATH")
				// Verify the PATH contains go binary
				assert.Contains(t, result["PATH"], "go")
			},
		},
		{
			name:         "install with bash format",
			toolSpec:     "python@3.12.0",
			outputFormat: "bash",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "export PATH=")
				assert.True(t, strings.HasPrefix(output, "export "))
				// Should contain python in the PATH
				assert.Contains(t, output, "python")
			},
		},
		{
			name:     "error on missing version",
			toolSpec: "node",
			validateOutput: func(t *testing.T, output string, err error) {
				require.Error(t, err)
				assert.Contains(t, output, "version cannot be empty")
			},
		},
		{
			name:     "error on invalid tool spec",
			toolSpec: "node@20@10",
			validateOutput: func(t *testing.T, output string, err error) {
				require.Error(t, err)
				assert.Contains(t, output, "invalid tool specification")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"tools", "install", tt.toolSpec}
			if tt.outputFormat != "" {
				args = append(args, "--format", tt.outputFormat)
			}

			cmd := command.New(testhelpers.BinPath(), args...)
			cmd.SetDir(tmpDir)

			output, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if tt.validateOutput != nil {
				tt.validateOutput(t, output, err)
			}
		})
	}
}

func TestToolsLatestCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		toolSpec       string
		installed      bool
		outputFormat   string
		provider       string
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name:     "get latest with version prefix",
			toolSpec: "node@20",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// Plaintext format should just output the version
				assert.True(t, strings.HasPrefix(output, "20."), "version should start with 20., but was: %s", output)
			},
		},
		{
			name:     "get latest without version prefix",
			toolSpec: "python",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// Plaintext format should just output the version
				assert.NotEmpty(t, output, "should return a version")
			},
		},
		{
			name:      "get latest installed version",
			toolSpec:  "ruby",
			installed: true,
			validateOutput: func(t *testing.T, output string, err error) {
				// This may error if no ruby versions are installed
				if err != nil {
					// Accept error for missing installed versions
					assert.True(t,
						strings.Contains(output, "no installed versions") ||
							strings.Contains(output, "not installed") ||
							strings.Contains(output, "failed"),
						"should indicate no installed versions available, but was: %s", output)
				} else {
					// Plaintext format should just output the version
					assert.NotEmpty(t, output, "should return a version")
				}
			},
		},
		{
			name:         "latest with JSON format",
			toolSpec:     "python@3",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				var result map[string]string
				jsonErr := json.Unmarshal([]byte(output), &result)
				require.NoError(t, jsonErr, "output should be valid JSON")
				assert.Contains(t, result, "tool", "JSON should contain tool name")
				assert.Contains(t, result, "version", "JSON should contain version")
				assert.Equal(t, "python", result["tool"])
				assert.True(t, strings.HasPrefix(result["version"], "3."), "version should start with 3., but was: %s", result["version"])
			},
		},
		{
			name:         "latest with plaintext format",
			toolSpec:     "go@1.21",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// In plaintext format, it should just output the version string
				assert.True(t, strings.HasPrefix(output, "1.21"), "plaintext output should be just the version starting with 1.21, but was: %s", output)
			},
		},
		{
			name:         "latest installed with JSON format",
			toolSpec:     "node@20",
			installed:    true,
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				// May error if no node 20.x versions are installed
				if err == nil {
					var result map[string]string
					jsonErr := json.Unmarshal([]byte(output), &result)
					require.NoError(t, jsonErr, "output should be valid JSON")
					assert.Equal(t, "node", result["tool"])
					assert.True(t, strings.HasPrefix(result["version"], "20"), "version should start with 20, but was: %s", result["version"])
				}
			},
		},
		{
			name:         "error on invalid format",
			toolSpec:     "ruby",
			outputFormat: "invalid",
			validateOutput: func(t *testing.T, output string, err error) {
				require.Error(t, err)
				assert.Contains(t, output, "invalid --format")
			},
		},
		{
			name:     "get latest with mise provider",
			toolSpec: "ruby",
			provider: "mise",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				assert.NotEmpty(t, output, "should return a version")
			},
		},
		{
			name:     "get latest with asdf provider",
			toolSpec: "node@20",
			provider: "asdf",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				assert.True(t, strings.HasPrefix(output, "20."), "version should start with 20., but was: %s", output)
			},
		},
		{
			name:         "get latest with asdf provider and JSON format",
			toolSpec:     "python@3",
			provider:     "asdf",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				var result map[string]string
				jsonErr := json.Unmarshal([]byte(output), &result)
				require.NoError(t, jsonErr, "output should be valid JSON")
				assert.Contains(t, result, "tool", "JSON should contain tool name")
				assert.Contains(t, result, "version", "JSON should contain version")
				assert.Equal(t, "python", result["tool"])
				assert.True(t, strings.HasPrefix(result["version"], "3."), "version should start with 3., but was: %s", result["version"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"tools", "latest", tt.toolSpec}
			if tt.installed {
				args = append(args, "--installed")
			}
			if tt.outputFormat != "" {
				args = append(args, "--format", tt.outputFormat)
			}
			if tt.provider != "" {
				args = append(args, "--provider", tt.provider)
			}

			cmd := command.New(testhelpers.BinPath(), args...)
			cmd.SetDir(tmpDir)

			output, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if tt.validateOutput != nil {
				tt.validateOutput(t, output, err)
			}
		})
	}
}
