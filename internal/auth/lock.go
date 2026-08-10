package auth

import (
	"context"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/filelock"
)

// lockStaleAfter/lockRefreshInterval size the lease for the OAuth ladder's
// read-refresh-write sequence, which can legitimately run across three
// sequential token requests, each with its own timeout — much slower than
// internal/config's lock, which is why this can't share a single fixed
// Options value with it.
//
// Vars rather than consts so tests can compress both timings.
var (
	lockStaleAfter      = 30 * time.Second
	lockRefreshInterval = 5 * time.Second
)

// Lock acquires an exclusive, cross-process lock around a read-refresh-write
// sequence on auth.yaml, so concurrent CLI invocations don't race the OAuth
// refresh token. Call the returned unlock func to release it.
func Lock(ctx context.Context) (unlock func(), err error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	return filelock.Lock(ctx, p, filelock.Options{StaleAfter: lockStaleAfter, RefreshInterval: lockRefreshInterval})
}
