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
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "deprecated_code_push",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"deprecated_code_push","workflow":"deprecated_code_push"}`, out)
	}

	t.Log("deprecated trigger test - PR mode")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "deprecated_pr",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"deprecated_pr","workflow":"deprecated_pr"}`, out)
	}

	t.Log("deprecated trigger test - pipeline")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "deprecated_pipeline",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"deprecated_pipeline","pipeline":"deprecated_pipeline"}`, out)
	}

	t.Log("new trigger test - code push")
	{
		config := map[string]interface{}{
			"config":      configPth,
			"push-branch": "code_push",
			"format":      "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"push-branch":"code_push","workflow":"code_push"}`, out)
	}

	t.Log("new trigger test - code push - no match")
	{
		config := map[string]interface{}{
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
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target","workflow":"pr_source_and_target"}`, out)
	}

	t.Log("new trigger test - pull request - defined source and target pattern  - no match")
	{
		config := map[string]interface{}{
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
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params-base64", toBase64(t, toJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target","workflow":"pr_source_and_target"}`, out)
	}

	t.Log("new trigger test - pull request - defined target pattern")
	{
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "pr_source",
			"pr-target-branch": "pr_target_only",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source","pr-target-branch":"pr_target_only","workflow":"pr_target"}`, out)
	}

	t.Log("new trigger test - pull request - defined source pattern")
	{
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "pr_source_only",
			"pr-target-branch": "pr_target",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"pr_source_only","pr-target-branch":"pr_target","workflow":"pr_source"}`, out)
	}

	t.Log("new trigger test - pipeline")
	{
		config := map[string]interface{}{
			"config":      configPth,
			"push-branch": "pipeline_code_push",
			"format":      "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pipeline":"pipeline_code_push","push-branch":"pipeline_code_push"}`, out)
	}

	t.Log("draft pr control test - draft pr disabled - ready to review pr trigger")
	{
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "no_draft_pr",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pr-source-branch":"no_draft_pr","workflow":"no_draft_pr"}`, out)
	}

	t.Log("draft pr control test - draft pr disabled - draft pr trigger")
	{
		config := map[string]interface{}{
			"config":           configPth,
			"pr-source-branch": "no_draft_pr",
			"pr-ready-state":   "draft",
			"format":           "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		require.Equal(t, `{"is_valid":true,"error":"no matching pipeline \u0026 workflow found with trigger params: push-branch: , pr-source-branch: no_draft_pr, pr-target-branch: , tag: "}`, out)
	}
}
