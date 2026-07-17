//go:build darwin || freebsd || netbsd || openbsd

package fileutil

import (
	"syscall"
	"time"
)

func atimeFromStat(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
}
