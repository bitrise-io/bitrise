package versionfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolVersions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ToolVersion
		wantErr bool
	}{
		{
			name: "valid .tool-versions",
			content: `ruby 3.2.0
nodejs 18.0.0
java openjdk-11`,
			want: []ToolVersion{
				{ToolName: "ruby", Version: "3.2.0"},
				{ToolName: "nodejs", Version: "18.0.0"},
				{ToolName: "java", Version: "openjdk-11"},
			},
			wantErr: false,
		},
		{
			name: "with comments and empty lines",
			content: `# This is a comment
ruby 3.2.0

# Another comment
nodejs 18.0.0
`,
			want: []ToolVersion{
				{ToolName: "ruby", Version: "3.2.0"},
				{ToolName: "nodejs", Version: "18.0.0"},
			},
			wantErr: false,
		},
		{
			name: "with extra whitespace",
			content: `  ruby   3.2.0  
  nodejs   18.0.0  `,
			want: []ToolVersion{
				{ToolName: "ruby", Version: "3.2.0"},
				{ToolName: "nodejs", Version: "18.0.0"},
			},
			wantErr: false,
		},
		{
			name:    "invalid format - missing version",
			content: `ruby`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty file",
			content: ``,
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, ".tool-versions")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := ParseToolVersions(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSingleToolVersion(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		want     ToolVersion
		wantErr  bool
	}{
		{
			name:     "ruby version",
			filename: ".ruby-version",
			content:  "3.2.0\n",
			want:     ToolVersion{ToolName: "ruby", Version: "3.2.0"},
			wantErr:  false,
		},
		{
			name:     "node version",
			filename: ".node-version",
			content:  "18.0.0",
			want:     ToolVersion{ToolName: "nodejs", Version: "18.0.0"},
			wantErr:  false,
		},
		{
			name:     "java version",
			filename: ".java-version",
			content:  "openjdk-11",
			want:     ToolVersion{ToolName: "java", Version: "openjdk-11"},
			wantErr:  false,
		},
		{
			name:     "python version with whitespace",
			filename: ".python-version",
			content:  "  3.11.0  \n",
			want:     ToolVersion{ToolName: "python", Version: "3.11.0"},
			wantErr:  false,
		},
		{
			name:     "go version",
			filename: ".go-version",
			content:  "1.21.0",
			want:     ToolVersion{ToolName: "golang", Version: "1.21.0"},
			wantErr:  false,
		},
		{
			name:     "empty version file",
			filename: ".ruby-version",
			content:  "",
			want:     ToolVersion{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := ParseSingleToolVersion(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInferToolName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{".ruby-version", "ruby"},
		{".node-version", "nodejs"},
		{".go-version", "golang"},
		{".java-version", "java"},
		{".python-version", "python"},
		{".terraform-version", "terraform"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := inferToolName(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some version files
	files := []string{".tool-versions", ".ruby-version", ".node-version"}
	for _, f := range files {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Create a non-version file
	err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)
	require.NoError(t, err)

	found, err := FindVersionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, found, 3)

	// Check that all expected files are found
	foundMap := make(map[string]bool)
	for _, f := range found {
		foundMap[filepath.Base(f)] = true
	}
	for _, expectedFile := range files {
		assert.True(t, foundMap[expectedFile], "expected to find %s", expectedFile)
	}
}

func TestParseVersionFile(t *testing.T) {
	t.Run(".tool-versions format", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, ".tool-versions")
		content := `ruby 3.2.0
nodejs 18.0.0`
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)

		tools, err := ParseVersionFile(path)
		require.NoError(t, err)
		assert.Len(t, tools, 2)
		assert.Equal(t, provider.ToolID("ruby"), tools[0].ToolName)
		assert.Equal(t, "3.2.0", tools[0].Version)
	})

	t.Run("single tool version format", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, ".ruby-version")
		err := os.WriteFile(path, []byte("3.2.0"), 0644)
		require.NoError(t, err)

		tools, err := ParseVersionFile(path)
		require.NoError(t, err)
		assert.Len(t, tools, 1)
		assert.Equal(t, provider.ToolID("ruby"), tools[0].ToolName)
		assert.Equal(t, "3.2.0", tools[0].Version)
	})
}
