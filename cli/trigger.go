package cli

import (
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

func parseTriggerParams(
	triggerPattern,
	pushBranch, prSourceBranch, prTargetBranch,
	bitriseConfigPath, bitriseConfigBase64Data,
	inventoryPath, inventoryBase64Data,
	jsonParams, base64JSONParams string) (RunAndTriggerParamsModel, error) {
	return parseRunAndTriggerParams("", triggerPattern, pushBranch, prSourceBranch, prTargetBranch, "", bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, jsonParams, base64JSONParams)
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

	pushBranch := c.String(PushBranchKey)
	prSourceBranch := c.String(PRSourceBranchKey)
	prTargetBranch := c.String(PRTargetBranchKey)

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

	triggerParams, err := parseTriggerParams(
		triggerPattern,
		pushBranch, prSourceBranch, prTargetBranch,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("Failed to parse trigger command params, error: %s", err)
	}

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.InventoryBase64Data, triggerParams.InventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath)
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
