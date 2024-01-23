//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_RunExitCode(t *testing.T) {
	configPth := "exit_code_test_bitrise.yml"

	t.Log("exit_code_test_fail")
	{
		cmd := command.New(binPath(), "run", "exit_code_test_fail", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("exit_code_test_ok")
	{
		cmd := command.New(binPath(), "run", "exit_code_test_ok", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("exit_code_test_sippable_fail")
	{
		cmd := command.New(binPath(), "run", "exit_code_test_sippable_fail", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("exit_code_test_sippable_ok")
	{
		cmd := command.New(binPath(), "run", "exit_code_test_sippable_ok", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}

func Test_TriggerExitCode(t *testing.T) {
	configPth := "exit_code_test_bitrise.yml"

	t.Log("exit_code_test_fail")
	{
		cmd := command.New(binPath(), "trigger", "exit_code_test_fail", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("exit_code_test_ok")
	{
		cmd := command.New(binPath(), "trigger", "exit_code_test_ok", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("exit_code_test_sippable_fail")
	{
		cmd := command.New(binPath(), "trigger", "exit_code_test_sippable_fail", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("exit_code_test_sippable_ok")
	{
		cmd := command.New(binPath(), "trigger", "exit_code_test_sippable_ok", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
