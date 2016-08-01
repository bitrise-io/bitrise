package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

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

	triggerPattern := ""

	inventoryBase64Data := ""
	inventoryPath := ""

	bitriseConfigBase64Data := ""
	bitriseConfigPath := ""

	params := map[string]string{}
	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	if jsonParams != "" {
		var err error
		params, err = parseJSONParams(jsonParams)
		if err != nil {
			return fmt.Errorf("Failed to parse json-params (%s), error: %s", jsonParams, err)
		}
	} else if jsonParamsBase64 != "" {
		var err error
		params, err = parseJSONParamsBase64(jsonParamsBase64)
		if err != nil {
			return fmt.Errorf("Failed to parse json-params (%s), error: %s", jsonParams, err)
		}
	}

	if len(params) > 0 {
		inventoryBase64Data = params[InventoryBase64Key]
		inventoryPath = params[InventoryKey]

		bitriseConfigBase64Data = params[ConfigBase64Key]
		bitriseConfigPath = params[ConfigKey]

		triggerPattern = params[PatternKey]
	} else {
		inventoryBase64Data = c.String(InventoryBase64Key)
		inventoryPath = c.String(InventoryKey)

		bitriseConfigBase64Data = c.String(ConfigBase64Key)
		bitriseConfigPath = c.String(ConfigKey)

		triggerPattern = c.String(PatternKey)
		if triggerPattern == "" && len(c.Args()) > 0 {
			triggerPattern = c.Args()[0]
		}

		deprecatedBitriseConfigPath := c.String(PathKey)
		if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
			log.Warn("'path' key is deprecated, use 'config' instead!")
			bitriseConfigPath = deprecatedBitriseConfigPath
		}
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Trigger filter validation
	if triggerPattern == "" {
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

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig.TriggerMap, triggerPattern, isPRMode)
	if err != nil {
		log.Fatalf("Failed to get workflow id by pattern, error: %s", err)
	}
	log.Infof("Pattern (%s) triggered workflow (%s) ", triggerPattern, workflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//

	return nil
}
