package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_GlobalFlagPRRun(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	t.Log("Should run in pr mode")
	{
		cmd := command.New(binPath(), "--pr", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in pr mode")
	{
		cmd := command.New(binPath(), "--pr=true", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in non pr mode")
	{
		cmd := command.New(binPath(), "--pr=false", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}

func Test_GlobalFlagPRTriggerCheck(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	prModeEnv := os.Getenv(configs.PRModeEnvKey)
	prIDEnv := os.Getenv(configs.PullRequestIDEnvKey)

	// cleanup Envs after these tests
	defer func() {
		require.NoError(t, os.Setenv(configs.PRModeEnvKey, prModeEnv))
		require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, prIDEnv))
	}()

	t.Log("global flag sets pr mode")
	{
		require.NoError(t, os.Setenv(configs.PRModeEnvKey, "false"))
		require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, ""))

		cmd := command.New(binPath(), "--pr", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets pr mode")
	{

		require.NoError(t, os.Setenv("PR", "false"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", ""))

		cmd := command.New(binPath(), "--pr=true", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets NOT pr mode")
	{
		require.NoError(t, os.Setenv("PR", "true"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", "ID"))

		cmd := command.New(binPath(), "--pr=true", "trigger-check", "master", "--config", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"master","workflow":"deprecated_pr"}`, out)
	}

	t.Log("global flag sets NOT pr mode")
	{
		require.NoError(t, os.Setenv("PR", "true"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", "ID"))

		cmd := command.New(binPath(), "--pr=false", "trigger-check", "master", "--config", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"master","workflow":"deprecated_code_push"}`, out)
	}
}

func Test_GlobalFlagPRTrigger(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	prModeEnv := os.Getenv(configs.PRModeEnvKey)
	prIDEnv := os.Getenv(configs.PullRequestIDEnvKey)

	// cleanup Envs after these tests
	defer func() {
		require.NoError(t, os.Setenv(configs.PRModeEnvKey, prModeEnv))
		require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, prIDEnv))
	}()

	require.NoError(t, os.Setenv(configs.PRModeEnvKey, "false"))
	require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, ""))

	t.Log("global flag sets pr mode")
	{
		cmd := command.New(binPath(), "--pr", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets pr mode")
	{
		cmd := command.New(binPath(), "--pr=true", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}
}

func Test_GlobalFlagCI(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	ciModeEnv := os.Getenv(configs.CIModeEnvKey)

	// cleanup Envs after these tests
	defer func() {
		require.NoError(t, os.Setenv(configs.CIModeEnvKey, ciModeEnv))
	}()

	t.Log("Should run in ci mode")
	{
		require.NoError(t, os.Setenv(configs.CIModeEnvKey, "false"))

		cmd := command.New(binPath(), "--ci", "run", "fail_in_not_ci_mode", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in ci mode")
	{
		require.NoError(t, os.Setenv(configs.CIModeEnvKey, "false"))

		cmd := command.New(binPath(), "--ci=true", "run", "fail_in_not_ci_mode", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in ci mode")
	{
		require.NoError(t, os.Setenv(configs.CIModeEnvKey, "true"))

		cmd := command.New(binPath(), "--ci=false", "run", "fail_in_ci_mode", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
