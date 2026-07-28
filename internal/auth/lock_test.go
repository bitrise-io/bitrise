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
