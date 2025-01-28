package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestStepBundles(t *testing.T) {
	configPth := "step_bundles_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "test_step_bundle_inputs", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello Bitrise!")
}
