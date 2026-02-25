package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetContainerConfig(t *testing.T) {
	tests := []struct {
		name           string
		input          ContainerReference
		expectedConfig *ContainerConfig
		errorContains  string
	}{
		// Success cases - nil and string formats
		{
			name:           "nil reference returns nil",
			input:          nil,
			expectedConfig: nil,
		},
		{
			name:  "string format - simple container ID",
			input: "redis",
			expectedConfig: &ContainerConfig{
				ContainerID: "redis",
				Recreate:    false,
			},
		},

		// Success cases - map format with recreate flag
		{
			name: "map[string]any format - empty config (no recreate)",
			input: map[string]any{
				"redis": map[any]any{},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "redis",
				Recreate:    false,
			},
		},
		{
			name: "map[string]any format - map[any]any recreate true",
			input: map[string]any{
				"postgres": map[any]any{
					"recreate": true,
				},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "postgres",
				Recreate:    true,
			},
		},
		{
			name: "map[string]any format - map[string]any recreate false",
			input: map[string]any{
				"mysql": map[string]any{
					"recreate": false,
				},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "mysql",
				Recreate:    false,
			},
		},
		{
			name: "map[any]any format - empty config (no recreate)",
			input: map[any]any{
				"redis": map[any]any{},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "redis",
				Recreate:    false,
			},
		},
		{
			name: "map[any]any format - recreate true",
			input: map[any]any{
				"node": map[any]any{
					"recreate": true,
				},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "node",
				Recreate:    true,
			},
		},
		{
			name: "map[any]any format - recreate false",
			input: map[any]any{
				"golang": map[any]any{
					"recreate": false,
				},
			},
			expectedConfig: &ContainerConfig{
				ContainerID: "golang",
				Recreate:    false,
			},
		},

		// Error cases - empty container id
		{
			name:          "string format - empty container id",
			input:         "",
			errorContains: "empty container id",
		},
		{
			name: "map[string]any format - empty container id",
			input: map[string]any{
				"": map[any]any{},
			},
			errorContains: "empty container id",
		},

		// Error cases - invalid types
		{
			name:          "error - invalid type (int)",
			input:         123,
			errorContains: "invalid container config type: int",
		},

		// Error cases - map validation
		{
			name: "error - map with multiple container IDs",
			input: map[string]any{
				"redis": map[any]any{
					"recreate": true,
				},
				"postgres": map[any]any{
					"recreate": false,
				},
			},
			errorContains: "invalid container config map length: 2",
		},
		{
			name:          "error - empty map",
			input:         map[string]any{},
			errorContains: "invalid container config map length: 0",
		},

		// Error cases - container value validation
		{
			name: "error - container value is not a map (string)",
			input: map[string]any{
				"redis": "invalid",
			},
			errorContains: "invalid container config value type: string",
		},

		// Error cases - config map validation
		{
			name: "error - config map has too many keys",
			input: map[string]any{
				"redis": map[any]any{
					"recreate": true,
					"restart":  "always",
				},
			},
			errorContains: "invalid container config value map length: 2",
		},
		{
			name: "error - config has one key but not 'recreate'",
			input: map[string]any{
				"redis": map[any]any{
					"restart": "always",
				},
			},
			errorContains: "missing recreate key in container config",
		},

		// Error cases - recreate value validation
		{
			name: "error - recreate value is string",
			input: map[string]any{
				"redis": map[any]any{
					"recreate": "yes",
				},
			},
			errorContains: "invalid recreate value type: string",
		},

		// Error cases - map key validation
		{
			name: "error - map with non-string key (int)",
			input: map[any]any{
				123: map[any]any{
					"recreate": true,
				},
			},
			errorContains: "invalid container config type: map[interface {}]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetContainerConfig(tt.input)

			if tt.errorContains != "" {
				require.EqualError(t, err, tt.errorContains)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedConfig, config)
			}
		})
	}
}
