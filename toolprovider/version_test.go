package toolprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersionString(t *testing.T) {
	tests := []struct {
		name          string
		toolName      ToolID
		versionString string
		expected      ToolRequest
	}{
		{
			name:          "Exact version",
			toolName:      "golang",
			versionString: "1.20.3",
			expected: ToolRequest{
				ToolName:           "golang",
				UnparsedVersion:    "1.20.3",
				ResolutionStrategy: ResolutionStrategyStrict,
			},
		},
		{
			name:          "Installed version",
			toolName:      "nodejs",
			versionString: "20:installed",
			expected: ToolRequest{
				ToolName:           "nodejs",
				UnparsedVersion:    "20",
				ResolutionStrategy: ResolutionStrategyLatestInstalled,
			},
		},
		{
			name:          "Latest version",
			toolName:      "ruby",
			versionString: "3.2:latest",
			expected: ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    "3.2",
				ResolutionStrategy: ResolutionStrategyLatestReleased,
			},
		},
		{
			name:          ":latest syntax",
			toolName:      "tuist",
			versionString: ":latest",
			expected: ToolRequest{
				ToolName:           "tuist",
				UnparsedVersion:    "",
				ResolutionStrategy: ResolutionStrategyLatestReleased,
			},
		},
		{
			name:          "latest as version alias syntax",
			toolName:      "python",
			versionString: "latest",
			expected: ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "latest",
				ResolutionStrategy: ResolutionStrategyStrict,
			},
		},
		{
			name:          "Empty version string",
			toolName:      "docker",
			versionString: "",
			expected: ToolRequest{
				ToolName:           "docker",
				UnparsedVersion:    "",
				ResolutionStrategy: ResolutionStrategyStrict,
			},
		},
		{
			name:          "installed as version alias syntax",
			toolName:      "air",
			versionString: "installed",
			expected: ToolRequest{
				ToolName:           "air",
				UnparsedVersion:    "installed",
				ResolutionStrategy: ResolutionStrategyStrict,
			},
		},
		{
			name:          ":installed syntax",
			toolName:      "elixir",
			versionString: ":installed",
			expected: ToolRequest{
				ToolName:           "elixir",
				UnparsedVersion:    "",
				ResolutionStrategy: ResolutionStrategyLatestInstalled,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainVersion, resolutionStrategy, err := ParseVersionString(tt.versionString)
			assert.NoError(t, err)
			toolRequest := ToolRequest{
				ToolName:           tt.toolName,
				UnparsedVersion:    plainVersion,
				ResolutionStrategy: resolutionStrategy,
			}
			assert.Equal(t, tt.expected, toolRequest)
		})
	}
}
