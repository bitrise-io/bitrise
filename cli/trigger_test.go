package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTriggerJSONParams(t *testing.T) {
	t.Log("it parses cli json params")
	{
		params, err := parseTriggerJSONParams(`{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master"}`)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}
}

func TestParseTriggerParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		triggerPattern := "*"
		gitTriggerEvent := "code-push"
		sourceBranch := "dev"
		targetBranch := "master"
		format := "json"
		jsonParams := ""
		base64JSONParams := ""

		params, err := parseTriggerParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "json", params.Format)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}

	t.Log("it parses json params")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		format := ""
		jsonParams := `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","format":"json"}`
		base64JSONParams := ""

		params, err := parseTriggerParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "json", params.Format)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}

	t.Log("it parses json params decoded in base64")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		format := ""
		jsonParams := ""
		base64JSONParams := toBase64(t, `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","format":"json"}`)

		params, err := parseTriggerParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "json", params.Format)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}

	t.Log("json params has priority over json params encoded in base 64")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		format := ""
		jsonParams := `{"pattern":"feature/","event":"pull-request","source-branch":"feature","target-branch":"dev","format":"raw"}`
		base64JSONParams := toBase64(t, `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","format":"json"}`)

		params, err := parseTriggerParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "feature/", params.TriggerPattern)
		require.Equal(t, "pull-request", params.GitTriggerEvent)
		require.Equal(t, "feature", params.SourceBranch)
		require.Equal(t, "dev", params.TargetBranch)
		require.Equal(t, "raw", params.Format)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}

	t.Log("cli params can override json params")
	{
		triggerPattern := "feature/"
		gitTriggerEvent := "pull-request"
		sourceBranch := "feature"
		targetBranch := "dev"
		format := "raw"
		jsonParams := `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","format":"json"}`
		base64JSONParams := ""

		params, err := parseTriggerParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "feature/", params.TriggerPattern)
		require.Equal(t, "pull-request", params.GitTriggerEvent)
		require.Equal(t, "feature", params.SourceBranch)
		require.Equal(t, "dev", params.TargetBranch)
		require.Equal(t, "raw", params.Format)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, "", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
	}
}

func TestParseTriggerCommandParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		triggerPattern := "*"
		gitTriggerEvent := "code-push"
		sourceBranch := "dev"
		targetBranch := "master"
		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := ""
		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := ""
		format := "json"
		jsonParams := ""
		base64JSONParams := ""

		params, err := parseTriggerCommandParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "bitrise.yml", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
		require.Equal(t, "json", params.Format)
	}

	t.Log("it parses json params")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		format := ""
		jsonParams := `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","config":"bitrise.yml","inventory":".secrets.bitrise.yml","format":"json"}`
		base64JSONParams := ""

		params, err := parseTriggerCommandParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "bitrise.yml", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
		require.Equal(t, "json", params.Format)
	}

	t.Log("it parses json params decoded in base64")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		format := ""
		jsonParams := ""
		base64JSONParams := toBase64(t, `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","config":"bitrise.yml","inventory":".secrets.bitrise.yml","format":"json"}`)

		params, err := parseTriggerCommandParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "bitrise.yml", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
		require.Equal(t, "json", params.Format)
	}

	t.Log("json params has priority over json params encoded in base 64")
	{
		triggerPattern := ""
		gitTriggerEvent := ""
		sourceBranch := ""
		targetBranch := ""
		bitriseConfigPath := ""
		bitriseConfigBase64Data := ""
		inventoryPath := ""
		inventoryBase64Data := ""
		format := ""
		jsonParams := `{"pattern":"feature/","event":"pull-request","source-branch":"feature","target-branch":"dev","config":"test-bitrise.yml","inventory":".test-secrets.bitrise.yml","format":"json"}`
		base64JSONParams := toBase64(t, `{"pattern":"*","event":"code-push","source-branch":"dev","target-branch":"master","config":"bitrise.yml","inventory":".secrets.bitrise.yml","format":"raw"}`)

		params, err := parseTriggerCommandParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "feature/", params.TriggerPattern)
		require.Equal(t, "pull-request", params.GitTriggerEvent)
		require.Equal(t, "feature", params.SourceBranch)
		require.Equal(t, "dev", params.TargetBranch)
		require.Equal(t, "test-bitrise.yml", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, ".test-secrets.bitrise.yml", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
		require.Equal(t, "json", params.Format)
	}

	t.Log("cli params can override json params")
	{
		triggerPattern := "*"
		gitTriggerEvent := "code-push"
		sourceBranch := "dev"
		targetBranch := "master"
		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := ""
		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := ""
		format := "json"
		jsonParams := `{"pattern":"feature/","event":"pull-request","source-branch":"feature","target-branch":"dev","config":"test-bitrise.yml","inventory":".test-secrets.bitrise.yml","format":"raw"}`
		base64JSONParams := ""

		params, err := parseTriggerCommandParams(
			triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			format,
			jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "*", params.TriggerPattern)
		require.Equal(t, "code-push", params.GitTriggerEvent)
		require.Equal(t, "dev", params.SourceBranch)
		require.Equal(t, "master", params.TargetBranch)
		require.Equal(t, "bitrise.yml", params.BitriseConfigParams.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigParams.BitriseConfigBase64Data)
		require.Equal(t, ".secrets.bitrise.yml", params.BitriseConfigParams.InventoryPath)
		require.Equal(t, "", params.BitriseConfigParams.InventoryBase64Data)
		require.Equal(t, "json", params.Format)
	}
}
