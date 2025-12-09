//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/versionfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionFileIntegration(t *testing.T) {
	t.Run("parse real .tool-versions file", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		content := `# Development tools
ruby 3.2.0
nodejs 20.0.0
golang 1.21.0
python 3.11.0`

		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		tools, err := versionfile.ParseVersionFile(toolVersionsPath)
		require.NoError(t, err)
		require.Len(t, tools, 4)

		assert.Equal(t, "ruby", string(tools[0].ToolName))
		assert.Equal(t, "3.2.0", tools[0].Version)
		assert.Equal(t, "nodejs", string(tools[1].ToolName))
		assert.Equal(t, "20.0.0", tools[1].Version)
		assert.Equal(t, "golang", string(tools[2].ToolName))
		assert.Equal(t, "1.21.0", tools[2].Version)
		assert.Equal(t, "python", string(tools[3].ToolName))
		assert.Equal(t, "3.11.0", tools[3].Version)
	})

	t.Run("parse .tool-versions with many tools", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		// Test with 7+ tools to ensure no arbitrary limits.
		content := `ruby 3.2.0
nodejs 20.0.0
golang 1.21.0
python 3.11.0
java openjdk-11
terraform 1.5.0
kubectl 1.28.0
rust 1.70.0`

		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		tools, err := versionfile.ParseVersionFile(toolVersionsPath)
		require.NoError(t, err)
		require.Len(t, tools, 8)

		// Verify all tools are parsed correctly.
		expected := []struct {
			name    string
			version string
		}{
			{"ruby", "3.2.0"},
			{"nodejs", "20.0.0"},
			{"golang", "1.21.0"},
			{"python", "3.11.0"},
			{"java", "openjdk-11"},
			{"terraform", "1.5.0"},
			{"kubectl", "1.28.0"},
			{"rust", "1.70.0"},
		}

		for i, exp := range expected {
			assert.Equal(t, exp.name, string(tools[i].ToolName))
			assert.Equal(t, exp.version, tools[i].Version)
		}
	})

	t.Run("find multiple version files in directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := map[string]string{
			".tool-versions":  "ruby 3.2.0\nnodejs 20.0.0",
			".ruby-version":   "3.2.1",
			".python-version": "3.11.0",
			".node-version":   "18.0.0",
		}

		for filename, content := range files {
			path := filepath.Join(tmpDir, filename)
			err := os.WriteFile(path, []byte(content), 0644)
			require.NoError(t, err)
		}

		// Create other files too
		err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)
		require.NoError(t, err)

		foundFiles, err := versionfile.FindVersionFiles(tmpDir)
		require.NoError(t, err)
		assert.Len(t, foundFiles, 4)

		foundMap := make(map[string]bool)
		for _, f := range foundFiles {
			foundMap[filepath.Base(f)] = true
		}

		for expectedFile := range files {
			assert.True(t, foundMap[expectedFile], "expected to find %s", expectedFile)
		}
	})

	t.Run("parse multiple version files with overlapping tools", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .tool-versions with ruby and nodejs.
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")
		err := os.WriteFile(toolVersionsPath, []byte("ruby 3.2.0\nnodejs 20.0.0"), 0644)
		require.NoError(t, err)

		// Create .ruby-version with different version.
		rubyVersionPath := filepath.Join(tmpDir, ".ruby-version")
		err = os.WriteFile(rubyVersionPath, []byte("3.2.1"), 0644)
		require.NoError(t, err)

		// Create .node-version with different version.
		nodeVersionPath := filepath.Join(tmpDir, ".node-version")
		err = os.WriteFile(nodeVersionPath, []byte("18.0.0"), 0644)
		require.NoError(t, err)

		// Parse all files and check for duplicates.
		foundFiles, err := versionfile.FindVersionFiles(tmpDir)
		require.NoError(t, err)

		var allTools []versionfile.ToolVersion
		for _, file := range foundFiles {
			tools, err := versionfile.ParseVersionFile(file)
			require.NoError(t, err)
			allTools = append(allTools, tools...)
		}

		require.Len(t, allTools, 4)

		toolCounts := make(map[string]int)
		for _, tool := range allTools {
			toolCounts[string(tool.ToolName)]++
		}

		assert.Equal(t, 2, toolCounts["ruby"], "ruby should appear twice")
		assert.Equal(t, 2, toolCounts["nodejs"], "nodejs should appear twice")
	})

	t.Run("parse version files with special characters", func(t *testing.T) {
		tmpDir := t.TempDir()

		testCases := []struct {
			filename string
			content  string
			wantTool string
			wantVer  string
		}{
			{
				filename: ".ruby-version",
				content:  "3.2.0-preview1",
				wantTool: "ruby",
				wantVer:  "3.2.0-preview1",
			},
			{
				filename: ".java-version",
				content:  "openjdk-11.0.2",
				wantTool: "java",
				wantVer:  "openjdk-11.0.2",
			},
			{
				filename: ".node-version",
				content:  "v20.0.0",
				wantTool: "nodejs",
				wantVer:  "v20.0.0",
			},
		}

		for _, tc := range testCases {
			path := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(path, []byte(tc.content), 0644)
			require.NoError(t, err)

			tools, err := versionfile.ParseVersionFile(path)
			require.NoError(t, err)
			require.Len(t, tools, 1)

			assert.Equal(t, tc.wantTool, string(tools[0].ToolName))
			assert.Equal(t, tc.wantVer, tools[0].Version)
		}
	})

	t.Run("malformed .tool-versions file", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")

		// Line with only tool name, no version.
		content := "ruby 3.2.0\nnodejs\npython 3.11.0"
		err := os.WriteFile(toolVersionsPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = versionfile.ParseVersionFile(toolVersionsPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})

	t.Run("empty individual version file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rubyVersionPath := filepath.Join(tmpDir, ".ruby-version")

		err := os.WriteFile(rubyVersionPath, []byte(""), 0644)
		require.NoError(t, err)

		_, err = versionfile.ParseVersionFile(rubyVersionPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("version file with only whitespace", func(t *testing.T) {
		tmpDir := t.TempDir()
		nodeVersionPath := filepath.Join(tmpDir, ".node-version")

		err := os.WriteFile(nodeVersionPath, []byte("   \n\t\n   "), 0644)
		require.NoError(t, err)

		_, err = versionfile.ParseVersionFile(nodeVersionPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}
