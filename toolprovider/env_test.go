package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestPrependPaths(t *testing.T) {
	tests := []struct {
		name       string
		pathEnv    string
		pathsToAdd []string
		expected   string
	}{
		{
			name:       "empty path env, single new path",
			pathEnv:    "",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin",
		},
		{
			name:       "empty path env, multiple new paths",
			pathEnv:    "",
			pathsToAdd: []string{"/usr/local/bin", "/opt/bin"},
			expected:   "/usr/local/bin:/opt/bin",
		},
		{
			name:       "prepend single path to existing",
			pathEnv:    "/usr/bin:/bin",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:       "prepend multiple paths to existing",
			pathEnv:    "/usr/bin:/bin",
			pathsToAdd: []string{"/usr/local/bin", "/opt/bin"},
			expected:   "/usr/local/bin:/opt/bin:/usr/bin:/bin",
		},
		{
			name:       "remove duplicate and prepend",
			pathEnv:    "/usr/bin:/usr/local/bin:/bin",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:       "duplicate at end",
			pathEnv:    "/usr/bin:/bin:/usr/local/bin",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:       "multiple new paths with one duplicate",
			pathEnv:    "/usr/bin:/opt/bin:/bin",
			pathsToAdd: []string{"/usr/local/bin", "/opt/bin"},
			expected:   "/usr/local/bin:/opt/bin:/usr/bin:/bin",
		},
		{
			name:       "empty paths list",
			pathEnv:    "/usr/bin:/bin",
			pathsToAdd: []string{},
			expected:   "/usr/bin:/bin",
		},
		{
			name:       "both empty",
			pathEnv:    "",
			pathsToAdd: []string{},
			expected:   "",
		},
		{
			name:       "filter empty string entries in existing path",
			pathEnv:    "/usr/bin::/bin",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:       "multiple duplicates",
			pathEnv:    "/usr/local/bin:/usr/bin:/usr/local/bin:/bin:/usr/local/bin",
			pathsToAdd: []string{"/usr/local/bin"},
			expected:   "/usr/local/bin:/usr/bin:/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prependPaths(tt.pathEnv, tt.pathsToAdd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToEnvmanEnvs(t *testing.T) {
	// Use a static PATH for the duraton of tests
	t.Setenv("PATH", "/usr/bin:/bin")

	tests := []struct {
		name        string
		activations []provider.EnvironmentActivation
		expected    []envmanModels.EnvironmentItemModel
	}{
		{
			name:        "empty activations",
			activations: []provider.EnvironmentActivation{},
			expected:    []envmanModels.EnvironmentItemModel{},
		},
		{
			name: "single activation with env vars only",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"FOO": "bar",
						"BAZ": "qux",
					},
					ContributedPaths: []string{},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"FOO": "bar"},
				{"BAZ": "qux"},
			},
		},
		{
			name: "single activation with paths only",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{},
					ContributedPaths:   []string{"/usr/local/bin", "/opt/bin"},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"PATH": "/usr/local/bin:/opt/bin:/usr/bin:/bin"},
			},
		},
		{
			name: "single activation with both env vars and paths",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"NODE_ENV": "development",
					},
					ContributedPaths: []string{"/usr/local/bin"},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"NODE_ENV": "development"},
				{"PATH": "/usr/local/bin:/usr/bin:/bin"},
			},
		},
		{
			name: "multiple activations",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"TOOL1_HOME": "/opt/tool1",
					},
					ContributedPaths: []string{"/opt/tool1/bin"},
				},
				{
					ContributedEnvVars: map[string]string{
						"TOOL2_VERSION": "1.2.3",
					},
					ContributedPaths: []string{"/opt/tool2/bin"},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"TOOL1_HOME": "/opt/tool1"},
				{"TOOL2_VERSION": "1.2.3"},
				{"PATH": "/opt/tool1/bin:/opt/tool2/bin:/usr/bin:/bin"},
			},
		},
		{
			name: "activation with empty path entries",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"TEST": "value",
					},
					ContributedPaths: []string{"", "/valid/path", ""},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"TEST": "value"},
				{"PATH": "/valid/path:/usr/bin:/bin"},
			},
		},
		{
			name: "no contributed paths but has env vars",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"ONLY_VAR": "value",
					},
					ContributedPaths: []string{},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{
				{"ONLY_VAR": "value"},
			},
		},
		{
			name: "all empty paths filtered out",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{},
					ContributedPaths:   []string{"", "", ""},
				},
			},
			expected: []envmanModels.EnvironmentItemModel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToEnvmanEnvs(tt.activations)

			assert.Len(t, result, len(tt.expected))

			for _, expectedItem := range tt.expected {
				found := false
				for _, actualItem := range result {
					if len(actualItem) == len(expectedItem) {
						match := true
						for key, expectedValue := range expectedItem {
							if actualValue, exists := actualItem[key]; !exists || actualValue != expectedValue {
								match = false
								break
							}
						}
						if match {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Expected item not found: %v", expectedItem)
			}
		})
	}
}

