package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/ryanuber/go-glob"
	"github.com/urfave/cli"
)

// --------------------
// Models
// --------------------

// BitriseConfigParamsModel ...
type BitriseConfigParamsModel struct {
	BitriseConfigPath       string `json:"config"`
	BitriseConfigBase64Data string `json:"config-base64"`

	InventoryPath       string `json:"inventory"`
	InventoryBase64Data string `json:"inventory-base64"`
}

func parseBitriseConfigJSONParams(jsonParams string) (BitriseConfigParamsModel, error) {
	params := BitriseConfigParamsModel{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return BitriseConfigParamsModel{}, err
	}
	return params, nil
}

func parseBitriseConfigParams(
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (BitriseConfigParamsModel, error) {

	params := BitriseConfigParamsModel{}
	var err error

	// Parse json params if exist
	if jsonParams == "" && base64JSONParams != "" {
		jsonParamsBytes, err := base64.StdEncoding.DecodeString(base64JSONParams)
		if err != nil {
			return BitriseConfigParamsModel{}, err
		}
		jsonParams = string(jsonParamsBytes)
	}

	if jsonParams != "" {
		params, err = parseBitriseConfigJSONParams(jsonParams)
		if err != nil {
			return BitriseConfigParamsModel{}, err
		}
	}

	// Owerride params
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

// --------------------
// Utility
// --------------------

func registerFatal(errorMsg string, warnings []string, format string) {
	message := ValidationItemModel{
		IsValid:  (len(errorMsg) > 0),
		Error:    errorMsg,
		Warnings: warnings,
	}

	if format == output.FormatRaw {
		for _, warning := range message.Warnings {
			log.Warnf("warning: %s", warning)
		}
		log.Fatal(message.Error)
	} else {
		bytes, err := json.Marshal(message)
		if err != nil {
			log.Fatalf("Failed to parse error model, error: %s", err)
		}

		fmt.Println(string(bytes))
		os.Exit(1)
	}
}

func validateTriggerMap(triggerMap []models.TriggerMapItemModel) error {
	for _, item := range triggerMap {
		if item.Pattern == "" {
			return fmt.Errorf("invalid trigger item: (%s) -> (%s), error: empty pattern", item.Pattern, item.WorkflowID)
		}

		if item.WorkflowID == "" {
			return fmt.Errorf("invalid trigger item: (%s) -> (%s), error: empty workflow id", item.Pattern, item.WorkflowID)
		}
	}

	return nil
}

func getWorkflowIDByPattern(triggerMap []models.TriggerMapItemModel, pattern string, isPullRequestMode bool) (string, error) {
	if err := validateTriggerMap(triggerMap); err != nil {
		return "", err
	}

	matchFoundButPullRequestModeNotAllowed := false
	for _, item := range triggerMap {
		if glob.Glob(item.Pattern, pattern) {
			if isPullRequestMode && !item.IsPullRequestAllowed {
				matchFoundButPullRequestModeNotAllowed = true
				continue
			}
			return item.WorkflowID, nil
		}

	}
	if matchFoundButPullRequestModeNotAllowed {
		return "", fmt.Errorf("Run triggered by pattern: (%s) in pull request mode, but matching workflow disabled in pull request mode", pattern)
	}
	return "", fmt.Errorf("Run triggered by pattern: (%s), but no matching workflow found", pattern)
}

// --------------------
// CLI command
// --------------------

func triggerCheck(c *cli.Context) error {
	warnings := []string{}

	//
	// Expand cli.Context
	prGlobalFlag := c.GlobalBool(PRKey)

	triggerPattern := c.String(PatternKey)
	if triggerPattern == "" && len(c.Args()) > 0 {
		triggerPattern = c.Args()[0]
	}

	gitTriggerEvent := c.String(GitTriggerEventKey)
	sourceBranch := c.String(SourceBranchKey)
	targetBranch := c.String(TargetBranchKey)

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	format := c.String(OuputFormatKey)

	triggerParams, err := parseTriggerCommandParams(
		triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		format,
		jsonParams, jsonParamsBase64)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to parse trigger check params, err: %s", err), warnings, triggerParams.Format)
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.BitriseConfigParams.InventoryBase64Data, triggerParams.BitriseConfigParams.InventoryPath)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create inventory, err: %s", err), warnings, triggerParams.Format)
	}

	// Config validation
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigParams.BitriseConfigBase64Data, triggerParams.BitriseConfigParams.BitriseConfigPath)
	warnings = append(warnings, warns...)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create config, err: %s", err), warnings, triggerParams.Format)
	}

	// Format validation
	if triggerParams.Format == "" {
		triggerParams.Format = output.FormatRaw
	} else if !(triggerParams.Format == output.FormatRaw || triggerParams.Format == output.FormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", triggerParams.Format), warnings, output.FormatJSON)
	}

	// Trigger filter validation
	if triggerParams.TriggerPattern == "" {
		registerFatal("No trigger pattern specified", warnings, triggerParams.Format)
	}
	//

	//
	// Main
	isPRMode, err := isPRMode(prGlobalFlag, inventoryEnvironments)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to check  PR mode, err: %s", err), warnings, triggerParams.Format)
	}

	workflowToRunID, err := getWorkflowIDByPattern(bitriseConfig.TriggerMap, triggerParams.TriggerPattern, isPRMode)
	if err != nil {
		registerFatal(err.Error(), warnings, triggerParams.Format)
	}

	switch triggerParams.Format {
	case output.FormatRaw:
		fmt.Printf("%s -> %s\n", triggerParams.TriggerPattern, colorstring.Blue(workflowToRunID))
		break
	case output.FormatJSON:
		triggerModel := map[string]string{
			"pattern":  triggerParams.TriggerPattern,
			"workflow": workflowToRunID,
		}
		bytes, err := json.Marshal(triggerModel)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to parse trigger model, err: %s", err), warnings, triggerParams.Format)
		}

		fmt.Println(string(bytes))
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", triggerParams.Format), warnings, output.FormatJSON)
	}
	//

	return nil
}
