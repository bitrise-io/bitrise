//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// Disabled: on macOS the asdf provider reports the requested version as
// installed, but the step that follows still finds a different `node` on PATH,
// so the tool is never activated. The same workflow passes on Linux. Re-enable
// once asdf activation reaches the step environment.
// Failing build: https://app.bitrise.io/app/665682b1-d109-4e81-b9ca-16f5e95e3403/build/b2e30dc0-8951-4e59-b789-720fcee0a8cf
// func TestAsdfToolProvider(t *testing.T) {
// 	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_asdf_bitrise.yml")
// 	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
// 	require.NoError(t, err, out)
// }

func TestMiseToolProvider(t *testing.T) {
	t.Parallel()

	out, err := runBitrise(t, "", nil, "run", "toolprovider_test", "--config", "toolprovider_test_mise_bitrise.yml")
	require.NoError(t, err, out)
}

func TestMiseNodeCorepack(t *testing.T) {
	t.Parallel()

	out, err := runBitrise(t, "", nil, "run", "node_corepack_test", "--config", "toolprovider_test_mise_bitrise.yml")
	require.NoError(t, err, out)
}

func TestWorkflowChaining(t *testing.T) {
	t.Parallel()

	out, err := runBitrise(t, "", nil, "run", "toolprovider_test", "--config", "toolprovider_test_workflow_chain_bitrise.yml")
	require.NoError(t, err, out)
}

// runBitrise runs bitrise with args in dir, the package dir when empty, and runs it once more
// if mise failed rebuilding its runtime symlinks. These tests install tools into the machine's
// own mise data dir in parallel, and after an install, mise rebuilds the runtime symlinks
// (installs/<tool>/latest and the like) of every installed tool without a lock, so two
// installs can remove the same link. The version is installed by then, and the second run only
// rebuilds the links. A build runs one bitrise at a time and never hits this.
func runBitrise(t *testing.T, dir string, envs []string, args ...string) (string, error) {
	run := func() (string, error) {
		return command.New(testhelpers.BinPath(), args...).
			SetDir(dir).
			AppendEnvs(envs...).
			RunAndReturnTrimmedCombinedOutput()
	}

	out, err := run()
	if err != nil && strings.Contains(out, "failed to rebuild runtime symlinks") {
		t.Log("mise raced another install rebuilding its runtime symlinks, running bitrise again")
		out, err = run()
	}
	return out, err
}