func TestConvertToEnvMap(t *testing.T) {
	// Use a static PATH for the duraton of tests
	t.Setenv("PATH", "/usr/bin:/bin")

	tests := []struct {
		name        string
		activations []provider.EnvironmentActivation
		expected    map[string]string
	}{
		{
			name:        "empty activations",
			activations: []provider.EnvironmentActivation{},
			expected:    map[string]string{},
		},
		{
			name: "single activation with env vars only",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"FOO": "bar",
						"BAZ": "qux",
					},
					ContributedPaths: []string{},
				},
			},
			expected: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		{
			name: "single activation with paths only",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{},
					ContributedPaths:   []string{"/usr/local/bin", "/opt/bin"},
				},
			},
			expected: map[string]string{
				"PATH": "/usr/local/bin:/opt/bin:/usr/bin:/bin",
			},
		},
		{
			name: "single activation with both env vars and paths",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"NODE_ENV": "development",
					},
					ContributedPaths: []string{"/usr/local/bin"},
				},
			},
			expected: map[string]string{
				"NODE_ENV": "development",
				"PATH":     "/usr/local/bin:/usr/bin:/bin",
			},
		},
		{
			name: "multiple activations",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"TOOL1_HOME": "/opt/tool1",
					},
					ContributedPaths: []string{"/opt/tool1/bin"},
				},
				{
					ContributedEnvVars: map[string]string{
						"TOOL2_VERSION": "1.2.3",
					},
					ContributedPaths: []string{"/opt/tool2/bin"},
				},
			},
			expected: map[string]string{
				"TOOL1_HOME":    "/opt/tool1",
				"TOOL2_VERSION": "1.2.3",
				"PATH":          "/opt/tool1/bin:/opt/tool2/bin:/usr/bin:/bin",
			},
		},
		{
			name: "activation with empty path entries",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"TEST": "value",
					},
					ContributedPaths: []string{"", "/valid/path", ""},
				},
			},
			expected: map[string]string{
				"TEST": "value",
				"PATH": "/valid/path:/usr/bin:/bin",
			},
		},
		{
			name: "no contributed paths but has env vars",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"ONLY_VAR": "value",
					},
					ContributedPaths: []string{},
				},
			},
			expected: map[string]string{
				"ONLY_VAR": "value",
			},
		},
		{
			name: "all empty paths filtered out",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{},
					ContributedPaths:   []string{"", "", ""},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "overlapping env vars should keep last",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{
						"SHARED_VAR": "first",
					},
					ContributedPaths: []string{},
				},
				{
					ContributedEnvVars: map[string]string{
						"SHARED_VAR": "second",
					},
					ContributedPaths: []string{},
				},
			},
			expected: map[string]string{
				"SHARED_VAR": "second",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// var pathPtr *string
			// if tt.name == "nil pathEnv uses system PATH" {
			// 	os.Setenv("PATH", "")
			// 	pathPtr = nil
			// } else {
			// 	pathPtr = &tt.pathEnv
			// }

			result := ConvertToEnvMap(tt.activations) 

			assert.Equal(t, tt.expected, result)
		})
	}
}
