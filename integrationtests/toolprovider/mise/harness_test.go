//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// isolatedEnvs keeps a mise provider's mise away from state the parallel tests would share:
//   - its own cache dir: with the shared one, `mise latest --installed` came back empty in CI
//   - no config files from the parent dirs of the working dir: mise looks for them up to the
//     root, and the macOS stacks keep a ~/.tool-versions above the checkout. Its ruby made
//     every test set up ruby-build, which garbled the output of the concurrent mise commands.
func isolatedEnvs(t *testing.T) map[string]string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	// mise compares the ceiling with the resolved working dir, e.g. /private/var on macOS.
	wd, err = filepath.EvalSymlinks(wd)
	require.NoError(t, err)

	return map[string]string{
		"MISE_CACHE_DIR":     t.TempDir(),
		"MISE_CEILING_PATHS": wd,
	}
}
