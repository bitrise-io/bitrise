package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

var (
	pluginName        = "tuist"
	pluginGitCloneURL = "https://github.com/tuist/asdf-tuist.git"
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
				ToolName:        provider.ToolID(pluginName),
				UnparsedVersion: "4.68.0",
				PluginURL:       &pluginGitCloneURL,
			},
			expected: &PluginSource{
				PluginName:  provider.ToolID(pluginName),
				GitCloneURL: pluginGitCloneURL,
			},
		},
		{
			name: "pluginIdentifier set with empty url and unknown tool ID",
			input: provider.ToolRequest{
				ToolName:        provider.ToolID(pluginName),
				UnparsedVersion: "4.68.0",
				PluginURL:       &emptyGitCloneURL,
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
