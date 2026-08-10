package config

import (
	"context"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/filelock"
)

// lockOptions sizes the lease for config.yml's read-modify-write sequence,
// which is always local disk I/O with no network calls — much faster than
// internal/auth's OAuth ladder, so no refresh is needed to survive a
// legitimately slow hold.
var lockOptions = filelock.Options{StaleAfter: 5 * time.Second}

// Lock acquires an exclusive, cross-process lock around a read-modify-write
// sequence on the global config file, so concurrent `bitrise config
// set`/`unset` invocations don't silently drop one of the two writes. Call
// the returned unlock func to release it.
func Lock(ctx context.Context) (unlock func(), err error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	return filelock.Lock(ctx, p, lockOptions)
}
