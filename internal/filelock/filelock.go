// Package filelock provides a simple, cross-process exclusive lock backed by
// a sentinel file, for serializing a read-modify-write sequence on some other
// file across separate CLI invocations. Shared by internal/auth (guarding
// auth.yaml against the OAuth refresh race) and internal/config (guarding
// config.yml against concurrent `config set`/`unset`).
package filelock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// retryDelay paces re-checks while waiting for a held lock. Var rather than
// const so tests can compress it.
var retryDelay = 50 * time.Millisecond

// Options configures Lock's staleness handling. Both durations should be
// picked relative to how long the caller's critical section can actually
// run.
type Options struct {
	// StaleAfter bounds how long a lock file is honored without being
	// refreshed. Exceeding it means the holder crashed (or never
	// refreshes, RefreshInterval == 0, and the write it was guarding
	// already finished) — not that the holder is merely slow.
	StaleAfter time.Duration
	// RefreshInterval re-stamps the held lock's mtime this often, so a
	// critical section slower than StaleAfter is never mistaken for a
	// crashed holder. Zero disables refreshing, which is fine as long as
	// StaleAfter comfortably exceeds the slowest realistic hold time.
	RefreshInterval time.Duration
}

// Lock acquires an exclusive, cross-process lock on path+".lock" (creating
// path's parent directory if needed), so a read-modify-write sequence on
// path isn't raced by another process doing the same. Call the returned
// unlock func to release it.
func Lock(ctx context.Context, path string, opts Options) (unlock func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	lockPath := path + ".lock"
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}
	nonce, err := lockNonce()
	if err != nil {
		return nil, err
	}
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			if err := writeLockNonce(f, nonce); err != nil {
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("write lock %s: %w", lockPath, err)
			}
			return newLockLease(lockPath, nonce, opts.RefreshInterval), nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create lock %s: %w", lockPath, err)
		}
		stale, statErr := lockIsStale(lockPath, opts.StaleAfter)
		if statErr != nil {
			if errors.Is(statErr, fs.ErrNotExist) {
				// The lock file vanished between our OpenFile and this stat
				// (the holder released it) — retry immediately.
				continue
			}
			return nil, fmt.Errorf("check lock %s: %w", lockPath, statErr)
		}
		if stale {
			_ = os.Remove(lockPath)
			continue
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for lock %s: %w", lockPath, ctx.Err())
		case <-time.After(retryDelay):
		}
	}
}

// lockNonce identifies this call's hold on the lock, so a lock reclaimed as
// stale is never mistaken for our own (see lockIsOwned).
func lockNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate lock nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func writeLockNonce(f *os.File, nonce string) error {
	if _, err := f.WriteString(nonce); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// newLockLease optionally keeps the held lock's mtime current until unlock
// (refreshInterval > 0), so a critical section slower than StaleAfter is
// never mistaken for a crashed process and stolen mid-flight. Both the
// renewal and the release are gated on still owning the lock, so a hold that
// did lapse and was reclaimed is neither kept alive nor deleted by its
// former holder.
func newLockLease(path, nonce string, refreshInterval time.Duration) (unlock func()) {
	if refreshInterval <= 0 {
		return func() {
			if lockIsOwned(path, nonce) {
				_ = os.Remove(path)
			}
		}
	}
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		t := time.NewTicker(refreshInterval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case now := <-t.C:
				if lockIsOwned(path, nonce) {
					_ = os.Chtimes(path, now, now)
				}
			}
		}
	}()
	return func() {
		close(stop)
		<-stopped
		if lockIsOwned(path, nonce) {
			_ = os.Remove(path)
		}
	}
}

func lockIsOwned(path, nonce string) bool {
	b, err := os.ReadFile(path)
	return err == nil && string(b) == nonce
}

// statFile is a var so tests can simulate stat failures other than
// fs.ErrNotExist (e.g. permission denied) without depending on real,
// platform-specific filesystem permission behavior.
var statFile = os.Stat

func lockIsStale(p string, staleAfter time.Duration) (bool, error) {
	info, err := statFile(p)
	if err != nil {
		return false, err
	}
	return time.Since(info.ModTime()) > staleAfter, nil
}
