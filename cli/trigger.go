package cli

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
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

	os.Exit(1)
}

func trigger(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	if !bitrise.CheckIsSetupWasDoneForVersion(c.App.Version) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(c.App.Version, false); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	}

	startTime := time.Now()

	// ------------------------
	// Input validation

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(c)
	if err != nil {
		log.Fatalf("Failed to create inventory, err: %s", err)
	}
	if err := checkCIAndPRModeFromSecrets(inventoryEnvironments); err != nil {
		log.Fatalf("Failed to check  PR and CI mode, err: %s", err)
	}

	// Config validation
	bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
	if err != nil {
		log.Fatalf("Failed to create bitrise config, err: %s", err)
	}

	// Trigger filter validation
	triggerPattern := ""
	if len(c.Args()) < 1 {
		log.Errorln("No workfow specified!")
	} else {
		triggerPattern = c.Args()[0]
	}

	if triggerPattern == "" {
		// no trigger filter specified
		//  list all the available ones and then exit
		printAvailableTriggerFilters(bitriseConfig.TriggerMap)
	}

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig, triggerPattern)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Infof("Pattern (%s) triggered workflow (%s) ", triggerPattern, workflowToRunID)

	// Run selected configuration
	buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	sendAnonymizedAnalytics(buildRunResults)
}
