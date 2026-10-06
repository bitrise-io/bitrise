package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectAndClearExports_ClearsAHalfWrittenStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "envstore.yml")
	require.NoError(t, os.WriteFile(path, []byte("envs:\n- KEY: [unterminated"), 0o600))
	var cleared []string

	exported, err := collectAndClearExports(path, func(p string) error {
		cleared = append(cleared, p)

		return nil
	})

	require.Error(t, err)
	assert.Empty(t, exported)
	assert.Equal(t, []string{path}, cleared, "the first step would otherwise fail reading the same file")
}

func TestCollectAndClearExports_ReturnsTheExportsAndClears(t *testing.T) {
	path := filepath.Join(t.TempDir(), "envstore.yml")
	require.NoError(t, os.WriteFile(path, []byte("envs:\n- PATH_ADDITION: /x\n"), 0o600))
	cleared := 0

	exported, err := collectAndClearExports(path, func(string) error {
		cleared++

		return nil
	})

	require.NoError(t, err)
	require.Len(t, exported, 1)
	assert.Equal(t, 1, cleared)
}
