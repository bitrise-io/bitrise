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

// TestStepBundleRunIfEncapsulation verifies that each Step Bundle's run_if is evaluated once, on
// Bundle entry, and not re-evaluated for the Bundle's later Steps, at both nesting levels. The outer
// Bundle contains an inner Bundle followed by an echo Step. The inner Bundle's first Step flips FLAG
// (the variable both run_ifs read) to false; despite that, the inner Bundle's echo Step and the outer
// Bundle's echo Step must both still run, because both run_if decisions were made on entry while FLAG
// was still true. With the per-Step re-evaluation bug both echo Steps would be skipped.
func TestStepBundleRunIfEncapsulation(t *testing.T) {
	configPth := "step_bundle_run_if_test_bitrise.yml"

	cmd := command.New(testhelpers.BinPath(), "run", "--output-format", "json", "encapsulation", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	stepOutputs := testhelpers.CollectStepOutputs(out, t)
	require.Equal(t, stepOutputs, []string{
		"flip_flag_inner\n",
		"flip_flag_outer\n",
	})
}
