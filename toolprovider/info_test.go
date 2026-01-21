package toolprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "lines with whitespace",
			input:    "  line1  \n\tline2\t\n  line3  ",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines filtered out",
			input:    "line1\n\n\nline2\n   \nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "carriage return handling",
			input:    "line1\r\nline2\r\n",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiseToolEntryParsing(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected map[string][]miseToolEntry
		wantErr  bool
	}{
		{
			name: "single tool single version",
			json: `{
				"go": [
					{
						"version": "1.21.0",
						"installed": true,
						"active": false
					}
				]
			}`,
			expected: map[string][]miseToolEntry{
				"go": {
					{Version: "1.21.0", Installed: true, Active: false},
				},
			},
		},
		{
			name: "single tool multiple versions",
			json: `{
				"node": [
					{"version": "20.10.0", "installed": true, "active": true},
					{"version": "18.0.0", "installed": true, "active": false}
				]
			}`,
			expected: map[string][]miseToolEntry{
				"node": {
					{Version: "20.10.0", Installed: true, Active: true},
					{Version: "18.0.0", Installed: true, Active: false},
				},
			},
		},
		{
			name: "multiple tools",
			json: `{
				"go": [{"version": "1.21.0", "installed": true, "active": true}],
				"python": [{"version": "3.12.0", "installed": true, "active": false}]
			}`,
			expected: map[string][]miseToolEntry{
				"go":     {{Version: "1.21.0", Installed: true, Active: true}},
				"python": {{Version: "3.12.0", Installed: true, Active: false}},
			},
		},
		{
			name: "with source information",
			json: `{
				"go": [
					{
						"version": "1.21.0",
						"requested_version": "1.21",
						"source": {
							"type": ".tool-versions",
							"path": "/path/to/.tool-versions"
						},
						"installed": true,
						"active": true
					}
				]
			}`,
			expected: map[string][]miseToolEntry{
				"go": {
					{
						Version:   "1.21.0",
						Requested: "1.21",
						Source: &struct {
							Type string `json:"type"`
							Path string `json:"path"`
						}{
							Type: ".tool-versions",
							Path: "/path/to/.tool-versions",
						},
						Installed: true,
						Active:    true,
					},
				},
			},
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:     "empty object",
			json:     `{}`,
			expected: map[string][]miseToolEntry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string][]miseToolEntry
			err := parseJSON([]byte(tt.json), &result)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))

			for tool, entries := range tt.expected {
				resultEntries, exists := result[tool]
				require.True(t, exists, "tool %s should exist", tool)
				require.Equal(t, len(entries), len(resultEntries), "tool %s should have %d entries", tool, len(entries))

				for i, entry := range entries {
					assert.Equal(t, entry.Version, resultEntries[i].Version)
					assert.Equal(t, entry.Installed, resultEntries[i].Installed)
					assert.Equal(t, entry.Active, resultEntries[i].Active)
					if entry.Source != nil {
						require.NotNil(t, resultEntries[i].Source)
						assert.Equal(t, entry.Source.Path, resultEntries[i].Source.Path)
						assert.Equal(t, entry.Source.Type, resultEntries[i].Source.Type)
					}
				}
			}
		})
	}
}

func TestListInstalledToolsUnsupportedProvider(t *testing.T) {
	tools, err := ListInstalledTools("unsupported", false, false)
	require.Error(t, err)
	assert.Nil(t, tools)
	assert.Contains(t, err.Error(), "unsupported tool provider")
}

// parseJSON is a helper function for testing JSON parsing
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
