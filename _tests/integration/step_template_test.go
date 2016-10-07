package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_StepTemplate(t *testing.T) {
	configPth := "bitrise.yml"

	t.Log("step template test")
	{
		cmd := cmdex.NewCommand(binPath(), "run", "test", "--config", configPth)
		cmd.SetDir("step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("bash toolkit step template test")
	{
		cmd := cmdex.NewCommand(binPath(), "run", "test", "--config", configPth)
		cmd.SetDir("bash_toolkit_step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
