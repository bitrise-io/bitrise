//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestAsdfToolProvider(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_asdf_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

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
