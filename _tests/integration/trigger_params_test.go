package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_TriggerParams(t *testing.T) {
	t.Log("deprecated trigger with pattern - pr allowed")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "deprecated_pr_allowed", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pattern - pr not allowed")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "--pattern", "deprecated_only_code_push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with push-branch")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "--push-branch", "code-push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pr-source-branch")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "--pr-source-branch", "pull_request_source_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pr-target-branch")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "--pr-target-branch", "pull_request_taget_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with tag")
	{
		cmd := cmdex.NewCommand(binPath(), "trigger", "--tag", "tag", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
