package cli

import (
	"encoding/base64"
	"encoding/json"
)

// --------------------
// Models
// --------------------

// RunAndTriggerParamsModel ...
type RunAndTriggerParamsModel struct {
	// Run Params
	WorkflowToRunID string `json:"workflow"`

	// Trigger Params
	TriggerPattern string `json:"pattern"`

	PushBranch     string `json:"push-branch"`
	PRSourceBranch string `json:"pr-source-branch"`
	PRTargetBranch string `json:"pr-target-branch"`
	IsDraftPR      bool   `json:"draft-pr"`
	Tag            string `json:"tag"`

	// Trigger Check Params
	Format string `json:"format"`

	// Bitrise Config Params
	BitriseConfigPath       string `json:"config"`
	BitriseConfigBase64Data string `json:"config-base64"`

	InventoryPath       string `json:"inventory"`
	InventoryBase64Data string `json:"inventory-base64"`
}

func parseRunAndTriggerJSONParams(jsonParams string) (RunAndTriggerParamsModel, error) {
	params := RunAndTriggerParamsModel{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return RunAndTriggerParamsModel{}, err
	}
	return params, nil
}

func parseRunAndTriggerParams(
	workflowToRunID,
	triggerPattern,
	pushBranch, prSourceBranch, prTargetBranch string, isDraftPR *bool, tag,
	format,
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (RunAndTriggerParamsModel, error) {
	params := RunAndTriggerParamsModel{}
	var err error

	// Parse json params if exist
	if jsonParams == "" && base64JSONParams != "" {
		jsonParamsBytes, err := base64.StdEncoding.DecodeString(base64JSONParams)
		if err != nil {
			return RunAndTriggerParamsModel{}, err
		}
		jsonParams = string(jsonParamsBytes)
	}

	if jsonParams != "" {
		params, err = parseRunAndTriggerJSONParams(jsonParams)
		if err != nil {
			return RunAndTriggerParamsModel{}, err
		}
	}

	// Override params
	if workflowToRunID != "" {
		params.WorkflowToRunID = workflowToRunID
	}

	if triggerPattern != "" {
		params.TriggerPattern = triggerPattern
	}

	if pushBranch != "" {
		params.PushBranch = pushBranch
	}
	if prSourceBranch != "" {
		params.PRSourceBranch = prSourceBranch
	}
	if prTargetBranch != "" {
		params.PRTargetBranch = prTargetBranch
	}
	if isDraftPR != nil {
		params.IsDraftPR = *isDraftPR
	}
	if tag != "" {
		params.Tag = tag
	}

	if format != "" {
		params.Format = format
	}

	if bitriseConfigPath != "" {
		params.BitriseConfigPath = bitriseConfigPath
	}
	if bitriseConfigBase64Data != "" {
		params.BitriseConfigBase64Data = bitriseConfigBase64Data
	}
	if inventoryPath != "" {
		params.InventoryPath = inventoryPath
	}
	if inventoryBase64Data != "" {
		params.InventoryBase64Data = inventoryBase64Data
	}

	return params, nil
}

func parseRunParams(
	workflowToRunID,
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (RunAndTriggerParamsModel, error) {
	return parseRunAndTriggerParams(workflowToRunID, "", "", "", "", nil, "", "", bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, jsonParams, base64JSONParams)
}

func parseTriggerParams(
	triggerPattern,
	pushBranch, prSourceBranch, prTargetBranch string, isDraftPR *bool, tag,
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (RunAndTriggerParamsModel, error) {
	return parseRunAndTriggerParams("", triggerPattern, pushBranch, prSourceBranch, prTargetBranch, isDraftPR, tag, "", bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, jsonParams, base64JSONParams)
}

func parseTriggerCheckParams(
	triggerPattern,
	pushBranch, prSourceBranch, prTargetBranch string, isDraftPR *bool, tag,
	format,
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (RunAndTriggerParamsModel, error) {
	return parseRunAndTriggerParams("", triggerPattern, pushBranch, prSourceBranch, prTargetBranch, isDraftPR, tag, format, bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, jsonParams, base64JSONParams)
}
