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
				assert.Contains(t, output, "version required")
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
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name:     "install latest with version prefix",
			toolSpec: "node@20",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "20")
				assert.Contains(t, output, "Tool setup complete")
			},
		},
		{
			name:     "install latest without version prefix",
			toolSpec: "python",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "python")
				assert.Contains(t, output, "Tool setup complete")
			},
		},
		{
			name:      "install latest installed version",
			toolSpec:  "node@20",
			installed: true,
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "20")
				assert.Contains(t, output, "latest installed")
			},
		},
		{
			name:         "latest with JSON format",
			toolSpec:     "go@1",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				var result map[string]string
				jsonErr := json.Unmarshal([]byte(output), &result)
				require.NoError(t, jsonErr)
				assert.Contains(t, result, "PATH")
			},
		},
		{
			name:         "latest with bash format",
			toolSpec:     "go@1.21",
			outputFormat: "bash",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "export PATH=")
				assert.Contains(t, output, "export GOROOT=")
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

			cmd := command.New(testhelpers.BinPath(), args...)
			cmd.SetDir(tmpDir)

			output, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if tt.validateOutput != nil {
				tt.validateOutput(t, output, err)
			}
		})
	}
}
