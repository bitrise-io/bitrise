//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_StepTemplate(t *testing.T) {
	configPth := "bitrise.yml"

	t.Log("step template test")
	{
		cmd := command.New(testhelpers.BinPath(), "run", "test", "--config", configPth)
		cmd.SetDir("step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("bash toolkit step template test")
	{
		cmd := command.New(testhelpers.BinPath(), "run", "test", "--config", configPth)
		cmd.SetDir("bash_toolkit_step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("go toolkit step template test")
	{
		cmd := command.New(testhelpers.BinPath(), "run", "test", "--config", configPth)
		cmd.SetDir("go_toolkit_step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}