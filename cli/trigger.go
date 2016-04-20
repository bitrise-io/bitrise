package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
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
	bitriseConfig, _, err := CreateBitriseConfigFromCLIParams(c)
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

	runAndExit(c, workflowToRunID)
}
