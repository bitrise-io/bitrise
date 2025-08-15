//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_TriggerParams(t *testing.T) {
	t.Log("deprecated trigger with pattern - pr allowed")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "deprecated_pr_allowed", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pattern - pr not allowed")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--pattern", "deprecated_only_code_push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger with push-branch")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--push-branch", "code-push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger with pr-source-branch")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--pr-source-branch", "pull_request_source_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger with pr-target-branch")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--pr-target-branch", "pull_request_taget_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger with tag")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--tag", "tag", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated pipeline trigger with pattern")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--pattern", "deprecated_pipeline_trigger", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("pipeline trigger with push-branch")
	{
		cmd := command.New(testhelpers.BinPath(), "trigger", "--push-branch", "pipeline_trigger", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}
}
