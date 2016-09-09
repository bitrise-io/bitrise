package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func TestMigratePatternToParams(t *testing.T) {
	t.Log("converts pattern in NON PR MODE to push-branch param")
	{
		isPullRequestMode := false
		params := RunAndTriggerParamsModel{
			TriggerPattern: "master",
		}

		convertedParams := migratePatternToParams(params, isPullRequestMode)

		require.Equal(t, "master", convertedParams.PushBranch)
		require.Equal(t, "", convertedParams.PRSourceBranch)
		require.Equal(t, "", convertedParams.PRTargetBranch)

		require.Equal(t, "", convertedParams.WorkflowToRunID)
		require.Equal(t, "", convertedParams.TriggerPattern)
		require.Equal(t, "", convertedParams.Format)
		require.Equal(t, "", convertedParams.BitriseConfigPath)
		require.Equal(t, "", convertedParams.BitriseConfigBase64Data)
		require.Equal(t, "", convertedParams.InventoryPath)
		require.Equal(t, "", convertedParams.InventoryBase64Data)
	}

	t.Log("converts pattern in PR MODE to pr-source-branch param")
	{
		isPullRequestMode := true
		params := RunAndTriggerParamsModel{
			TriggerPattern: "master",
		}

		convertedParams := migratePatternToParams(params, isPullRequestMode)

		require.Equal(t, "", convertedParams.PushBranch)
		require.Equal(t, "master", convertedParams.PRSourceBranch)
		require.Equal(t, "", convertedParams.PRTargetBranch)

		require.Equal(t, "", convertedParams.WorkflowToRunID)
		require.Equal(t, "", convertedParams.TriggerPattern)
		require.Equal(t, "", convertedParams.Format)
		require.Equal(t, "", convertedParams.BitriseConfigPath)
		require.Equal(t, "", convertedParams.BitriseConfigBase64Data)
		require.Equal(t, "", convertedParams.InventoryPath)
		require.Equal(t, "", convertedParams.InventoryBase64Data)
	}
}

/*
func getWorkflowIDByTriggerParams(triggerMap models.TriggerMapModel, pattern string, isPullRequestMode bool, params RunAndTriggerParamsModel) (string, error) {
	if pattern != "" {
		// Deprecated trigger item
		return getWorkflowIDByParamsInCompatibleMode(triggerMap, pattern, isPullRequestMode)
	}

	for _, item := range triggerMap {
		match, err := item.MatchWithParams(params.PushBranch, params.PRSourceBranch, params.PRTargetBranch)
		if err != nil {
			return "", err
		}
		if match {
			return item.WorkflowID, nil
		}
	}

	return "", fmt.Errorf("Run triggered with params: push-branch: %s, pr-source-branch: %s, pr-target-branch: %s, but no matching workflow found", params.PushBranch, params.PRSourceBranch, params.PRTargetBranch)
}
*/

func TestGetWorkflowIDByParamsInCompatibleMode(t *testing.T) {
	configStr := `
trigger_map:
- pattern: master
  is_pull_request_allowed: false
  workflow: master
- pattern: feature/*
  is_pull_request_allowed: true
  workflow: feature
- pattern: "*"
  is_pull_request_allowed: true
  workflow: primary

workflows:
  test:
  master:
  feature:
  primary:
`
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	t.Log("Default pattern defined & Non pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	t.Log("Default pattern defined &  Pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	configStr = `
  trigger_map:
  - pattern: master
    is_pull_request_allowed: false
    workflow: master
  - pattern: feature/*
    is_pull_request_allowed: true
    workflow: feature

  workflows:
    test:
    master:
    feature:
    primary:
  `
	config, warnings, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	t.Log("No default pattern defined & Non pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("No default pattern defined & Pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}
}
