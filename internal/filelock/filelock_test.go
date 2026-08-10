package filelock

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

func testOptions() Options {
	return Options{StaleAfter: 30 * time.Second, RefreshInterval: 5 * time.Second}
}

func TestLock_ExcludesSecondAcquirer(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")

	unlock, err := Lock(context.Background(), p, testOptions())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx, p, testOptions())
	assert.Error(t, err, "a second acquirer should not get the lock while it's held")

	unlock()

	unlock2, err := Lock(context.Background(), p, testOptions())
	require.NoError(t, err, "lock should be acquirable again after release")
	unlock2()
}

func TestLock_ReclaimsStaleLock(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")
	opts := testOptions()

	lockPath := p + ".lock"
	require.NoError(t, os.WriteFile(lockPath, nil, 0o600))
	old := time.Now().Add(-opts.StaleAfter - time.Second)
	require.NoError(t, os.Chtimes(lockPath, old, old))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	unlock, err := Lock(ctx, p, opts)
	require.NoError(t, err, "a stale lock should be reclaimed rather than blocking forever")
	unlock()
}

func TestLock_PropagatesStatErrorOtherThanNotExist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")
	lockPath := p + ".lock"
	require.NoError(t, os.WriteFile(lockPath, nil, 0o600))

	wantErr := errors.New("permission denied")
	orig := statFile
	statFile = func(name string) (os.FileInfo, error) { return nil, wantErr }
	defer func() { statFile = orig }()

	_, err := Lock(context.Background(), p, testOptions())
	require.Error(t, err, "a real stat failure must not be silently retried forever")
	assert.ErrorIs(t, err, wantErr)
}

func TestLock_RetriesImmediatelyWhenLockFileVanishes(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")
	lockPath := p + ".lock"
	require.NoError(t, os.WriteFile(lockPath, nil, 0o600))

	// Simulate another process releasing the lock in the window between our
	// OpenFile (which saw it exist) and this stat call: the file disappears
	// out from under us, and lockIsStale surfaces that as fs.ErrNotExist.
	orig := statFile
	statFile = func(name string) (os.FileInfo, error) {
		_ = os.Remove(lockPath)
		return nil, fs.ErrNotExist
	}
	defer func() { statFile = orig }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	unlock, err := Lock(ctx, p, testOptions())
	require.NoError(t, err, "should retry once the file is actually gone, not error out or hang")
	unlock()
}

func TestLock_CtxTimeoutWhileWaiting(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")

	unlock, err := Lock(context.Background(), p, testOptions())
	require.NoError(t, err)
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx, p, testOptions())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for lock")
}

// TestLock_HeldLockSurvivesTheStaleWindow covers the case a plain mtime check
// gets wrong: a holder can legitimately keep the lock across a critical
// section longer than StaleAfter as long as RefreshInterval keeps re-stamping
// it, and a waiter must not then reclaim the lock mid-hold.
func TestLock_HeldLockSurvivesTheStaleWindow(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")
	opts := Options{StaleAfter: 100 * time.Millisecond, RefreshInterval: 10 * time.Millisecond}

	unlock, err := Lock(context.Background(), p, opts)
	require.NoError(t, err)
	defer unlock()

	time.Sleep(5 * opts.StaleAfter)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = Lock(ctx, p, opts)
	assert.Error(t, err, "a still-held lock must not be reclaimed just because the hold outlived StaleAfter")
}

func TestLock_NoRefreshDoesNotSurviveTheStaleWindow(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")
	opts := Options{StaleAfter: 50 * time.Millisecond} // RefreshInterval: 0

	unlock, err := Lock(context.Background(), p, opts)
	require.NoError(t, err)
	defer unlock()

	time.Sleep(5 * opts.StaleAfter)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	unlock2, err := Lock(ctx, p, opts)
	require.NoError(t, err, "with RefreshInterval 0 the lock is expected to go stale and be reclaimed once StaleAfter elapses")
	unlock2()
}

func TestLock_UnlockLeavesAReclaimedLockAlone(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")

	unlock, err := Lock(context.Background(), p, testOptions())
	require.NoError(t, err)

	// Stand in for another process reclaiming the lock as stale and taking it:
	// same path, different owner.
	lockPath := p + ".lock"
	require.NoError(t, os.WriteFile(lockPath, []byte("another-process"), 0o600))

	unlock()

	got, err := os.ReadFile(lockPath)
	require.NoError(t, err, "unlock must not delete a lock held by someone else")
	assert.Equal(t, "another-process", string(got))
}

func TestLock_NilContextDoesNotPanic(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource")

	//nolint:staticcheck // SA1012: deliberately passing a nil context.Context — callers' cmd.Context() can be nil in tests that invoke RunE directly instead of Execute()
	unlock, err := Lock(nil, p, testOptions())
	require.NoError(t, err)
	unlock()
}
