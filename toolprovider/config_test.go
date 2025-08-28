package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetToolRequests(t *testing.T) {
	tests := []struct {
		name       string
		config     models.BitriseDataModel
		workflowID string
		expected   []provider.ToolRequest
		wantErr    bool
	}{
		{
			name: "global tools only - empty tools",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: nil,
					},
				},
			},
			workflowID: "test",
			expected:   []provider.ToolRequest{},
			wantErr:    false,
		},
		{
			name: "global tools only - multiple tools with different version strategies",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
					"nodejs": "20:installed",
					"ruby":   "3.2:latest",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: nil,
					},
				},
			},
			workflowID: "test",
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
			name: "workflow tools only - no global tools",
			config: models.BitriseDataModel{
				Tools: nil,
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"python": "3.9.0",
							"node":   "16:latest",
						},
					},
				},
			},
			workflowID: "test",
			expected: []provider.ToolRequest{
				{
					ToolName:           "python",
					UnparsedVersion:    "3.9.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "node",
					UnparsedVersion:    "16",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "workflow tools override global tools",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"python": "3.8.0",
					"ruby":   "2.7.0",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"python": "3.9.0",     // Override global python version
							"node":   "16:latest", // Additional tool not in global
						},
					},
				},
			},
			workflowID: "test",
			expected: []provider.ToolRequest{
				{
					ToolName:           "python",
					UnparsedVersion:    "3.9.0", // Should use workflow version
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "ruby",
					UnparsedVersion:    "2.7.0", // Should use global version
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "node",
					UnparsedVersion:    "16", // Workflow-only tool
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "workflow tools unset some global tools",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"python": "3.8.0",
					"ruby":   "2.7.0",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"python": "3.9.0",     // Override global python version
							"node":   "16:latest", // Additional tool not in global
							"ruby":   "unset",     // Unset global ruby version
						},
					},
				},
			},
			workflowID: "test",
			expected: []provider.ToolRequest{
				{
					ToolName:           "python",
					UnparsedVersion:    "3.9.0", // Should use workflow version
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "node",
					UnparsedVersion:    "16", // Workflow-only tool
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "tools with plugin identifiers",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang":      "1.20.3",
					"custom-tool": "1.0.0",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"nodejs": "20:installed",
						},
					},
				},
				ToolConfig: &models.ToolConfigModel{
					ExtraPlugins: map[models.ToolID]string{
						"custom-tool": "https://github.com/example/custom-tool-plugin",
						"nodejs":      "https://github.com/example/nodejs-plugin",
					},
				},
			},
			workflowID: "test",
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
					PluginIdentifier:   stringPtr("https://github.com/example/nodejs-plugin"),
				},
			},
			wantErr: false,
		},
		{
			name: "no global tools with workflow tools",
			config: models.BitriseDataModel{
				Tools: nil,
				Workflows: map[string]models.WorkflowModel{
					"deploy": {
						Tools: models.ToolsModel{
							"ruby": "2.7:installed",
							"go":   "1.19.5",
						},
					},
				},
			},
			workflowID: "deploy",
			expected: []provider.ToolRequest{
				{
					ToolName:           "ruby",
					UnparsedVersion:    "2.7",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "go",
					UnparsedVersion:    "1.19.5",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "empty global tools with workflow tools",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"java": "11.0.16",
						},
					},
				},
			},
			workflowID: "test",
			expected: []provider.ToolRequest{
				{
					ToolName:           "java",
					UnparsedVersion:    "11.0.16",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "tool with no ToolConfig",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {},
				},
				ToolConfig: nil,
			},
			workflowID: "test",
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
			name: "tool with ToolConfig but nil ExtraPlugins",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"golang": "1.20.3",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {},
				},
				ToolConfig: &models.ToolConfigModel{
					Provider:     "asdf",
					ExtraPlugins: nil,
				},
			},
			workflowID: "test",
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
			name: "both global and workflow tools empty",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{},
					},
				},
			},
			workflowID: "test",
			expected:   []provider.ToolRequest{},
			wantErr:    false,
		},
		{
			name: "both global and workflow tools nil",
			config: models.BitriseDataModel{
				Tools: nil,
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: nil,
					},
				},
			},
			workflowID: "test",
			expected:   []provider.ToolRequest{},
			wantErr:    false,
		},
		{
			name: "whitespace in version strings",
			config: models.BitriseDataModel{
				Tools: models.ToolsModel{
					"python": "  3.9.0  ",
				},
				Workflows: map[string]models.WorkflowModel{
					"test": {
						Tools: models.ToolsModel{
							"node": "\t16:latest\n",
							"ruby": " 2.7:installed ",
						},
					},
				},
			},
			workflowID: "test",
			expected: []provider.ToolRequest{
				{
					ToolName:           "python",
					UnparsedVersion:    "3.9.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "node",
					UnparsedVersion:    "16",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
					PluginIdentifier:   nil,
				},
				{
					ToolName:           "ruby",
					UnparsedVersion:    "2.7",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
					PluginIdentifier:   nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getToolRequests(tt.config, tt.workflowID)

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
