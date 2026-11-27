// Package symlink maintains the links the CLI keeps on PATH. Depends on nothing
// but the standard library, so any package can import it.
package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Ensure points name at target. The link is swapped in with a rename, so a
// concurrent caller or a process resolving name never sees it missing.
func Ensure(target, name string) error {
	if current, err := os.Readlink(name); err == nil && current == target {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp-%d-%d", name, os.Getpid(), time.Now().UnixNano())
	if err := os.Symlink(target, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, name); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
