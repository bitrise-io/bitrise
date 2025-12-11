package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestConvertToOutputFormat(t *testing.T) {
	tests := []struct {
		name      string
		envs      []provider.EnvironmentActivation
		format    string
		wantLines []string // lines that must be present in the output
		want      string   // exact match (used for formats with deterministic order like JSON)
	}{
		{
			name:      "empty envs",
			envs:      []provider.EnvironmentActivation{},
			format:    outputFormatPlaintext,
			want:      "",
			wantLines: []string{},
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
			wantLines: []string{
				"Env vars to activate installed tools:",
				"NODE_VERSION=18.0.0",
				"NPM_CONFIG=/path/to/config",
				"PATH=/usr/local/node/bin:/usr/local/npm/bin:$PATH",
			},
		},
		{
			name: "json format with env vars and paths",
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
  "PATH": "/usr/local/go/bin:$PATH"
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
			want: "export JAVA_HOME=\"/usr/lib/jvm/java-17\"\n" +
				"export PATH=\"/usr/lib/jvm/java-17/bin:$PATH\"\n",
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
			format: outputFormatPlaintext,
			wantLines: []string{
				"Env vars to activate installed tools:",
				"TOOL_VERSION=2.0.0",
				"PATH=/path/one:/path/two:$PATH",
			},
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
			want:   "export VAR_WITH_SPACE=\"value with spaces\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToOutputFormat(tt.envs, tt.format)
			require.NoError(t, err)

			// If wantLines is specified, check that each line is present
			if len(tt.wantLines) > 0 {
				for _, line := range tt.wantLines {
					require.Contains(t, got, line, "output should contain line: %s", line)
				}
			} else if tt.want != "" {
				// For exact match (e.g., JSON with deterministic order)
				require.Equal(t, tt.want, got)
			} else {
				// Empty output
				require.Equal(t, "", got)
			}
		})
	}
}

func TestConvertToOutputFormat_InvalidFormat(t *testing.T) {
	envs := []provider.EnvironmentActivation{
		{
			ContributedEnvVars: map[string]string{"KEY": "value"},
		},
	}

	_, err := convertToOutputFormat(envs, "invalid")
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
			got := isYMLConfig(tt.path)
			require.Equal(t, tt.want, got)
		})
	}
}
