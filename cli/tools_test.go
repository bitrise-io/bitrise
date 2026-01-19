package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestConvertToOutputFormat(t *testing.T) {
	// Set a static $PATH for predictable output
	t.Setenv("PATH", "/usr/bin")

	tests := []struct {
		name   string
		envs   []provider.EnvironmentActivation
		format string
		want   string
	}{
		{
			name:   "empty envs and plaintext format",
			envs:   []provider.EnvironmentActivation{},
			format: outputFormatPlaintext,
			want:   "No new tools were installed.",
		},
		{
			name:   "empty envs and JSON format",
			envs:   []provider.EnvironmentActivation{},
			format: outputFormatJSON,
			want:   "{}",
		},
		{
			name:   "empty envs and Bash format",
			envs:   []provider.EnvironmentActivation{},
			format: outputFormatBash,
			want:   "# No new tools were installed.",
		},
		{
			name: "plaintext format with env vars and paths",
			envs: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"NODE_VERSION": "18.0.0",
						"NPM_CONFIG":   "/path/to/config",
					},
					ContributedPaths: []string{"/usr/local/node/bin", "/usr/local/npm/bin"},
				},
			},
			format: outputFormatPlaintext,
			want:   "\x1b[32;1m✓ Tools activated for subsequent steps in the workflow\x1b[0m\n\x1b[33;1m! If you need tools in the current shell session, run\x1b[0m \x1b[36;1meval \"$(bitrise tools setup --format bash ...)\"\x1b[0m \x1b[33;1minstead.\x1b[0m\n",
		},
		{
			name: "JSON format with env vars and paths",
			envs: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"GO_VERSION": "1.21.0",
					},
					ContributedPaths: []string{"/usr/local/go/bin"},
				},
			},
			format: outputFormatJSON,
			want: `{
  "GO_VERSION": "1.21.0",
  "PATH": "/usr/local/go/bin:/usr/bin"
}`,
		},
		{
			name: "bash format with env vars and paths",
			envs: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"JAVA_HOME": "/usr/lib/jvm/java-17",
					},
					ContributedPaths: []string{"/usr/lib/jvm/java-17/bin"},
				},
			},
			format: outputFormatBash,
			want:   "export JAVA_HOME=\"/usr/lib/jvm/java-17\"\nexport PATH=\"/usr/lib/jvm/java-17/bin:/usr/bin\"\n# \x1b[33;1mNOTE: Tools have been installed, but they need to be activated for the current shell session.\x1b[0m\n# Make sure to run \x1b[36;1meval \"$(bitrise tools setup --format bash ...)\"\x1b[0m instead\n",
		},
		{
			name: "multiple activations deduplicate env vars",
			envs: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"TOOL_VERSION": "1.0.0",
					},
					ContributedPaths: []string{"/path/one"},
				},
				{
					ContributedEnvVars: map[string]string{
						"TOOL_VERSION": "2.0.0", // Should override
					},
					ContributedPaths: []string{"/path/two"},
				},
			},
			format: outputFormatJSON,
			want: `{
  "PATH": "/path/one:/path/two:/usr/bin",
  "TOOL_VERSION": "2.0.0"
}`,
		},
		{
			name: "bash format quotes values properly",
			envs: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"VAR_WITH_SPACE": "value with spaces",
					},
				},
			},
			format: outputFormatBash,
			want:   "export VAR_WITH_SPACE=\"value with spaces\"\n# \x1b[33;1mNOTE: Tools have been installed, but they need to be activated for the current shell session.\x1b[0m\n# Make sure to run \x1b[36;1meval \"$(bitrise tools setup --format bash ...)\"\x1b[0m instead\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToOutputFormat(tt.envs, tt.format, true)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestConvertToOutputFormat_InvalidFormat(t *testing.T) {
	envs := []provider.EnvironmentActivation{
		{
			ContributedEnvVars: map[string]string{"KEY": "value"},
		},
	}

	_, err := convertToOutputFormat(envs, "invalid", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported output format")
}

