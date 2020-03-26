package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_TriggerParams(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("deprecated trigger with pattern - pr allowed")
	{
		cmd := command.New(binPath(), "trigger", "deprecated_pr_allowed", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pattern - pr not allowed")
	{
		cmd := command.New(binPath(), "trigger", "--pattern", "deprecated_only_code_push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with push-branch")
	{
		cmd := command.New(binPath(), "trigger", "--push-branch", "code-push", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pr-source-branch")
	{
		cmd := command.New(binPath(), "trigger", "--pr-source-branch", "pull_request_source_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with pr-target-branch")
	{
		cmd := command.New(binPath(), "trigger", "--pr-target-branch", "pull_request_taget_branch", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("deprecated trigger with tag")
	{
		cmd := command.New(binPath(), "trigger", "--tag", "tag", "--config", "trigger_params_test_bitrise.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
