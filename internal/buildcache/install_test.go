package buildcache

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVersion = "3.14.1"

func TestLink(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "tools", binaryName)

	require.NoError(t, link("/v1/bin", name))
	require.NoError(t, link("/v1/bin", name))
	got, err := os.Readlink(name)
	require.NoError(t, err)
	assert.Equal(t, "/v1/bin", got)

	require.NoError(t, link("/v2/bin", name))
	got, err = os.Readlink(name)
	require.NoError(t, err)
	assert.Equal(t, "/v2/bin", got)

	entries, err := os.ReadDir(filepath.Dir(name))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func testLogger() log.Logger {
	return log.NewLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Writer: io.Discard})
}
