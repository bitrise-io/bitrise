package mise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessEnvs(t *testing.T) {
	tests := []struct {
		name                     string
		envs                     envOutput
		currentPath              string
		expectedContributedVars  map[string]string
		expectedContributedPaths []string
	}{
		{
			name:                     "empty envs",
			envs:                     envOutput{},
			currentPath:              "/usr/bin:/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: nil,
		},
		{
			name: "only non-PATH env vars",
			envs: envOutput{
				"NODE_ENV":    "production",
				"JAVA_HOME":   "/opt/java",
				"PYTHON_PATH": "/opt/python/lib",
			},
			currentPath: "/usr/bin:/bin",
			expectedContributedVars: map[string]string{
				"NODE_ENV":    "production",
				"JAVA_HOME":   "/opt/java",
				"PYTHON_PATH": "/opt/python/lib",
			},
			expectedContributedPaths: nil,
		},
		{
			name: "PATH with new directories",
			envs: envOutput{
				"PATH": "/opt/mise/bin:/opt/tool/bin:/usr/bin:/bin",
			},
			currentPath:              "/usr/bin:/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: []string{"/opt/mise/bin", "/opt/tool/bin"},
		},
		{
			name: "PATH with no new directories",
			envs: envOutput{
				"PATH": "/usr/bin:/bin",
			},
			currentPath:              "/usr/bin:/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: nil,
		},
		{
			name: "PATH with some new and some existing directories",
			envs: envOutput{
				"PATH": "/opt/mise/bin:/usr/bin:/opt/tool/bin:/bin",
			},
			currentPath:              "/usr/bin:/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: []string{"/opt/mise/bin", "/opt/tool/bin"},
		},
		{
			name: "mixed env vars and PATH",
			envs: envOutput{
				"NODE_ENV":  "development",
				"PATH":      "/opt/node/bin:/usr/bin:/bin",
				"JAVA_HOME": "/opt/java",
			},
			currentPath: "/usr/bin:/bin",
			expectedContributedVars: map[string]string{
				"NODE_ENV":  "development",
				"JAVA_HOME": "/opt/java",
			},
			expectedContributedPaths: []string{"/opt/node/bin"},
		},
		{
			name: "PATH with empty current PATH",
			envs: envOutput{
				"PATH": "/opt/mise/bin:/opt/tool/bin",
			},
			currentPath:              "",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: []string{"/opt/mise/bin", "/opt/tool/bin"},
		},
		{
			name: "empty PATH in envs",
			envs: envOutput{
				"PATH":     "",
				"NODE_ENV": "test",
			},
			currentPath: "/usr/bin:/bin",
			expectedContributedVars: map[string]string{
				"NODE_ENV": "test",
			},
			expectedContributedPaths: nil,
		},
		{
			name: "PATH with single directory",
			envs: envOutput{
				"PATH": "/opt/tool/bin",
			},
			currentPath:              "/usr/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: []string{"/opt/tool/bin"},
		},
		{
			name: "duplicate paths should not be added",
			envs: envOutput{
				"PATH": "/opt/new/bin:/usr/bin:/opt/new/bin:/bin",
			},
			currentPath:              "/usr/bin:/bin",
			expectedContributedVars:  map[string]string{},
			expectedContributedPaths: []string{"/opt/new/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the current PATH environment
			t.Setenv("PATH", tt.currentPath)

			result := processEnvOutput(tt.envs)

			require.Equal(t, tt.expectedContributedVars, result.ContributedEnvVars, "ContributedEnvVars should match")
			require.Equal(t, tt.expectedContributedPaths, result.ContributedPaths, "ContributedPaths should match")
		})
	}
}

func TestProcessEnvs_PreservesOrder(t *testing.T) {
	// Test that the order of contributed paths is preserved (first paths in mise PATH that are not in current PATH)
	envs := envOutput{
		"PATH": "/first/bin:/second/bin:/usr/bin:/third/bin:/bin",
	}

	t.Setenv("PATH", "/usr/bin:/bin")

	result := processEnvOutput(envs)

	expectedPaths := []string{"/first/bin", "/second/bin", "/third/bin"}
	require.Equal(t, expectedPaths, result.ContributedPaths, "Order of contributed paths should be preserved")
}
