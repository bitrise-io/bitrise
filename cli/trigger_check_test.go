package cli

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func toBase64(t *testing.T, str string) string {
	bytes := base64.StdEncoding.EncodeToString([]byte(str))
	return string(bytes)
}

func TestParseBitriseConfigJSONParams(t *testing.T) {
	t.Log("it parses cli json params")
	{
		params, err := parseBitriseConfigJSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml"}`)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("it parses cli json params decoded in base64")
	{
		configBase64 := toBase64(t, "my config content")
		inventoryBase64 := toBase64(t, "my secrets content")
		jsonParams := fmt.Sprintf(`{"config-base64":"%s","inventory-base64":"%s"}`, configBase64, inventoryBase64)

		params, err := parseBitriseConfigJSONParams(jsonParams)
		require.NoError(t, err)

		require.Equal(t, "", params.BitriseConfigPath)
		require.Equal(t, configBase64, params.BitriseConfigBase64Data)
		require.Equal(t, "", params.InventoryPath)
		require.Equal(t, inventoryBase64, params.InventoryBase64Data)
	}

	t.Log("it parses bitrise config related data from run command json params")
	{
		params, err := parseBitriseConfigJSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml","workflow":"primary"}`)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("it parses bitrise config related data from trigger command json params")
	{
		params, err := parseBitriseConfigJSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml","pattern":"master","source-branch":"dev","target-barnch":"master","event":"code-push"}`)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}
}

func TestParseBitriseConfigParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := ""
		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := ""
		jsonParams := ""
		base64JSONParams := ""

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("it parses cli params decoded in base64")
	{
		bitriseConfigPath := ""
		bitriseConfigBase64Data := toBase64(t, "my config content")
		inventoryPath := ""
		inventoryBase64Data := toBase64(t, "my secrets content")
		jsonParams := ""
		base64JSONParams := ""

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "", params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)
		require.Equal(t, "", params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}

	t.Log("it parses json params")
	{
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		jsonParams := `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml","pattern":"master","source-branch":"dev","target-barnch":"master","event":"code-push"}`
		base64JSONParams := ""

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("it parses json params decoded in base64")
	{
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		jsonParams := ""
		base64JSONParams := toBase64(t, `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml","pattern":"master","source-branch":"dev","target-barnch":"master","event":"code-push"}`)

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("json params has priority over json params encoded in base 64")
	{
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		jsonParams := `{"config":"test-bitrise.yml","inventory":".test-secrets.bitrise.yml","pattern":"dev","source-branch":"feature","target-barnch":"dev","event":"pull-requiest"}`
		base64JSONParams := toBase64(t, `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml","pattern":"master","source-branch":"dev","target-barnch":"master","event":"code-push"}`)

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "test-bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".test-secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("cli params can override json params")
	{
		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := ""
		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := ""
		jsonParams := `{"config":"test-bitrise.yml","inventory":".test-secrets.bitrise.yml","pattern":"dev","source-branch":"feature","target-barnch":"dev","event":"pull-requiest"}`
		base64JSONParams := ""

		params, err := parseBitriseConfigParams(
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
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
