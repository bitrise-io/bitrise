package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsure(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "tools", "tool")

	require.NoError(t, Ensure("/v1/bin", name))
	require.NoError(t, Ensure("/v1/bin", name))
	got, err := os.Readlink(name)
	require.NoError(t, err)
	assert.Equal(t, "/v1/bin", got)

	require.NoError(t, Ensure("/v2/bin", name))
	got, err = os.Readlink(name)
	require.NoError(t, err)
	assert.Equal(t, "/v2/bin", got)

	entries, err := os.ReadDir(filepath.Dir(name))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}
