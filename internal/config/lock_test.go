package config

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock_AcquireRelease(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unlock, err := Lock(t.Context())
	require.NoError(t, err)
	unlock()

	unlock2, err := Lock(t.Context())
	require.NoError(t, err)
	unlock2()
}

func TestLock_BlocksConcurrentHolder(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unlock, err := Lock(t.Context())
	require.NoError(t, err)
	defer unlock()

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for lock")
}

func TestLock_NilContextDoesNotPanic(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	//nolint:staticcheck // SA1012: deliberately passing a nil context.Context — RunE's cmd.Context() is nil in every test that calls RunE directly instead of Execute(), which Lock must tolerate
	unlock, err := Lock(nil)
	require.NoError(t, err)
	unlock()
}
