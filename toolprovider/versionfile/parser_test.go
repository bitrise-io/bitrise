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

			got, err := parseToolVersionsFile(path)
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

			got, err := parseSingleToolVersion(path)
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
		want     provider.ToolID
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
			got := inferToolID(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some version files
	files := []struct {
		directory string
		filename  string
	}{{"", ".tool-versions"}, {"", ".ruby-version"}, {"", ".node-version"}, {"", ".fvmrc"}, {".fvm", "fvm_config.json"}}
	for _, f := range files {
		fullDirectory := filepath.Join(tmpDir, f.directory)
		if f.directory != "" {
			err := os.MkdirAll(fullDirectory, 0755)
			require.NoError(t, err)
		} else {
			fullDirectory = tmpDir
		}
		err := os.WriteFile(filepath.Join(fullDirectory, f.filename), []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Create a non-version file
	err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644)
	require.NoError(t, err)

	found, err := FindVersionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, found, 5)

	// Check that all expected files are found
	foundMap := make(map[string]bool)
	for _, f := range found {
		foundMap[filepath.Base(f)] = true
	}
	for _, toolFile := range files {
		assert.True(t, foundMap[toolFile.filename], "expected to find %s", toolFile.filename)
	}
}

func TestParseVersionFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		content  string
		want     []ToolVersion
	}{
		{
			name:     ".tool-versions format",
			filePath: ".tool-versions",
			content:  "ruby 3.2.0\nnodejs 18.0.0",
			want: []ToolVersion{
				{ToolName: "ruby", Version: "3.2.0"},
				{ToolName: "nodejs", Version: "18.0.0"},
			},
		},
		{
			name:     "single tool version format",
			filePath: ".ruby-version",
			content:  "3.2.0",
			want:     []ToolVersion{{ToolName: "ruby", Version: "3.2.0"}},
		},
		{
			name:     ".fvmrc format",
			filePath: ".fvmrc",
			content:  `{"flutter": "3.22.0"}`,
			want:     []ToolVersion{{ToolName: "flutter", Version: "3.22.0"}},
		},
		{
			name:     ".nvmrc format",
			filePath: ".nvmrc",
			content:  "v18.0.0",
			want:     []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:     "fvm_config.json format",
			filePath: "fvm_config.json",
			content:  `{"flutterSdkVersion": "3.22.0"}`,
			want:     []ToolVersion{{ToolName: "flutter", Version: "3.22.0"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filePath)
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0644))

			got, err := Parse(path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFVMRC(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ToolVersion
		wantErr bool
	}{
		{
			name:    "exact version",
			content: `{"flutter": "3.19.0"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "3.19.0"}},
		},
		{
			name:    "version with channel suffix",
			content: `{"flutter": "3.19.0@stable"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "3.19.0-stable"}},
		},
		{
			name:    "latest",
			content: `{"flutter": "latest"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "latest"}},
		},
		{
			name:    "channel only is rejected",
			content: `{"flutter": "stable"}`,
			wantErr: true,
		},
		{
			name:    "missing flutter key",
			content: `{"dart": "3.0.0"}`,
			wantErr: true,
		},
		{
			name:    "empty flutter value",
			content: `{"flutter": ""}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			content: `not json`,
			wantErr: true,
		},
		{
			name:    "with flavors",
			content: `{"flutter": "3.19.0", "flavors": {"development": "3.22.0@beta", "production": "3.19.0"}}`,
			want: []ToolVersion{
				{ToolName: "flutter", Version: "3.19.0"},
				{ToolName: "flutter", Version: "3.22.0-beta"},
			},
		},
		{
			name:    "with flavors - channel only is rejected",
			content: `{"flutter": "3.19.0", "flavors": {"development": "beta"}}`,
			wantErr: true,
		},
		{
			name:    "with flavors - deduplicates main version",
			content: `{"flutter": "3.19.0", "flavors": {"production": "3.19.0"}}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "3.19.0"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, ".fvmrc")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := parseFVMRC(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseNVMRC(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ToolVersion
		wantErr bool
	}{
		{
			name:    "version without v prefix",
			content: "18.0.0",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "version with v prefix",
			content: "v18.0.0",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "version with newline",
			content: "18.0.0\n",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "version with v prefix and newline",
			content: "v20.10.0\n",
			want:    []ToolVersion{{ToolName: "node", Version: "20.10.0"}},
		},
		{
			name:    "version with whitespace",
			content: "  v18.0.0  \n",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "with comment lines",
			content: "# Node version\nv18.0.0\n",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "skips environment variable assignments",
			content: "NODE_VERSION=18.0.0\nv20.0.0",
			want:    []ToolVersion{{ToolName: "node", Version: "20.0.0"}},
		},
		{
			name:    "first non-comment line wins",
			content: "# Comment\n18.0.0\n20.0.0",
			want:    []ToolVersion{{ToolName: "node", Version: "18.0.0"}},
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			content: "  \n  \n",
			wantErr: true,
		},
		{
			name:    "only comments",
			content: "# Comment only\n# Another comment",
			wantErr: true,
		},
		{
			name:    "major version only",
			content: "18",
			want:    []ToolVersion{{ToolName: "node", Version: "18"}},
		},
		{
			name:    "lts alias",
			content: "lts/*",
			want:    []ToolVersion{{ToolName: "node", Version: "lts/*"}},
		},
		{
			name:    "only v prefix without version",
			content: "v",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, ".nvmrc")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := parseNVMRC(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFVMConfigJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ToolVersion
		wantErr bool
	}{
		{
			name:    "exact version",
			content: `{"flutterSdkVersion": "3.19.0"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "3.19.0"}},
		},
		{
			name:    "version with channel suffix",
			content: `{"flutterSdkVersion": "3.19.0@stable"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "3.19.0-stable"}},
		},
		{
			name:    "latest",
			content: `{"flutterSdkVersion": "latest"}`,
			want:    []ToolVersion{{ToolName: "flutter", Version: "latest"}},
		},
		{
			name:    "channel only is rejected",
			content: `{"flutterSdkVersion": "stable"}`,
			wantErr: true,
		},
		{
			name:    "missing flutterSdkVersion key",
			content: `{"dart": "3.0.0"}`,
			wantErr: true,
		},
		{
			name:    "empty flutterSdkVersion value",
			content: `{"flutterSdkVersion": ""}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			content: `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "fvm_config.json")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := parseFVMConfigJSON(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
