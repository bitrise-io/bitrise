//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// isolatedEnvs keeps a mise provider's mise away from state that parallel tests would share:
//   - its own cache dir: with a shared one, `mise latest --installed` came back empty in CI.
//   - no config files from the dirs above the test package: mise reads config files from the
//     working dir and from every dir above it, and the macOS stacks keep a ~/.tool-versions
//     above the checkout. The ruby in it made every test set up ruby-build at the same time,
//     and their errors ended up in the output of other mise commands.
func isolatedEnvs(t *testing.T) map[string]string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	// mise resolves symlinks in the working dir before it compares it with the ceiling, so a
	// ceiling with a symlink in it never matches and mise ignores it. On macOS, for example,
	// /var is a symlink to /private/var.
	wd, err = filepath.EvalSymlinks(wd)
	require.NoError(t, err)

	return map[string]string{
		"MISE_CACHE_DIR": t.TempDir(),
		// mise stops looking for config files at this dir instead of going up to the root.
		"MISE_CEILING_PATHS": wd,
	}
}
