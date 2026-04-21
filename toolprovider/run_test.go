package toolprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindGitHubTokenEnv(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedName  string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "empty map",
			envs:          map[string]string{},
			expectedFound: false,
		},
		{
			name:          "no known tokens",
			envs:          map[string]string{"SOME_OTHER_VAR": "value"},
			expectedFound: false,
		},
		{
			name:          "GITHUB_TOKEN present",
			envs:          map[string]string{"GITHUB_TOKEN": "ghp_abc123"},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "ghp_abc123",
			expectedFound: true,
		},
		{
			name:          "MISE_GITHUB_TOKEN present",
			envs:          map[string]string{"MISE_GITHUB_TOKEN": "ghp_xyz"},
			expectedName:  "MISE_GITHUB_TOKEN",
			expectedValue: "ghp_xyz",
			expectedFound: true,
		},
		{
			name: "GITHUB_TOKEN takes priority over MISE_GITHUB_TOKEN",
			envs: map[string]string{
				"GITHUB_TOKEN":      "github_value",
				"MISE_GITHUB_TOKEN": "mise_value",
			},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "github_value",
			expectedFound: true,
		},
		{
			name: "GITHUB_TOKEN takes priority over GITHUB_API_TOKEN",
			envs: map[string]string{
				"GITHUB_TOKEN":     "github_value",
				"GITHUB_API_TOKEN": "api_value",
			},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "github_value",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, found := findGitHubTokenEnv(tt.envs)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
