package integration

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_GlobalFlagRun(t *testing.T) {
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_bitrise.yml"

	t.Log("Should run in pr mode")
	{
		cmd := cmdex.NewCommand(binPath(), "--pr", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in pr mode")
	{
		cmd := cmdex.NewCommand(binPath(), "--pr=true", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("Should run in non pr mode")
	{
		cmd := cmdex.NewCommand(binPath(), "--pr=false", "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}

func Test_GlobalFlagTriggerCheck(t *testing.T) {
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_bitrise.yml"

	prModeEnv := os.Getenv(configs.PRModeEnvKey)
	prIDEnv := os.Getenv(configs.PullRequestIDEnvKey)

	// cleanup Envs after these tests
	defer func() {
		require.NoError(t, os.Setenv(configs.PRModeEnvKey, prModeEnv))
		require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, prIDEnv))
	}()

	t.Log("global flag sets pr mode")
	{
		require.NoError(t, os.Setenv("PR", "false"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", ""))

		cmd := cmdex.NewCommand(binPath(), "--pr", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets pr mode")
	{

		require.NoError(t, os.Setenv("PR", "false"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", ""))

		cmd := cmdex.NewCommand(binPath(), "--pr=true", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets NOT pr mode")
	{
		require.NoError(t, os.Setenv("PR", "true"))
		require.NoError(t, os.Setenv("PULL_REQUEST_ID", "ID"))

		cmd := cmdex.NewCommand(binPath(), "--pr=false", "trigger-check", "master", "--config", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"master","workflow":"deprecated_code_push"}`, out)
	}
}

func Test_GlobalFlagTrigger(t *testing.T) {
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_bitrise.yml"

	prModeEnv := os.Getenv(configs.PRModeEnvKey)
	prIDEnv := os.Getenv(configs.PullRequestIDEnvKey)

	// cleanup Envs after these tests
	defer func() {
		require.NoError(t, os.Setenv(configs.PRModeEnvKey, prModeEnv))
		require.NoError(t, os.Setenv(configs.PullRequestIDEnvKey, prIDEnv))
	}()

	require.NoError(t, os.Setenv("PR", "false"))
	require.NoError(t, os.Setenv("PULL_REQUEST_ID", ""))

	t.Log("global flag sets pr mode")
	{
		cmd := cmdex.NewCommand(binPath(), "--pr", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("global flag sets pr mode")
	{
		cmd := cmdex.NewCommand(binPath(), "--pr=true", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}
}
