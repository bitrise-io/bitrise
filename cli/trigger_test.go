package cli

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func toBase64(t *testing.T, str string) string {
	bytes := base64.StdEncoding.EncodeToString([]byte(str))
	return string(bytes)
}

func TestParseRunOrTriggerJSONParams(t *testing.T) {
	t.Log("it parses json string-string map")
	{
		params, err := parseRunOrTriggerJSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`)
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("null is a valid json struct")
	{
		params, err := parseRunOrTriggerJSONParams(`null`)
		require.NoError(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, "", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigPath)
	}

	t.Log("it returns error for invalid json")
	{
		params, err := parseRunOrTriggerJSONParams("primary")
		require.Error(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, "", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigPath)
	}
}

func TestParseRunOrTriggerBase64JSONParams(t *testing.T) {
	t.Log("it parses base 64 encoded json string-string map")
	{
		base64JsonParams := toBase64(t, `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`)

		params, err := parseRunOrTriggerBase64JSONParams(base64JsonParams)
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("it returns error for not base 64 encoded json string-string map")
	{
		params, err := parseRunOrTriggerBase64JSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`)
		require.Error(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, "", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigPath)
	}
}

func TestParseRunOrTriggerParams(t *testing.T) {
	t.Log("it creates params from cli flags")
	{
		workflowToRunID := "primary"
		triggerPattern := ""

		bitriseConfigBase64Data := ""
		bitriseConfigPath := "bitrise.yml"

		inventoryBase64Data := ""
		inventoryPath := ".secrets.bitrise.yml"

		jsonParams := ""
		jsonParamsBase64 := ""

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("it creates params from cli flags used in ci mode")
	{
		workflowToRunID := ""
		triggerPattern := "master"

		bitriseConfigBase64Data := toBase64(t, "bitrise.yml")
		bitriseConfigPath := ""

		inventoryBase64Data := toBase64(t, ".secrets.bitrise.yml")
		inventoryPath := ""

		jsonParams := ""
		jsonParamsBase64 := ""

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "master", params.TriggerPattern)

		require.Equal(t, toBase64(t, ".secrets.bitrise.yml"), params.InventoryBase64Data)
		require.Equal(t, "", params.InventoryPath)

		require.Equal(t, toBase64(t, "bitrise.yml"), params.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigPath)
	}

	t.Log("it creates params from json params")
	{
		workflowToRunID := ""
		triggerPattern := ""

		bitriseConfigBase64Data := ""
		bitriseConfigPath := ""

		inventoryBase64Data := ""
		inventoryPath := ""

		jsonParams := `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`
		jsonParamsBase64 := ""

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)
		require.Equal(t, "", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("it creates params from base64 json params")
	{
		workflowToRunID := ""
		triggerPattern := ""

		bitriseConfigBase64Data := ""
		bitriseConfigPath := ""

		inventoryBase64Data := ""
		inventoryPath := ""

		jsonParams := ""
		jsonParamsBase64 := toBase64(t, `{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "pattern":"master"}`)

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "master", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("it creates params with preference jsonParams > jsonParamsBase64")
	{
		workflowToRunID := ""
		triggerPattern := ""

		bitriseConfigBase64Data := ""
		bitriseConfigPath := ""

		inventoryBase64Data := ""
		inventoryPath := ""

		jsonParams := `{"config":"test_bitrise.yml", "pattern":"develop"}`
		jsonParamsBase64 := toBase64(t, `{"config":"integration_bitrise.yml", "workflow":"fallback"}`)

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "", params.WorkflowToRunID)
		require.Equal(t, "develop", params.TriggerPattern)

		require.Equal(t, "", params.InventoryBase64Data)
		require.Equal(t, "", params.InventoryPath)

		require.Equal(t, "", params.BitriseConfigBase64Data)
		require.Equal(t, "test_bitrise.yml", params.BitriseConfigPath)
	}

	t.Log("cli params can owerride jsonParams and jsonParamsBase64")
	{
		workflowToRunID := "primary"
		triggerPattern := "master"

		inventoryBase64Data := "asd"
		inventoryPath := ".secrets.bitrise.yml"

		bitriseConfigBase64Data := "abc"
		bitriseConfigPath := "bitrise.yml"

		jsonParams := ""
		jsonParamsBase64 := toBase64(t, `{"workflow":"fallback","pattern":"develop","config-base64":"qwe","config":"integration_bitrise.yml","inventory-base64": "rtz","inventory":".secrets.yml"}`)

		params, err := parseRunOrTriggerParams(
			workflowToRunID, triggerPattern,
			inventoryBase64Data, inventoryPath,
			bitriseConfigBase64Data, bitriseConfigPath,
			jsonParams, jsonParamsBase64,
		)
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)
		require.Equal(t, "master", params.TriggerPattern)

		require.Equal(t, "asd", params.InventoryBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)

		require.Equal(t, "abc", params.BitriseConfigBase64Data)
		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
	}
}
