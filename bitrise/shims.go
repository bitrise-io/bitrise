package bitrise

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/bitrise-io/bitrise/v3/internal/symlink"
	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/bitrise-io/bitrise/v3/tools"
	"github.com/bitrise-io/go-utils/colorstring"
)

// SetupShims creates or removes the shims. The shim on disk is the opt-in itself, so every
// other setup (e.g. the one after a CLI update) leaves it as it is. Only --clean removes it
// too, as it deletes all of ~/.bitrise.
func SetupShims(enabled bool) error {
	log.Print()
	log.Infof("Checking CLI shims...")

	if !enabled {
		for _, tool := range tools.ShimmedTools() {
			err := os.Remove(shimPath(tool))
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return fmt.Errorf("remove %s shim: %w", tool, err)
			}
			log.Printf("%s %s shim removed", colorstring.Green("[OK]"), tool)
		}
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate the running CLI: %w", err)
	}
	for _, tool := range tools.ShimmedTools() {
		if err := symlink.Ensure(exe, shimPath(tool)); err != nil {
			return fmt.Errorf("create %s shim: %w", tool, err)
		}
		log.Printf("%s %s shim: %s -> %s", colorstring.Green("[OK]"), tool, shimPath(tool), exe)
	}
	return nil
}

// RefreshShims points the existing shims at the running CLI. Its path changes when it is
// reinstalled or upgraded (os.Executable resolves symlinks on Linux), which would leave a
// shim dangling or pointing at an older CLI.
func RefreshShims() error {
	for _, tool := range tools.ShimmedTools() {
		pth := shimPath(tool)
		if _, err := os.Lstat(pth); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}

		// Only looked up once a shim exists, so runs without shims do nothing.
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locate the running CLI: %w", err)
		}
		if err := symlink.Ensure(exe, pth); err != nil {
			return fmt.Errorf("refresh %s shim: %w", tool, err)
		}
	}
	return nil
}

func shimPath(tool string) string {
	return filepath.Join(configs.GetBitriseShimsDirPath(), tool)
}
