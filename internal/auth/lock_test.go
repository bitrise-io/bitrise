package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock_ExcludesSecondAcquirer(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unlock, err := Lock(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx)
	assert.Error(t, err, "a second acquirer should not get the lock while it's held")

	unlock()

	unlock2, err := Lock(context.Background())
	require.NoError(t, err, "lock should be acquirable again after release")
	unlock2()
}

func TestLock_CtxTimeoutWhileWaiting(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unlock, err := Lock(context.Background())
	require.NoError(t, err)
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for lock")
}

// TestLock_HeldLockSurvivesTheStaleWindow covers the case a plain mtime check
// gets wrong: the OAuth ladder can hold the lock across three sequential token
// requests, longer than lockStaleAfter, and a waiter must not then reclaim the
// lock and spend the same refresh token. The general staleness/reclaim/nonce
// mechanics live in internal/filelock's own tests; this just confirms auth
// wires its (staleAfter, refreshInterval) pair correctly.
func TestLock_HeldLockSurvivesTheStaleWindow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	compressLockTimings(t, 100*time.Millisecond, 10*time.Millisecond)

	unlock, err := Lock(context.Background())
	require.NoError(t, err)
	defer unlock()

	time.Sleep(5 * lockStaleAfter)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx)
	assert.Error(t, err, "a still-held lock must not be reclaimed just because the hold outlived lockStaleAfter")
}

func compressLockTimings(t *testing.T, staleAfter, refreshInterval time.Duration) {
	t.Helper()
	origStale, origRefresh := lockStaleAfter, lockRefreshInterval
	lockStaleAfter, lockRefreshInterval = staleAfter, refreshInterval
	t.Cleanup(func() { lockStaleAfter, lockRefreshInterval = origStale, origRefresh })
}
