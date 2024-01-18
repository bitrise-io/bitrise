//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_GlobalFlagPRRun(t *testing.T) {
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
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	t.Run("global flag sets pr mode", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		cmd := command.New(binPath(), "--pr", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	})

	t.Run("global flag sets pr mode", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		cmd := command.New(binPath(), "--pr=true", "trigger-check", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	})

	t.Run("global flag sets NOT pr mode", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "true")
		t.Setenv(configs.PullRequestIDEnvKey, "ID")

		cmd := command.New(binPath(), "--pr=true", "trigger-check", "master", "--config", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"master","workflow":"deprecated_pr"}`, out)
	})

	t.Run("global flag sets NOT pr mode", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "true")
		t.Setenv(configs.PullRequestIDEnvKey, "ID")

		cmd := command.New(binPath(), "--pr=false", "trigger-check", "master", "--config", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"master","workflow":"deprecated_code_push"}`, out)
	})
}

func Test_GlobalFlagPRTrigger(t *testing.T) {
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	t.Setenv(configs.PRModeEnvKey, "false")
	t.Setenv(configs.PullRequestIDEnvKey, "")

	t.Run("global flag sets pr mode", func(t *testing.T) {
		cmd := command.New(binPath(), "--pr", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	})

	t.Run("global flag sets pr mode", func(t *testing.T) {
		cmd := command.New(binPath(), "--pr=true", "trigger", "deprecated_pr", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	})
}

func Test_GlobalFlagCI(t *testing.T) {
	configPth := "global_flag_test_bitrise.yml"
	secretsPth := "global_flag_test_secrets.yml"

	t.Run("Should run in ci mode", func(t *testing.T) {
		t.Setenv(configs.CIModeEnvKey, "false")

		cmd := command.New(binPath(), "--ci", "run", "fail_in_not_ci_mode", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	})

	t.Run("Should run in ci mode", func(t *testing.T) {
		t.Setenv(configs.CIModeEnvKey, "false")

		cmd := command.New(binPath(), "--ci=true", "run", "fail_in_not_ci_mode", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	})

	t.Run("Should run in ci mode", func(t *testing.T) {
		t.Setenv(configs.CIModeEnvKey, "true")

		cmd := command.New(binPath(), "--ci=false", "run", "fail_in_ci_mode", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	})
}
