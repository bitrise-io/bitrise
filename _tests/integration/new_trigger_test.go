package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_NewTrigger(t *testing.T) {
	configPth := "new_trigger_test_bitrise.yml"

	t.Log("deprecated trigger test")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "deprecated_code_push",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"deprecated_code_push","workflows":["deprecated_code_push"]}`, out)
	}

	t.Log("deprecated trigger test - PR mode")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "deprecated_pr",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"deprecated_pr","workflows":["deprecated_pr"]}`, out)
	}

	t.Log("new trigger test - code push")
	{
		config := map[string]string{
			"config":      configPth,
			"push-branch": "code_push",
			"format":      "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"push-branch":"code_push","workflows":["code_push"]}`, out)
	}

	t.Log("new trigger test - code push - no match")
	{
		config := map[string]string{
			"config":      configPth,
			"push-branch": "no_match",
			"format":      "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("new trigger test - pull request - defined source and target pattern")
	{
		config := map[string]string{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target","workflows":["pr_source_and_target"]}`, out)
	}

	t.Log("new trigger test - pull request - defined source and target pattern  - no match")
	{
		config := map[string]string{
			"config":           configPth,
			"pr-source-branch": "no_match",
			"pr-target-branch": "no_match",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("new trigger test base64 - pull request - defined source and target pattern")
	{
		config := map[string]string{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params-base64", toBase64(t, toJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target","workflows":["pr_source_and_target"]}`, out)
	}

	t.Log("new trigger test - pull request - defined target pattern")
	{
		config := map[string]string{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target_only",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target_only","workflows":["pr_target"]}`, out)
	}

	t.Log("new trigger test - pull request - defined source pattern")
	{
		config := map[string]string{
			"config":           configPth,
			"pr-source-branch": "pr_source_only",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source_only","pr-target-branch":"pr_target","workflows":["pr_source"]}`, out)
	}
}
