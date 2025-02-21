package cli

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/require"
)

func toBase64(str string) string {
	bytes := base64.StdEncoding.EncodeToString([]byte(str))
	return string(bytes)
}

func toJSON(t *testing.T, stringStringMap map[string]interface{}) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func TestParseRunAndTriggerJSONParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		paramsMap := map[string]interface{}{
			WorkflowKey: "primary",

			PatternKey:        "master",
			PushBranchKey:     "deploy",
			PRSourceBranchKey: "development",
			PRTargetBranchKey: "release",
			PRReadyStateKey:   models.PullRequestReadyStateReadyForReview,
			TagKey:            "0.9.0",

			OuputFormatKey: "json",

			ConfigKey:       "bitrise.yml",
			ConfigBase64Key: toBase64("bitrise.yml"),

			InventoryKey:       ".secrets.bitrise.yml",
			InventoryBase64Key: toBase64(".secrets.bitrise.yml"),
		}
		params, err := parseRunAndTriggerJSONParams(toJSON(t, paramsMap))
		require.NoError(t, err)

		require.Equal(t, "primary", params.WorkflowToRunID)

		require.Equal(t, "master", params.TriggerPattern)
		require.Equal(t, "deploy", params.PushBranch)
		require.Equal(t, "development", params.PRSourceBranch)
		require.Equal(t, "release", params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyStateReadyForReview, params.PRReadyState)
		require.Equal(t, "0.9.0", params.Tag)

		require.Equal(t, "json", params.Format)

		require.Equal(t, "bitrise.yml", params.BitriseConfigPath)
		require.Equal(t, toBase64("bitrise.yml"), params.BitriseConfigBase64Data)

		require.Equal(t, ".secrets.bitrise.yml", params.InventoryPath)
		require.Equal(t, toBase64(".secrets.bitrise.yml"), params.InventoryBase64Data)
	}

	t.Log("it fails for invalid json")
	{
		params, err := parseRunAndTriggerJSONParams("master")
		require.Error(t, err)

		require.Equal(t, "", params.WorkflowToRunID)

		require.Equal(t, "", params.TriggerPattern)
		require.Equal(t, "", params.PushBranch)
		require.Equal(t, "", params.PRSourceBranch)
		require.Equal(t, "", params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyState(""), params.PRReadyState)

		require.Equal(t, "", params.Format)

		require.Equal(t, "", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)

		require.Equal(t, "", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}
}

func TestParseRunAndTriggerParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		workflow := "primary"

		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateReadyForReview
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		jsonParams := ""
		base64JSONParams := ""

		params, err := parseRunAndTriggerParams(
			workflow,
			pattern,
			pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
			format,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams,
		)
		require.NoError(t, err)

		require.Equal(t, workflow, params.WorkflowToRunID)

		require.Equal(t, pattern, params.TriggerPattern)
		require.Equal(t, pushBranch, params.PushBranch)
		require.Equal(t, prSourceBranch, params.PRSourceBranch)
		require.Equal(t, prTargetBranch, params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyStateReadyForReview, params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}

	t.Log("it parses json params")
	{
		workflow := "primary"

		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateDraft
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		paramsMap := map[string]interface{}{
			WorkflowKey: workflow,

			PatternKey:        pattern,
			PushBranchKey:     pushBranch,
			PRSourceBranchKey: prSourceBranch,
			PRTargetBranchKey: prTargetBranch,
			PRReadyStateKey:   prReadyState,
			TagKey:            tag,
			OuputFormatKey:    format,

			ConfigKey:       bitriseConfigPath,
			ConfigBase64Key: bitriseConfigBase64Data,

			InventoryKey:       inventoryPath,
			InventoryBase64Key: inventoryBase64Data,
		}

		jsonParams := toJSON(t, paramsMap)
		base64JSONParams := ""

		params, err := parseRunAndTriggerParams("", "", "", "", "", "", "", "", "", "", "", "", jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, workflow, params.WorkflowToRunID)

		require.Equal(t, pattern, params.TriggerPattern)
		require.Equal(t, pushBranch, params.PushBranch)
		require.Equal(t, prSourceBranch, params.PRSourceBranch)
		require.Equal(t, prTargetBranch, params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyState("draft"), params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}

	t.Log("it parses json params decoded in base64")
	{
		workflow := "primary"

		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateDraft
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		paramsMap := map[string]interface{}{
			WorkflowKey: workflow,

			PatternKey:        pattern,
			PushBranchKey:     pushBranch,
			PRSourceBranchKey: prSourceBranch,
			PRTargetBranchKey: prTargetBranch,
			PRReadyStateKey:   prReadyState,
			TagKey:            tag,
			OuputFormatKey:    format,

			ConfigKey:       bitriseConfigPath,
			ConfigBase64Key: bitriseConfigBase64Data,

			InventoryKey:       inventoryPath,
			InventoryBase64Key: inventoryBase64Data,
		}

		jsonParams := ""
		base64JSONParams := toBase64(toJSON(t, paramsMap))

		params, err := parseRunAndTriggerParams("", "", "", "", "", "", "", "", "", "", "", "", jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, workflow, params.WorkflowToRunID)

		require.Equal(t, pattern, params.TriggerPattern)
		require.Equal(t, pushBranch, params.PushBranch)
		require.Equal(t, prSourceBranch, params.PRSourceBranch)
		require.Equal(t, prTargetBranch, params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyState("draft"), params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}

	t.Log("json params has priority over json params encoded in base 64")
	{
		workflow := "primary"

		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateReadyForReview
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		paramsMap := map[string]interface{}{
			WorkflowKey: workflow,

			PatternKey:        pattern,
			PushBranchKey:     pushBranch,
			PRSourceBranchKey: prSourceBranch,
			PRTargetBranchKey: prTargetBranch,
			PRReadyStateKey:   prReadyState,
			TagKey:            tag,
			OuputFormatKey:    format,

			ConfigKey:       bitriseConfigPath,
			ConfigBase64Key: bitriseConfigBase64Data,

			InventoryKey:       inventoryPath,
			InventoryBase64Key: inventoryBase64Data,
		}

		jsonParams := `{"workflow":"test","pr-ready-state":"draft"}`
		base64JSONParams := toBase64(toJSON(t, paramsMap))

		params, err := parseRunAndTriggerParams("", "", "", "", "", "", "", "", "", "", "", "", jsonParams, base64JSONParams)
		require.NoError(t, err)

		require.Equal(t, "test", params.WorkflowToRunID)

		require.Equal(t, "", params.TriggerPattern)
		require.Equal(t, "", params.PushBranch)
		require.Equal(t, "", params.PRSourceBranch)
		require.Equal(t, "", params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyStateDraft, params.PRReadyState)
		require.Equal(t, "", params.Tag)

		require.Equal(t, "", params.Format)

		require.Equal(t, "", params.BitriseConfigPath)
		require.Equal(t, "", params.BitriseConfigBase64Data)

		require.Equal(t, "", params.InventoryPath)
		require.Equal(t, "", params.InventoryBase64Data)
	}

	t.Log("cli params can override json params")
	{
		workflow := "primary"

		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateDraft
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		jsonParams := `{"workflow":"test","pattern":"feature/","config":"test.bitrise.yml","inventory":".test.secrets.bitrise.yml"}`
		base64JSONParams := ""

		params, err := parseRunAndTriggerParams(
			workflow,
			pattern,
			pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
			format,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams,
		)
		require.NoError(t, err)

		require.Equal(t, workflow, params.WorkflowToRunID)

		require.Equal(t, pattern, params.TriggerPattern)
		require.Equal(t, pushBranch, params.PushBranch)
		require.Equal(t, prSourceBranch, params.PRSourceBranch)
		require.Equal(t, prTargetBranch, params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyStateDraft, params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}
}

func TestParseRunParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		workflow := "primary"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		jsonParams := ""
		base64JSONParams := ""

		params, err := parseRunParams(
			workflow,
			bitriseConfigPath, bitriseConfigBase64Data,
			inventoryPath, inventoryBase64Data,
			jsonParams, base64JSONParams,
		)
		require.NoError(t, err)

		require.Equal(t, workflow, params.WorkflowToRunID)

		require.Equal(t, "", params.TriggerPattern)
		require.Equal(t, "", params.PushBranch)
		require.Equal(t, "", params.PRSourceBranch)
		require.Equal(t, "", params.PRTargetBranch)
		require.Equal(t, models.PullRequestReadyState(""), params.PRReadyState)
		require.Equal(t, "", params.Tag)

		require.Equal(t, "", params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}
}

func TestParseTriggerParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateDraft
		tag := "0.9.0"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		jsonParams := ""
		base64JSONParams := ""

		params, err := parseTriggerParams(
			pattern,
			pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
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
		require.Equal(t, models.PullRequestReadyStateDraft, params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, "", params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}
}

func TestParseTriggerCheckParams(t *testing.T) {
	t.Log("it parses cli params")
	{
		pattern := "*"
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := "master"
		prReadyState := models.PullRequestReadyStateDraft
		tag := "0.9.0"
		format := "json"

		bitriseConfigPath := "bitrise.yml"
		bitriseConfigBase64Data := toBase64("bitrise.yml")

		inventoryPath := ".secrets.bitrise.yml"
		inventoryBase64Data := toBase64(".secrets.bitrise.yml")

		jsonParams := ""
		base64JSONParams := ""

		params, err := parseTriggerCheckParams(
			pattern,
			pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
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
		require.Equal(t, models.PullRequestReadyStateDraft, params.PRReadyState)
		require.Equal(t, tag, params.Tag)

		require.Equal(t, format, params.Format)

		require.Equal(t, bitriseConfigPath, params.BitriseConfigPath)
		require.Equal(t, bitriseConfigBase64Data, params.BitriseConfigBase64Data)

		require.Equal(t, inventoryPath, params.InventoryPath)
		require.Equal(t, inventoryBase64Data, params.InventoryBase64Data)
	}
}
