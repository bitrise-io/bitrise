package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestParseVersionString(t *testing.T) {
	tests := []struct {
		name          string
		toolName      provider.ToolID
		versionString string
		expected      provider.ToolRequest
	}{
		{
			name:          "Exact version",
			toolName:      "golang",
			versionString: "1.20.3",
			expected: provider.ToolRequest{
				ToolName:           "golang",
				UnparsedVersion:    "1.20.3",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
		},
		{
			name:          "Installed version",
			toolName:      "nodejs",
			versionString: "20:installed",
			expected: provider.ToolRequest{
				ToolName:           "nodejs",
				UnparsedVersion:    "20",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
		},
		{
			name:          "Latest version",
			toolName:      "ruby",
			versionString: "3.2:latest",
			expected: provider.ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    "3.2",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
		},
		{
			name:          ":latest syntax",
			toolName:      "tuist",
			versionString: ":latest",
			expected: provider.ToolRequest{
				ToolName:           "tuist",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
		},
		{
			name:          "latest as version alias syntax",
			toolName:      "python",
			versionString: "latest",
			expected: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "latest",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
		},
		{
			name:          "Empty version string",
			toolName:      "docker",
			versionString: "",
			expected: provider.ToolRequest{
				ToolName:           "docker",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
		},
		{
			name:          "installed as version alias syntax",
			toolName:      "air",
			versionString: "installed",
			expected: provider.ToolRequest{
				ToolName:           "air",
				UnparsedVersion:    "installed",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
		},
		{
			name:          ":installed syntax",
			toolName:      "elixir",
			versionString: ":installed",
			expected: provider.ToolRequest{
				ToolName:           "elixir",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainVersion, resolutionStrategy, err := ParseVersionString(tt.versionString)
			assert.NoError(t, err)
			toolRequest := provider.ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    plainVersion,
				ResolutionStrategy: resolutionStrategy,
			}
			assert.Equal(t, tt.expected, toolRequest)
		})
	}
}
