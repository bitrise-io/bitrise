package cli

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func toBase64(t *testing.T, str string) string {
	bytes := base64.StdEncoding.EncodeToString([]byte(str))
	return string(bytes)
}

func toJSON(t *testing.T, stringStringMap map[string]string) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func TestParseTriggerCheckParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64(t, "bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(t, ".secrets.bitrise.yml")

		jsonParams := ""
		base64JSONParams := ""

		params, err := parseTriggerCheckParams(
			pattern,
			pushBranch, prSourceBranch, prTargetBranch,
			format,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams,
		)
		require.NoError(t, err)

		require.Equal(t, "", params.WorkflowToRunID)

		require.Equal(t, pattern, params.TriggerPattern)
		require.Equal(t, pushBranch, params.PushBranch)
		require.Equal(t, prSourceBranch, params.PRSourceBranch)
		require.Equal(t, prTargetBranch, params.PRTargetBranch)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}
}

func TestGetWorkflowIDByPattern(t *testing.T) {
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
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	t.Log("Default pattern defined &  Pull request mode")
	{
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", true)
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
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("No default pattern defined & Pull request mode")
	{
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}
}
