package auth

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

// lockStaleAfter bounds how long a lock file is honored after its holder last
// renewed it. A live holder re-stamps the mtime every lockRefreshInterval (see
// newLockLease), so exceeding this means the holder died — not that its refresh
// is slow, which the OAuth ladder's network calls can legitimately be.
//
// Vars rather than consts so tests can compress both timings.
var (
	lockStaleAfter      = 30 * time.Second
	lockRefreshInterval = 5 * time.Second
)

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
	nonce, err := lockNonce()
	if err != nil {
		return nil, err
	}
	for {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			if err := writeLockNonce(f, nonce); err != nil {
				_ = os.Remove(p)
				return nil, fmt.Errorf("write lock %s: %w", p, err)
			}
			return newLockLease(p, nonce), nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create lock %s: %w", p, err)
		}
		stale, statErr := lockIsStale(p)
		if statErr != nil {
			if errors.Is(statErr, fs.ErrNotExist) {
				// The lock file vanished between our OpenFile and this stat
				// (the holder released it) — retry immediately.
				continue
			}
			return nil, fmt.Errorf("check lock %s: %w", p, statErr)
		}
		if stale {
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

// lockNonce identifies this process's hold on the lock, so a lock reclaimed as
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

// newLockLease keeps the held lock's mtime current until unlock, so a refresh
// that is slow (three sequential token requests, each with its own timeout) is
// never mistaken for a crashed process and stolen mid-flight. Both the renewal
// and the release are gated on still owning the lock, so a hold that did lapse
// and was reclaimed is neither kept alive nor deleted by its former holder.
func newLockLease(path, nonce string) (unlock func()) {
	interval := lockRefreshInterval
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		t := time.NewTicker(interval)
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

// statLockFile is a var so tests can simulate stat failures other than
// fs.ErrNotExist (e.g. permission denied) without depending on real,
// platform-specific filesystem permission behavior.
var statLockFile = os.Stat

func lockIsStale(p string) (bool, error) {
	info, err := statLockFile(p)
	if err != nil {
		return false, err
	}
	return time.Since(info.ModTime()) > lockStaleAfter, nil
}
