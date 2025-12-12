package toolprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeToolRequests(t *testing.T) {
	tests := []struct {
		name          string
		versionFiles  map[string]string // filename -> content
		wantErr       bool
		wantTools     []provider.ToolRequest
		errContains   string
	}{
		{
			name:          "empty version file list with no files in cwd",
			versionFiles:  map[string]string{},
			wantErr:       false,
			wantTools:     nil,
		},
		{
			name: "single .tool-versions file with one tool",
			versionFiles: map[string]string{
				".tool-versions": "nodejs 18.0.0",
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "18.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
			},
		},
		{
			name: ".tool-versions file with multiple tools",
			versionFiles: map[string]string{
				".tool-versions": `ruby 3.2.0
nodejs 18.0.0
python 3.11.0`,
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "ruby",
					UnparsedVersion:    "3.2.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "18.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "python",
					UnparsedVersion:    "3.11.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
			},
		},
		{
			name: ".tool-versions with :latest resolution strategy",
			versionFiles: map[string]string{
				".tool-versions": "nodejs 18:latest",
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "18",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
				},
			},
		},
		{
			name: ".tool-versions with :installed resolution strategy",
			versionFiles: map[string]string{
				".tool-versions": "python 3.11:installed",
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "python",
					UnparsedVersion:    "3.11",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
				},
			},
		},
		{
			name: "single-tool version files mixed",
			versionFiles: map[string]string{
				".tool-versions":  "ruby 2.7.0",
				".node-version":   "16.0.0",
				".python-version": "3.10.0",
			},
			wantErr:       false,
			wantTools:     []provider.ToolRequest{
				{
					ToolName:           "ruby",
					UnparsedVersion:    "2.7.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "16.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "python",
					UnparsedVersion:    "3.10.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
			},
		},
		{
			name: ".tool-versions with comments and blank lines",
			versionFiles: map[string]string{
				".tool-versions": `# This is a comment
ruby 3.2.0

# Another comment
nodejs 18.0.0`,
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "ruby",
					UnparsedVersion:    "3.2.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "18.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
			},
		},
		{
			name: "invalid version format in file",
			versionFiles: map[string]string{
				".tool-versions": "ruby",
			},
			wantErr: true,
			errContains: "invalid format, expected '<tool> <version>'",
		},
		{
			name: "various tools with mixed resolution strategies",
			versionFiles: map[string]string{
				".tool-versions": `nodejs 18.0.0
python 3.11:latest
ruby 3.2:installed`,
			},
			wantErr:       false,
			wantTools: []provider.ToolRequest{
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "18.0.0",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "python",
					UnparsedVersion:    "3.11",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
				},
				{
					ToolName:           "ruby",
					UnparsedVersion:    "3.2",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create version files
			for filename, content := range tt.versionFiles {
				path := filepath.Join(tmpDir, filename)
				err := os.WriteFile(path, []byte(content), 0644)
				require.NoError(t, err)
			}

			// Build explicit file paths for the test
			var filePaths []string
			for filename := range tt.versionFiles {
				filePaths = append(filePaths, filepath.Join(tmpDir, filename))
			}

			// If we have explicit paths, use them. Otherwise test auto-discovery.
			var versionFilePaths []string
			if len(filePaths) > 0 {
				versionFilePaths = filePaths
			}

			got, err := makeToolRequests(versionFilePaths, true)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.wantTools), len(got))

			if tt.wantTools != nil {
				assert.ElementsMatch(t, tt.wantTools, got)
			}
		})
	}
}

func TestMakeToolRequestsDuplicateTools(t *testing.T) {
	t.Run("duplicate tools from multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create two files with the same tool but different versions
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")
		err := os.WriteFile(toolVersionsPath, []byte("nodejs 18.0.0"), 0644)
		require.NoError(t, err)

		nodeVersionPath := filepath.Join(tmpDir, ".node-version")
		err = os.WriteFile(nodeVersionPath, []byte("16.0.0"), 0644)
		require.NoError(t, err)

		got, err := makeToolRequests([]string{toolVersionsPath, nodeVersionPath}, true)

		// Both tools should be in the result (duplicates are preserved)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(got))
		assert.Equal(t, "18.0.0", got[0].UnparsedVersion)
		assert.Equal(t, "16.0.0", got[1].UnparsedVersion)
	})
}
