package toolprovider

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestPrependPath(t *testing.T) {
	tests := []struct {
		name        string
		pathEnv     string
		newPath     string
		expected    string
	}{
		{
			name:        "empty path env",
			pathEnv:     "",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin",
		},
		{
			name:        "prepend to existing path",
			pathEnv:     "/usr/bin:/bin",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:        "remove duplicate and prepend",
			pathEnv:     "/usr/bin:/usr/local/bin:/bin",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:        "duplicate at end",
			pathEnv:     "/usr/bin:/bin:/usr/local/bin",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:        "single path duplicate",
			pathEnv:     "/usr/local/bin",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin",
		},
		{
			name:        "empty new path",
			pathEnv:     "/usr/bin:/bin",
			newPath:     "",
			expected:    ":/usr/bin:/bin",
		},
		{
			name:        "multiple duplicates",
			pathEnv:     "/usr/local/bin:/usr/bin:/usr/local/bin:/bin:/usr/local/bin",
			newPath:     "/usr/local/bin",
			expected:    "/usr/local/bin:/usr/bin:/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prependPath(tt.pathEnv, tt.newPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToEnvmanEnvs(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name        string
		activations []provider.EnvironmentActivation
		pathEnv     string
		expected    []envmanModels.EnvironmentItemModel
	}{
		{
			name:        "empty activations",
			activations: []provider.EnvironmentActivation{},
			pathEnv:     "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
			expected: []envmanModels.EnvironmentItemModel{
				{"TEST": "value"},
				{"PATH": "/valid/path:/usr/bin:/bin"},
			},
		},
		{
			name: "empty PATH environment",
			activations: []provider.EnvironmentActivation{
				{
					ContributedEnvVars: map[string]string{},
					ContributedPaths:   []string{"/new/path"},
				},
			},
			pathEnv: "",
			expected: []envmanModels.EnvironmentItemModel{
				{"PATH": "/new/path"},
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
			pathEnv: "/usr/bin:/bin",
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
			pathEnv: "/usr/bin:/bin",
			expected: []envmanModels.EnvironmentItemModel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("PATH", tt.pathEnv)
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
