//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestStepBundleRunIf(t *testing.T) {
	configPth := "step_bundle_run_if_test_bitrise.yml"

	cmd := command.New(testhelpers.BinPath(), "run", "--output-format", "json", "test", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	stepOutputs := testhelpers.CollectStepOutputs(out, t)
	require.Equal(t, stepOutputs, []string{
		"script\n",
		"run_if_test_1.script\n",
		"run_if_test_3.script\n",
	})
}

// TestStepBundleRunIfEncapsulation verifies that a Step Bundle's run_if is evaluated once, on Bundle
// entry, and not re-evaluated for later Steps. The Bundle's first Step flips FLAG (the variable the
// run_if reads), but the second Step must still run because the decision was made at entry. See BIVS-3383.
func TestStepBundleRunIfEncapsulation(t *testing.T) {
	configPth := "step_bundle_run_if_test_bitrise.yml"

	cmd := command.New(testhelpers.BinPath(), "run", "--output-format", "json", "encapsulation", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	stepOutputs := testhelpers.CollectStepOutputs(out, t)
	require.Equal(t, stepOutputs, []string{
		"encapsulation.second\n",
	})
}
