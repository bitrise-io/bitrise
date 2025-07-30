package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetToolRequests(t *testing.T) {
	tests := []struct {
		name     string
		config   models.BitriseDataModel
		expected []provider.ToolRequest
		wantErr  bool
	}{
		{
			name: "Empty tools",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{},
			},
			expected: []provider.ToolRequest{},
			wantErr:  false,
		},
		{
			name: "Multiple tools with different version strategies",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
					"nodejs": "20:installed",
					"ruby":   "3.2:latest",
				},
			},
			expected: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "20",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "ruby",
					UnparsedVersion:    "3.2",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple tools with some having plugin identifiers",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang":      "1.20.3",
					"custom-tool": "1.0.0",
					"nodejs":      "20:installed",
				},
				ToolConfig: &models.ToolConfigModel{
					ExtraPlugins: map[models.ToolID]string{
						"custom-tool": "https://github.com/example/custom-tool-plugin",
					},
				},
			},
			expected: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "custom-tool",
					UnparsedVersion:    "1.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   stringPtr("https://github.com/example/custom-tool-plugin"),
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "20",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Tool with nil ToolConfig",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
				},
				ToolConfig: nil,
			},
			expected: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Tool with ToolConfig but nil ExtraPlugins",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
				},
				ToolConfig: &models.ToolConfigModel{
					Provider:     "asdf",
					ExtraPlugins: nil,
				},
			},
			expected: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getToolRequests(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
