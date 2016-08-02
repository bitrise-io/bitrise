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

// RunOrTriggerParamsModel ...
type RunOrTriggerParamsModel struct {
	WorkflowToRunID string `json:"workflow"`
	TriggerPattern  string `json:"pattern"`

	InventoryBase64Data string `json:"inventory-base64"`
	InventoryPath       string `json:"inventory"`

	BitriseConfigBase64Data string `json:"config-base64"`
	BitriseConfigPath       string `json:"config"`
}

func parseRunOrTriggerJSONParams(jsonParams string) (RunOrTriggerParamsModel, error) {
	params := RunOrTriggerParamsModel{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return RunOrTriggerParamsModel{}, err
	}
	return params, nil
}

func parseRunOrTriggerBase64JSONParams(base64JSONParams string) (RunOrTriggerParamsModel, error) {
	jsonParamsBytes, err := base64.StdEncoding.DecodeString(base64JSONParams)
	if err != nil {
		return RunOrTriggerParamsModel{}, err
	}
	return parseRunOrTriggerJSONParams(string(jsonParamsBytes))
}

func parseRunOrTriggerParams(workflowToRunID, triggerPattern, inventoryBase64Data, inventoryPath, bitriseConfigBase64Data, bitriseConfigPath, jsonParams, base64JSONParams string) (RunOrTriggerParamsModel, error) {
	if jsonParams != "" {
		return parseRunOrTriggerJSONParams(jsonParams)
	} else if base64JSONParams != "" {
		return parseRunOrTriggerBase64JSONParams(base64JSONParams)
	} else {
		return RunOrTriggerParamsModel{
			WorkflowToRunID: workflowToRunID,
			TriggerPattern:  triggerPattern,

			InventoryBase64Data: inventoryBase64Data,
			InventoryPath:       inventoryPath,

			BitriseConfigBase64Data: bitriseConfigBase64Data,
			BitriseConfigPath:       bitriseConfigPath,
		}, nil
	}
}

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

func trigger(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	// Expand cli.Context
	prGlobalFlag := c.GlobalBool(PRKey)
	ciGlobalFlag := c.GlobalBool(CIKey)

	triggerPattern := c.String(PatternKey)
	if triggerPattern == "" && len(c.Args()) > 0 {
		triggerPattern = c.Args()[0]
	}

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

	params, err := parseRunOrTriggerParams(
		"", triggerPattern,
		inventoryBase64Data, inventoryPath,
		bitriseConfigBase64Data, bitriseConfigPath,
		jsonParams, jsonParamsBase64,
	)
	if err != nil {
		return fmt.Errorf("Failed to parse command params, error: %s", err)
	}

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(params.InventoryBase64Data, params.InventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(params.BitriseConfigBase64Data, params.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Trigger filter validation
	if params.TriggerPattern == "" {
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

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig.TriggerMap, params.TriggerPattern, isPRMode)
	if err != nil {
		log.Fatalf("Failed to get workflow id by pattern, error: %s", err)
	}
	log.Infof("Pattern (%s) triggered workflow (%s) ", params.TriggerPattern, workflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//

	return nil
}
