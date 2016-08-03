package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

// --------------------
// Models
// --------------------

// TriggerParamsModel ...
type TriggerParamsModel struct {
	TriggerPattern  string `json:"pattern"`
	GitTriggerEvent string `json:"event"`
	SourceBranch    string `json:"source-branch"`
	TargetBranch    string `json:"target-branch"`

	Format string `json:"format"`

	BitriseConfigParams BitriseConfigParamsModel
}

func parseTriggerJSONParams(jsonParams string) (TriggerParamsModel, error) {
	params := TriggerParamsModel{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return TriggerParamsModel{}, err
	}
	return params, nil
}

func parseTriggerParams(
	triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
	format,
	jsonParams, base64JSONParams string) (TriggerParamsModel, error) {
	params := TriggerParamsModel{}
	var err error

	// Parse json params if exist
	if jsonParams == "" && base64JSONParams != "" {
		jsonParamsBytes, err := base64.StdEncoding.DecodeString(base64JSONParams)
		if err != nil {
			return TriggerParamsModel{}, err
		}
		jsonParams = string(jsonParamsBytes)
	}

	if jsonParams != "" {
		params, err = parseTriggerJSONParams(jsonParams)
		if err != nil {
			return TriggerParamsModel{}, err
		}
	}

	// Owerride params
	if triggerPattern != "" {
		params.TriggerPattern = triggerPattern
	}
	if gitTriggerEvent != "" {
		params.GitTriggerEvent = gitTriggerEvent
	}
	if sourceBranch != "" {
		params.SourceBranch = sourceBranch
	}
	if targetBranch != "" {
		params.TargetBranch = targetBranch
	}
	if format != "" {
		params.Format = format
	}

	return params, nil
}

func parseTriggerCommandParams(
	triggerPattern, gitTriggerEvent, sourceBranch, targetBranch, // trigger params
	bitriseConfigPath, bitriseConfigBase64Data, // bitrise config params
	inventoryPath, inventoryBase64Data,
	format,
	jsonParams, base64JSONParams string) (TriggerParamsModel, error) { // json params

	bitriseConfigParams, err := parseBitriseConfigParams(bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, jsonParams, base64JSONParams)
	if err != nil {
		return TriggerParamsModel{}, err
	}

	triggerParams, err := parseTriggerParams(triggerPattern, gitTriggerEvent, sourceBranch, targetBranch, format, jsonParams, base64JSONParams)
	if err != nil {
		return TriggerParamsModel{}, err
	}

	triggerParams.BitriseConfigParams = bitriseConfigParams

	return triggerParams, nil
}

// --------------------
// Utility
// --------------------

func printAvailableTriggerFilters(triggerMap []models.TriggerMapItemModel) {
	log.Infoln("The following trigger filters are available:")
	for _, triggerItem := range triggerMap {
		log.Infoln(" * " + triggerItem.Pattern)
	}

	fmt.Println()
	log.Infoln("You can trigger a workflow with:")
	log.Infoln("-> bitrise trigger the-trigger-filter")
	fmt.Println()
}

// --------------------
// CLI command
// --------------------

func trigger(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	// Expand cli.Context
	prGlobalFlag := c.GlobalBool(PRKey)
	ciGlobalFlag := c.GlobalBool(CIKey)

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
		log.Warn("'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	triggerParams, err := parseTriggerCommandParams(
		triggerPattern, gitTriggerEvent, sourceBranch, targetBranch,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		"",
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("Failed to parse trigger command params, error: %s", err)
	}

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.BitriseConfigParams.InventoryBase64Data, triggerParams.BitriseConfigParams.InventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigParams.BitriseConfigBase64Data, triggerParams.BitriseConfigParams.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Trigger filter validation
	if triggerParams.TriggerPattern == "" {
		// no trigger filter specified
		//  list all the available ones and then exit
		log.Error("No pattern specified!")
		printAvailableTriggerFilters(bitriseConfig.TriggerMap)
		os.Exit(1)
	}
	//

	// Main
	isPRMode, err := isPRMode(prGlobalFlag, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  PR mode, error: %s", err)
	}

	if err := registerPrMode(isPRMode); err != nil {
		log.Fatalf("Failed to register  PR mode, error: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlag, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  CI mode, error: %s", err)
	}

	if err := registerCIMode(isCIMode); err != nil {
		log.Fatalf("Failed to register  CI mode, error: %s", err)
	}

	workflowToRunID, err := getWorkflowIDByPattern(bitriseConfig.TriggerMap, triggerParams.TriggerPattern, isPRMode)
	if err != nil {
		log.Fatalf("Failed to get workflow id by pattern, error: %s", err)
	}
	log.Infof("Pattern (%s) triggered workflow (%s) ", triggerParams.TriggerPattern, workflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//

	return nil
}
