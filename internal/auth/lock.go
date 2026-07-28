package auth

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// lockStaleAfter bounds how long a lock file is honored before it's assumed
// to be left behind by a crashed process rather than an in-progress refresh
// (which involves a couple of network round-trips, nowhere near this long).
const lockStaleAfter = 30 * time.Second

const lockRetryDelay = 50 * time.Millisecond

// Lock acquires an exclusive, cross-process lock around a read-refresh-write
// sequence on auth.yaml, so concurrent CLI invocations don't race the OAuth
// refresh token. Call the returned unlock func to release it.
func Lock(ctx context.Context) (unlock func(), err error) {
	p, err := lockPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	for {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = f.Close()
			return func() { _ = os.Remove(p) }, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create lock %s: %w", p, err)
		}
		if stale, statErr := lockIsStale(p); statErr == nil && stale {
			_ = os.Remove(p)
			continue
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for auth lock: %w", ctx.Err())
		case <-time.After(lockRetryDelay):
		}
	}
}

func lockPath() (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	return p + ".lock", nil
}

func lockIsStale(p string) (bool, error) {
	info, err := os.Stat(p)
	if err != nil {
		return false, err
	}
	return time.Since(info.ModTime()) > lockStaleAfter, nil
}
