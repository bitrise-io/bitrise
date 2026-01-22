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
				"TOOL VERSION[:SUFFIX]",
				"--provider",
				"--format",
				"nodejs 20.10.0",
			},
		},
		{
			name: "latest help shows usage",
			args: []string{"tools", "latest", "--help"},
			contains: []string{
				"TOOL [VERSION[:SUFFIX]]",
				"--provider",
				"--format",
				"nodejs 20",
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
			errContains: "requires 2 arguments",
		},
		{
			name:        "only tool name (missing version)",
			args:        []string{"tools", "install", "nodejs"},
			expectError: true,
			errContains: "requires 2 arguments",
		},
		{
			name:        "too many arguments",
			args:        []string{"tools", "install", "nodejs", "20.10.0", "extra"},
			expectError: true,
			errContains: "requires 2 arguments",
		},
		{
			name:        "latest no arguments",
			args:        []string{"tools", "latest"},
			expectError: true,
			errContains: "requires 1 or 2 arguments",
		},
		{
			name:        "latest too many arguments",
			args:        []string{"tools", "latest", "nodejs", "20", "extra"},
			expectError: true,
			errContains: "requires 1 or 2 arguments",
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
		toolName       string
		toolVersion    string
		outputFormat   string
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name:         "install exact node version",
			toolName:     "nodejs",
			toolVersion:  "20.10.0",
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
			toolName:     "go",
			toolVersion:  "1.21.5",
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
			toolName:     "python",
			toolVersion:  "3.12.0",
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
			name:        "install with :latest suffix",
			toolName:    "nodejs",
			toolVersion: "22:latest",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "22")
			},
		},
		{
			name:        "install with :installed suffix",
			toolName:    "ruby",
			toolVersion: "3:installed",
			validateOutput: func(t *testing.T, output string, err error) {
				// May error if no ruby 3.x versions are installed
				if err != nil {
					assert.True(t,
						strings.Contains(output, "no installed versions") ||
							strings.Contains(output, "not installed") ||
							strings.Contains(output, "failed"),
						"should indicate error, but was: %s", output)
				}
			},
		},
		{
			name:     "error on missing version argument",
			toolName: "nodejs",
			validateOutput: func(t *testing.T, output string, err error) {
				require.Error(t, err)
				assert.Contains(t, output, "requires 2 arguments")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"tools", "install", tt.toolName}
			if tt.toolVersion != "" {
				args = append(args, tt.toolVersion)
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

func TestToolsLatestCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		toolName       string
		toolVersion    string
		outputFormat   string
		provider       string
		validateOutput func(t *testing.T, output string, err error)
	}{
		{
			name:     "get latest without version prefix",
			toolName: "nodejs",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// Plaintext format should just output the version
				assert.NotEmpty(t, output, "should return a version")
			},
		},
		{
			name:        "get latest with version prefix",
			toolName:    "nodejs",
			toolVersion: "20",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// Plaintext format should just output the version
				assert.True(t, strings.HasPrefix(output, "20."), "version should start with 20., but was: %s", output)
			},
		},
		{
			name:     "get latest without version",
			toolName: "python",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// Plaintext format should just output the version
				assert.NotEmpty(t, output, "should return a version")
			},
		},
		{
			name:        "get latest installed version",
			toolName:    "ruby",
			toolVersion: "installed",
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
			toolName:     "python",
			toolVersion:  "3",
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
			toolName:     "go",
			toolVersion:  "1.21",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				// In plaintext format, it should just output the version string
				assert.True(t, strings.HasPrefix(output, "1.21"), "plaintext output should be just the version starting with 1.21, but was: %s", output)
			},
		},
		{
			name:         "latest with :installed suffix and JSON format",
			toolName:     "nodejs",
			toolVersion:  "20:installed",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				// May error if no node 20.x versions are installed
				if err == nil {
					var result map[string]string
					jsonErr := json.Unmarshal([]byte(output), &result)
					require.NoError(t, jsonErr, "output should be valid JSON")
					assert.Equal(t, "nodejs", result["tool"])
					assert.True(t, strings.HasPrefix(result["version"], "20"), "version should start with 20, but was: %s", result["version"])
				}
			},
		},
		{
			name:         "error on invalid format",
			toolName:     "ruby",
			outputFormat: "invalid",
			validateOutput: func(t *testing.T, output string, err error) {
				require.Error(t, err)
				assert.Contains(t, output, "invalid --format")
			},
		},
		{
			name:     "get latest with mise provider",
			toolName: "ruby",
			provider: "mise",
			validateOutput: func(t *testing.T, output string, err error) {
				require.NoError(t, err, "output: %s", output)
				assert.NotEmpty(t, output, "should return a version")
			},
		},
		{
			name:        "get latest with asdf provider",
			toolName:    "nodejs",
			toolVersion: "20",
			provider:    "asdf",
			validateOutput: func(t *testing.T, output string, err error) {
				// asdf may not support prefix matching for all versions - it may require exact versions
				// If it fails, accept that as a known limitation
				if err != nil {
					assert.True(t,
						strings.Contains(output, "no match for requested version") ||
							strings.Contains(output, "failed to install"),
						"should fail with version matching error if asdf doesn't support prefix, but was: %s", output)
				} else {
					assert.True(t, strings.HasPrefix(output, "20."), "version should start with 20., but was: %s", output)
				}
			},
		},
		{
			name:         "get latest with asdf provider and JSON format",
			toolName:     "python",
			toolVersion:  "3",
			provider:     "asdf",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string, err error) {
				// asdf may not support prefix matching for all versions - it may require exact versions
				// If it fails, accept that as a known limitation
				if err != nil {
					assert.True(t,
						strings.Contains(output, "no match for requested version") ||
							strings.Contains(output, "failed to install"),
						"should fail with version matching error if asdf doesn't support prefix, but was: %s", output)
				} else {
					var result map[string]string
					jsonErr := json.Unmarshal([]byte(output), &result)
					require.NoError(t, jsonErr, "output should be valid JSON")
					assert.Contains(t, result, "tool", "JSON should contain tool name")
					assert.Contains(t, result, "version", "JSON should contain version")
					assert.Equal(t, "python", result["tool"])
					assert.True(t, strings.HasPrefix(result["version"], "3."), "version should start with 3., but was: %s", result["version"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"tools", "latest", tt.toolName}
			if tt.toolVersion != "" {
				args = append(args, tt.toolVersion)
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
