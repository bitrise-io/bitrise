package versionfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestPackageJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

func TestParsePackageJSON_EnginesNode(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"name": "my-app",
		"engines": {
			"node": "^20.0.0"
		}
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, ToolVersion{ToolName: "nodejs", Version: "^20.0.0", IsConstraint: true}, tools[0])
}

func TestParsePackageJSON_EnginesMultiple(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"engines": {
			"node": ">=18",
			"npm": ">=9",
			"yarn": "^4.0.0",
			"pnpm": "~8.0.0"
		}
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 4)

	// node first, then npm, pnpm, yarn (alphabetical for package managers)
	assert.Equal(t, "nodejs", string(tools[0].ToolName))
	assert.Equal(t, ">=18", tools[0].Version)
	assert.True(t, tools[0].IsConstraint)

	assert.Equal(t, "npm", string(tools[1].ToolName))
	assert.Equal(t, ">=9", tools[1].Version)
	assert.True(t, tools[1].IsConstraint)

	assert.Equal(t, "pnpm", string(tools[2].ToolName))
	assert.Equal(t, "~8.0.0", tools[2].Version)
	assert.True(t, tools[2].IsConstraint)

	assert.Equal(t, "yarn", string(tools[3].ToolName))
	assert.Equal(t, "^4.0.0", tools[3].Version)
	assert.True(t, tools[3].IsConstraint)
}

func TestParsePackageJSON_PackageManager(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"packageManager": "yarn@4.0.0"
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, ToolVersion{ToolName: "yarn", Version: "4.0.0", IsConstraint: false}, tools[0])
}

func TestParsePackageJSON_PackageManagerWithHash(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"packageManager": "pnpm@8.15.4+sha256.abc123"
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "pnpm", string(tools[0].ToolName))
	assert.Equal(t, "8.15.4", tools[0].Version)
	assert.False(t, tools[0].IsConstraint)
}

func TestParsePackageJSON_PackageManagerOverridesEngines(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"engines": {
			"node": "^20.0.0",
			"yarn": ">=3.0.0"
		},
		"packageManager": "yarn@4.0.0"
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 2)

	// node from engines
	assert.Equal(t, "nodejs", string(tools[0].ToolName))
	assert.Equal(t, "^20.0.0", tools[0].Version)
	assert.True(t, tools[0].IsConstraint)

	// yarn from packageManager (overrides engines)
	assert.Equal(t, "yarn", string(tools[1].ToolName))
	assert.Equal(t, "4.0.0", tools[1].Version)
	assert.False(t, tools[1].IsConstraint)
}

func TestParsePackageJSON_NoToolFields(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"name": "my-app",
		"version": "1.0.0"
	}`)

	_, err := Parse(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tool version requirements found")
}

func TestParsePackageJSON_InvalidJSON(t *testing.T) {
	path := writeTestPackageJSON(t, `not json`)

	_, err := Parse(path)
	assert.Error(t, err)
}

func TestParsePackageJSON_EnginesNotObject(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"engines": "not an object"
	}`)

	_, err := Parse(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "engines")
}

func TestParsePackageJSON_InvalidPackageManager(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"no at sign", `{"packageManager": "yarn4.0.0"}`},
		{"empty name", `{"packageManager": "@4.0.0"}`},
		{"unsupported name", `{"packageManager": "bun@1.0.0"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTestPackageJSON(t, tt.value)
			_, err := Parse(path)
			assert.Error(t, err)
		})
	}
}

func TestParsePackageJSON_UnsupportedEnginesIgnored(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"engines": {
			"node": "^20.0.0",
			"vscode": "^1.80.0"
		}
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "nodejs", string(tools[0].ToolName))
}

func TestParsePackageJSON_NpmPackageManager(t *testing.T) {
	path := writeTestPackageJSON(t, `{
		"packageManager": "npm@10.0.0"
	}`)

	tools, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "npm", string(tools[0].ToolName))
	assert.Equal(t, "10.0.0", tools[0].Version)
}

func TestParsePackageJSON_FindVersionFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	err := os.WriteFile(path, []byte(`{"engines":{"node":"20"}}`), 0644)
	require.NoError(t, err)

	files, err := FindVersionFiles(dir)
	require.NoError(t, err)
	assert.Contains(t, files, path)
}
