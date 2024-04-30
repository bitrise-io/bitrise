package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/stretchr/testify/require"
)

func TestStepmanJSONStepLibStepInfo(t *testing.T) {
	// setup
	require.NoError(t, configs.InitPaths())

	// Valid params -- Err should empty, output filled
	require.Equal(t, nil, StepmanSetup("https://github.com/bitrise-io/bitrise-steplib"))

	info, err := StepmanStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "0.9.0")
	require.NoError(t, err)
	require.NotEqual(t, "", info.JSON())

	// Invalid params -- Err returned, output is invalid
	info, err = StepmanStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "2.x")
	require.Error(t, err)
}
