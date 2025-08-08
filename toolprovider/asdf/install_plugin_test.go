package asdf

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

var (
	pluginNameKnown   = "nodejs"
	pluginNameUnknown = "unknown-tool"
	pluginGitCloneURL = "https://github.com/asdf-vm/asdf-nodejs.git"
	emptyGitCloneURL  = ""
)

func TestResolvePluginSource(t *testing.T) {
	tests := []struct {
		name     string
		input    provider.ToolRequest
		expected *PluginSource
	}{
		{
			name: "pluginIdentifier set in correct format",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginNameUnknown),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &pluginGitCloneURL,
			},
			expected: &PluginSource{
				PluginName:  provider.ToolID(pluginNameUnknown),
				GitCloneURL: pluginGitCloneURL,
			},
		},
		{
			name: "pluginIdentifier set with empty url but known tool ID",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginNameKnown),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &emptyGitCloneURL,
			},
			expected: &PluginSource{
				PluginName:  provider.ToolID(pluginNameKnown),
				GitCloneURL: pluginGitCloneURL,
			},
		},
		{
			name: "pluginIdentifier set with empty url and unknown tool ID",
			input: provider.ToolRequest{
				ToolName:         provider.ToolID(pluginNameUnknown),
				UnparsedVersion:  "18.16.0",
				PluginIdentifier: &emptyGitCloneURL,
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePluginSource(tt.input)
			assert.Equal(t, tt.expected, got, "Expected plugin source to match")
		})
	}
}
