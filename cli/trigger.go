package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/codegangsta/cli"
)

func printAvailableTriggerFiltersAndExit(triggerMap []models.TriggerMapItemModel) {
	log.Infoln("The following trigger filters are available:")
	for _, triggerItem := range triggerMap {
		log.Infoln(" * " + triggerItem.Pattern)
	}

	fmt.Println()
	log.Infoln("You can trigger a workflow with:")
	log.Infoln("-> bitrise trigger the-trigger-filter")
	fmt.Println()

	os.Exit(1)
}

func trigger(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	// Expand cli.Context
	prGlobalFlag := c.GlobalBool(PRKey)
	ciGlobalFlag := c.GlobalBool(CIKey)

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		log.Warn("'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	triggerPattern := ""
	if len(c.Args()) < 1 {
		log.Fatal("No pattern specified!")
	} else {
		triggerPattern = c.Args()[0]
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, err: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, err: %s", err)
	}

	// Trigger filter validation
	if triggerPattern == "" {
		// no trigger filter specified
		//  list all the available ones and then exit
		printAvailableTriggerFiltersAndExit(bitriseConfig.TriggerMap)
	}
	//

	// Main
	isPRMode, err := isPRMode(prGlobalFlag, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  PR mode, err: %s", err)
	}

	if err := registerPrMode(isPRMode); err != nil {
		log.Fatalf("Failed to register  PR mode, err: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlag, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  CI mode, err: %s", err)
	}

	if err := registerCIMode(isCIMode); err != nil {
		log.Fatalf("Failed to register  CI mode, err: %s", err)
	}

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig.TriggerMap, triggerPattern, isPRMode)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Infof("Pattern (%s) triggered workflow (%s) ", triggerPattern, workflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//
}
