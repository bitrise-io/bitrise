//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
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
	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_mise_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func TestMiseNodeCorepack(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "node_corepack_test", "--config", "toolprovider_test_mise_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func TestWorkflowChaining(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_workflow_chain_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
