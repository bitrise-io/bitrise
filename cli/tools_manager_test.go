package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestConvertToOutputFormat(t *testing.T) {
	tests := []struct {
		name   string
		envs   []provider.EnvironmentActivation
		format string
		want   string
	}{
		{
			name:   "empty envs",
			envs:   []provider.EnvironmentActivation{},
			format: outputFormatPlaintext,
			want:   "",
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
			want: "Env vars to activate installed tools:\n" +
				"NODE_VERSION=18.0.0\n" +
				"NPM_CONFIG=/path/to/config\n" +
				"PATH=/usr/local/node/bin:/usr/local/npm/bin:$PATH\n",
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
			want: "Env vars to activate installed tools:\n" +
				"TOOL_VERSION=2.0.0\n" +
				"PATH=/path/one:/path/two:$PATH\n",
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
