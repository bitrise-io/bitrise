package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_RunExitCode(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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
