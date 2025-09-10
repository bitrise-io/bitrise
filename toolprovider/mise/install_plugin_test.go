package mise

import (
	"errors"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRegistryChecker is a mock implementation of RegistryChecker for testing
type MockRegistryChecker struct {
	registryTools map[string]bool
}

func NewMockRegistryChecker() *MockRegistryChecker {
	return &MockRegistryChecker{
		registryTools: make(map[string]bool),
	}
}

func (m *MockRegistryChecker) SetToolInRegistry(toolName string, exists bool) {
	m.registryTools[toolName] = exists
}

func (m *MockRegistryChecker) isPluginInRegistry(name string) error {
	if exists, found := m.registryTools[name]; found && exists {
		return nil
	}
	return errors.New("tool not found in registry")
}

func TestPluginToInstall(t *testing.T) {
	tests := []struct {
		name           string
		tool           provider.ToolRequest
		registryTools  map[string]bool
		expectedPlugin *PluginSource
		expectedError  string
	}{
		{
			name: "empty tool name should return error",
			tool: provider.ToolRequest{
				ToolName:        "",
				UnparsedVersion: "1.0.0",
			},
			expectedError: "tool name is not defined for plugin installation",
		},
		{
			name: "custom plugin URL provided should return plugin source",
			tool: provider.ToolRequest{
				ToolName:        "custom-tool",
				UnparsedVersion: "1.0.0",
				PluginURL:       stringPtr("https://github.com/custom/plugin.git"),
			},
			expectedPlugin: &PluginSource{
				PluginName:  "custom-tool",
				GitCloneURL: "https://github.com/custom/plugin.git",
			},
		},
		{
			name: "custom plugin URL with whitespace should be trimmed",
			tool: provider.ToolRequest{
				ToolName:        "custom-tool",
				UnparsedVersion: "1.0.0",
				PluginURL:       stringPtr("  https://github.com/custom/plugin.git  "),
			},
			expectedPlugin: &PluginSource{
				PluginName:  "custom-tool",
				GitCloneURL: "https://github.com/custom/plugin.git",
			},
		},
		{
			name: "empty plugin URL should fallback to registry check",
			tool: provider.ToolRequest{
				ToolName:        "known-tool",
				UnparsedVersion: "1.0.0",
				PluginURL:       stringPtr(""),
			},
			registryTools: map[string]bool{
				"known-tool": true,
			},
			expectedPlugin: nil, // registry tool, no plugin needed
		},
		{
			name: "core tool nodejs should return nil (no plugin needed)",
			tool: provider.ToolRequest{
				ToolName:        "nodejs",
				UnparsedVersion: "18.0.0",
			},
			expectedPlugin: nil,
		},
		{
			name: "core tool go should return nil (no plugin needed)",
			tool: provider.ToolRequest{
				ToolName:        "go",
				UnparsedVersion: "1.19.0",
			},
			expectedPlugin: nil,
		},
		{
			name: "tool in registry should return nil (no plugin needed)",
			tool: provider.ToolRequest{
				ToolName:        "registry-tool",
				UnparsedVersion: "2.0.0",
			},
			registryTools: map[string]bool{
				"registry-tool": true,
			},
			expectedPlugin: nil,
		},
		{
			name: "unknown tool not in registry should return ToolInstallError",
			tool: provider.ToolRequest{
				ToolName:        "unknown-tool",
				UnparsedVersion: "1.0.0",
			},
			registryTools: map[string]bool{
				"unknown-tool": false,
			},
			expectedError: "This tool integration (unknown-tool) is not tested or vetted by Bitrise.",
		},
		{
			name: "core tool with custom plugin URL should use custom URL",
			tool: provider.ToolRequest{
				ToolName:        "nodejs", // This is a core tool
				UnparsedVersion: "18.0.0",
				PluginURL:       stringPtr("https://github.com/custom/nodejs-plugin.git"),
			},
			expectedPlugin: &PluginSource{
				PluginName:  "nodejs",
				GitCloneURL: "https://github.com/custom/nodejs-plugin.git",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChecker := NewMockRegistryChecker()
			for tool, exists := range tt.registryTools {
				mockChecker.SetToolInRegistry(tool, exists)
			}

			result, err := pluginToInstall(tt.tool, mockChecker)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)

				// Check if it's a ToolInstallError
				var toolError provider.ToolInstallError
				if errors.As(err, &toolError) {
					assert.Equal(t, tt.tool.ToolName, toolError.ToolName)
					assert.Equal(t, tt.tool.UnparsedVersion, toolError.RequestedVersion)
					assert.Contains(t, toolError.Recommendation, string(tt.tool.ToolName))
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPlugin, result)
			}
		})
	}
}

func TestPluginToInstall_CoreToolsList(t *testing.T) {
	// Test all core tools to ensure they don't require plugin installation
	mockChecker := NewMockRegistryChecker()

	for _, coreTool := range miseCoreTools {
		t.Run("core_tool_"+coreTool, func(t *testing.T) {
			tool := provider.ToolRequest{
				ToolName:        provider.ToolID(coreTool),
				UnparsedVersion: "1.0.0",
			}

			result, err := pluginToInstall(tool, mockChecker)

			require.NoError(t, err)
			assert.Nil(t, result, "Core tool %s should not require plugin installation", coreTool)
		})
	}
}

func TestPluginToInstall_ErrorDetailsValidation(t *testing.T) {
	// Test that ToolInstallError contains the expected fields
	mockChecker := NewMockRegistryChecker()
	mockChecker.SetToolInRegistry("unknown-tool", false)

	tool := provider.ToolRequest{
		ToolName:        "unknown-tool",
		UnparsedVersion: "2.3.4",
	}

	result, err := pluginToInstall(tool, mockChecker)

	require.Error(t, err)
	assert.Nil(t, result)

	var toolError provider.ToolInstallError
	require.ErrorAs(t, err, &toolError)

	assert.Equal(t, provider.ToolID("unknown-tool"), toolError.ToolName)
	assert.Equal(t, "2.3.4", toolError.RequestedVersion)
	assert.Contains(t, toolError.Cause, "This tool integration (unknown-tool) is not tested or vetted by Bitrise.")
	assert.Contains(t, toolError.Recommendation, "unknown-tool: https://github/url/to/asdf/plugin/repo.git")
}

func TestMiseCoreTools_Consistency(t *testing.T) {
	// Known aliases in the miseCoreTools list (not in the core backend, but can be used)
	aliases := []string{
		"golang", // go
		"nodejs", // node
	}

	expectedUniqueTools := 12
	expectedAliases := 2
	expectedTotalLength := expectedUniqueTools + expectedAliases

	assert.Len(t, miseCoreTools, expectedTotalLength,
		"miseCoreTools should have %d unique tools + %d aliases = %d total entries",
		expectedUniqueTools, expectedAliases, expectedTotalLength)

	// Verify that the known aliases are present
	for _, alias := range aliases {
		assert.Contains(t, miseCoreTools, alias, "Expected alias %s should be in miseCoreTools", alias)
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, tool := range miseCoreTools {
		assert.False(t, seen[tool], "Duplicate tool found in miseCoreTools: %s", tool)
		seen[tool] = true
	}
}

// stringPtr helper for creating string pointers
func stringPtr(s string) *string {
	return &s
}