func TestIsYMLConfig(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"bitrise.yml", true},
		{"bitrise.yaml", true},
		{"config.YML", true},
		{"config.YAML", true},
		{"/path/to/bitrise.yml", true},
		{".tool-versions", false},
		{".ruby-version", false},
		{"package.json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isBitriseConfig(tt.path)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseToolSpec(t *testing.T) {
	tests := []struct {
		name        string
		toolSpec    string
		isLatest    bool
		wantTool    string
		wantVersion string
		wantErr     bool
		errContains string
	}{
		{
			name:        "exact version with @",
			toolSpec:    "node@20.10.0",
			isLatest:    false,
			wantTool:    "node",
			wantVersion: "20.10.0",
			wantErr:     false,
		},
		{
			name:        "exact version with @ for python",
			toolSpec:    "python@3.12.1",
			isLatest:    false,
			wantTool:    "python",
			wantVersion: "3.12.1",
			wantErr:     false,
		},
		{
			name:        "version prefix with @ for latest",
			toolSpec:    "node@20",
			isLatest:    true,
			wantTool:    "node",
			wantVersion: "20",
			wantErr:     false,
		},
		{
			name:        "tool without version for latest",
			toolSpec:    "node",
			isLatest:    true,
			wantTool:    "node",
			wantVersion: "",
			wantErr:     false,
		},
		{
			name:        "tool without version for install (error)",
			toolSpec:    "node",
			isLatest:    false,
			wantErr:     true,
			errContains: "version required",
		},
		{
			name:        "empty version after @ for install (error)",
			toolSpec:    "node@",
			isLatest:    false,
			wantErr:     true,
			errContains: "version cannot be empty",
		},
		{
			name:        "empty tool name (error)",
			toolSpec:    "@20.10.0",
			isLatest:    false,
			wantErr:     true,
			errContains: "tool name cannot be empty",
		},
		{
			name:        "multiple @ symbols (error)",
			toolSpec:    "node@20@10",
			isLatest:    false,
			wantErr:     true,
			errContains: "invalid tool specification",
		},
		{
			name:        "tool with version prefix for latest",
			toolSpec:    "python@3.12",
			isLatest:    true,
			wantTool:    "python",
			wantVersion: "3.12",
			wantErr:     false,
		},
		{
			name:        "tool without @ for install (error)",
			toolSpec:    "ruby3.2.0",
			isLatest:    false,
			wantErr:     true,
			errContains: "version required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotTool, gotVersion string
			var err error

			if tt.isLatest {
				gotTool, gotVersion, err = parseToolSpec(tt.toolSpec, false)
			} else {
				gotTool, gotVersion, err = parseToolSpec(tt.toolSpec, true)
			}

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantTool, gotTool)
				require.Equal(t, tt.wantVersion, gotVersion)
			}
		})
	}
}

func TestExposeEnvsWithEnvman(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bitrise_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configs.BitriseWorkDirPath = tempDir

	envstorePath := filepath.Join(tempDir, "input_envstore.yml")
	configs.InputEnvstorePath = envstorePath

	t.Run("successfully exposes envs to existing envstore", func(t *testing.T) {
		// Create the envstore first (simulating workflow initialization)
		err := os.WriteFile(envstorePath, []byte("envs: []\n"), 0644)
		require.NoError(t, err, "should create envstore file")

		activations := []provider.EnvironmentActivation{
			{
				ContributedEnvVars: map[string]string{
					"TEST_VAR":     "test_value",
					"NODE_VERSION": "20.0.0",
				},
				ContributedPaths: []string{"/test/path"},
			},
		}

		result := exposeEnvsWithEnvman(activations, false)
		require.True(t, result, "exposeEnvsWithEnvman should return true")

		_, err = os.Stat(envstorePath)
		require.NoError(t, err, "envstore file should still exist")
	})

	t.Run("handles empty activations gracefully", func(t *testing.T) {
		os.Remove(envstorePath)
		// Create the envstore first
		err := os.WriteFile(envstorePath, []byte("envs: []\n"), 0644)
		require.NoError(t, err)

		activations := []provider.EnvironmentActivation{}

		result := exposeEnvsWithEnvman(activations, true)

		require.True(t, result, "exposeEnvsWithEnvman should return true even with empty activations")

		_, err = os.Stat(envstorePath)
		require.NoError(t, err, "envstore file should still exist")
	})

	t.Run("returns false when envstore does not exist", func(t *testing.T) {
		t.Setenv(configs.EnvstorePathEnvKey, "")
		originalPath := configs.InputEnvstorePath
		configs.InputEnvstorePath = "/path/to/nonexistent/envstore.yml"
		defer func() { configs.InputEnvstorePath = originalPath }()

		activations := []provider.EnvironmentActivation{
			{
				ContributedEnvVars: map[string]string{"KEY": "value"},
			},
		}

		// Test with silent=true (no warnings)
		result := exposeEnvsWithEnvman(activations, true)
		require.False(t, result, "should return false when envstore does not exist")

		// Test with silent=false and verify warning is logged
		var buf bytes.Buffer
		log.InitGlobalLogger(log.LoggerOpts{
			LoggerType:      log.ConsoleLogger,
			Producer:        log.BitriseCLI,
			DebugLogEnabled: false,
			Writer:          &buf,
		})

		result = exposeEnvsWithEnvman(activations, false)
		require.False(t, result, "should return false when envstore does not exist")

		// Verify warning message was logged
		logOutput := buf.String()
		require.Contains(t, logOutput, "Envstore not found", "should log warning about missing envstore")
		require.Contains(t, logOutput, "/path/to/nonexistent/envstore.yml", "should include the envstore path in warning")
	})

	t.Run("respects ENVMAN_ENVSTORE_PATH environment variable", func(t *testing.T) {
		customEnvstorePath := filepath.Join(tempDir, "custom_envstore.yml")
		// Create the custom envstore first
		err := os.WriteFile(customEnvstorePath, []byte("envs: []\n"), 0644)
		require.NoError(t, err)

		t.Setenv(configs.EnvstorePathEnvKey, customEnvstorePath)

		activations := []provider.EnvironmentActivation{
			{
				ContributedEnvVars: map[string]string{
					"GOPATH": "/go",
				},
			},
		}

		result := exposeEnvsWithEnvman(activations, false)
		require.True(t, result, "exposeEnvsWithEnvman should return true")

		// Verify custom envstore was used (still exists)
		_, err = os.Stat(customEnvstorePath)
		require.NoError(t, err, "custom envstore file should exist")
	})
}
