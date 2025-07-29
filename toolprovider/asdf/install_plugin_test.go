package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

var (
	pluginName                = "nodejs"
	pluginGitCloneURL         = "https://github.com/asdf-vm/asdf-nodejs.git"
	fullPluginId              = pluginName + "::" + pluginGitCloneURL
	nameOnlySeparatorPluginId = pluginName + "::"
	urlOnlySeparatorPluginId  = "::" + pluginGitCloneURL
	multipleSeparatorPluginId = pluginName + "::" + "latest" + "::" + pluginGitCloneURL
)

func TestResolvePluginSource(t *testing.T) {
	tests := []struct {
		name     string
		input    provider.ToolRequest
		expected PluginSource
		wantErr  bool
	}{
		{
			name: "pluginIdentifier set in correct format",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &fullPluginId,
			},
			expected: PluginSource{
				provider.ToolID(pluginName),
				pluginGitCloneURL,
			},
		},
		{
			name: "pluginIdentifier set with name only",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &pluginName,
			},
			expected: PluginSource{
				PluginName: provider.ToolID(pluginName),
				GitCloneURL:  "",
			},
		},
		{
			name: "pluginIdentifier set with url only",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &pluginGitCloneURL,
			},
			expected: PluginSource{},
			wantErr:  true, // Expecting an error because only URL is provided without a name
		},
		{
			name: "pluginIdentifier set with empty url",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &nameOnlySeparatorPluginId,
			},
			expected: PluginSource{
				PluginName: provider.ToolID(pluginName),
				GitCloneURL:  "",
			},
		},
		{
			name: "pluginIdentifier set with empty name",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &urlOnlySeparatorPluginId,
			},
			expected: PluginSource{},
			wantErr:  true,
		},
		{
			name: "pluginIdentifier set with multiple separators",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginName),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &multipleSeparatorPluginId,
			},
			expected: PluginSource{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchPluginSource(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for input: %v", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input: %v", tt.input)
				assert.Equal(t, tt.expected, *got, "Expected plugin source to match")
			}
		})
	}
}
