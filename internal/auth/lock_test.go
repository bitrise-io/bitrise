package auth

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

func TestLock_ReclaimsStaleLock(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	p, err := lockPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
	require.NoError(t, os.WriteFile(p, nil, 0o600))
	old := time.Now().Add(-lockStaleAfter - time.Second)
	require.NoError(t, os.Chtimes(p, old, old))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	unlock, err := Lock(ctx)
	require.NoError(t, err, "a stale lock should be reclaimed rather than blocking forever")
	unlock()
}

func TestLock_PropagatesStatErrorOtherThanNotExist(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	p, err := lockPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
	require.NoError(t, os.WriteFile(p, nil, 0o600))

	wantErr := errors.New("permission denied")
	orig := statLockFile
	statLockFile = func(name string) (os.FileInfo, error) { return nil, wantErr }
	defer func() { statLockFile = orig }()

	_, err = Lock(context.Background())
	require.Error(t, err, "a real stat failure must not be silently retried forever")
	assert.ErrorIs(t, err, wantErr)
}

func TestLock_RetriesImmediatelyWhenLockFileVanishes(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	p, err := lockPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
	require.NoError(t, os.WriteFile(p, nil, 0o600))

	// Simulate another process releasing the lock in the window between our
	// OpenFile (which saw it exist) and this stat call: the file disappears
	// out from under us, and lockIsStale surfaces that as fs.ErrNotExist.
	orig := statLockFile
	statLockFile = func(name string) (os.FileInfo, error) {
		_ = os.Remove(p)
		return nil, fs.ErrNotExist
	}
	defer func() { statLockFile = orig }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	unlock, err := Lock(ctx)
	require.NoError(t, err, "should retry once the file is actually gone, not error out or hang")
	unlock()
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
	assert.Contains(t, err.Error(), "timed out waiting for auth lock")
}

// TestLock_HeldLockSurvivesTheStaleWindow covers the case a plain mtime check
// gets wrong: the OAuth ladder can hold the lock across three sequential token
// requests, longer than lockStaleAfter, and a waiter must not then reclaim the
// lock and spend the same refresh token.
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

func TestLock_UnlockLeavesAReclaimedLockAlone(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unlock, err := Lock(context.Background())
	require.NoError(t, err)

	// Stand in for another process reclaiming the lock as stale and taking it:
	// same path, different owner.
	p, err := lockPath()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(p, []byte("another-process"), 0o600))

	unlock()

	got, err := os.ReadFile(p)
	require.NoError(t, err, "unlock must not delete a lock held by someone else")
	assert.Equal(t, "another-process", string(got))
}

func compressLockTimings(t *testing.T, staleAfter, refreshInterval time.Duration) {
	t.Helper()
	origStale, origRefresh := lockStaleAfter, lockRefreshInterval
	lockStaleAfter, lockRefreshInterval = staleAfter, refreshInterval
	t.Cleanup(func() { lockStaleAfter, lockRefreshInterval = origStale, origRefresh })
}
