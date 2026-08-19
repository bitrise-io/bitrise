//go:build linux

package fileutil

import (
	"syscall"
	"time"
)

func atimeFromStat(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}
