//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/toolprovider/mise"
	"github.com/stretchr/testify/require"
)

// bootstrappedMiseInstallDir has mise installed before any test runs, and the tests share it
// instead of each bootstrapping its own. Bootstrapping writes bin/mise, and a process that a
// parallel test forks meanwhile inherits the open file until it execs: running that mise then
// fails with "text file busy" (https://go.dev/issue/22315). A build bootstraps mise once.
var bootstrappedMiseInstallDir string

func TestMain(m *testing.M) {
	installDir, err := os.MkdirTemp("", "mise-install-")
	if err == nil {
		var miseProvider *mise.MiseToolProvider
		miseProvider, err = mise.NewToolProvider(installDir, filepath.Join(installDir, "data"), false, true, nil)
		if err == nil {
			err = miseProvider.Bootstrap()
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap mise: %s\n", err)
		os.Exit(1)
	}
	bootstrappedMiseInstallDir = installDir

	code := m.Run()
	_ = os.RemoveAll(installDir)
	os.Exit(code)
}

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
