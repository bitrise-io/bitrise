package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

	depManagerBrew     = "brew"
	depManagerTryCheck = "_"
)

func aboutUtilityWorkflows() {
	log.Infoln("Note about utility workflows:")
	log.Infoln("Utility workflow names start with '_' (example: _my_utility_workflow),")
	log.Infoln(" these can't be triggered directly but can be used by other workflows")
	log.Infoln(" in the before_run and after_run blocks.")
}

func printAboutUtilityWorkflows() {
	log.Error("Utility workflows can't be triggered directly")
	fmt.Println()
	log.Infoln("Note about utility workflows:")
	log.Infoln("Utility workflow names start with '_' (example: _my_utility_workflow),")
	log.Infoln(" these can't be triggered directly but can be used by other workflows")
	log.Infoln(" in the before_run and after_run blocks.")
	os.Exit(1)
}

func printAvailableWorkflows(config models.BitriseDataModel) {
	workflowNames := []string{}
	utilityWorkflowNames := []string{}

	for wfName := range config.Workflows {
		if strings.HasPrefix(wfName, "_") {
			utilityWorkflowNames = append(utilityWorkflowNames, wfName)
		} else {
			workflowNames = append(workflowNames, wfName)
		}
	}
	sort.Strings(workflowNames)
	sort.Strings(utilityWorkflowNames)

	if len(workflowNames) > 0 {
		log.Infoln("The following workflows are available:")
		for _, wfName := range workflowNames {
			log.Infoln(" * " + wfName)
		}

		fmt.Println()
		log.Infoln("You can run a selected workflow with:")
		log.Infoln("-> bitrise run the-workflow-name")
		fmt.Println()
	} else {
		log.Infoln("No workflows are available!")
	}

	if len(utilityWorkflowNames) > 0 {
		log.Infoln("The following utility workflows also defined:")
		for _, wfName := range utilityWorkflowNames {
			log.Infoln(" * " + wfName)
		}

		fmt.Println()
		aboutUtilityWorkflows()
		fmt.Println()
	}

	os.Exit(1)
}

func run(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)
	log.Debugln("[BITRISE_CLI] - Run")

	if !bitrise.CheckIsSetupWasDoneForVersion(c.App.Version) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(c.App.Version, false); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	}

	startTime := time.Now()

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

	// Workflow validation
	workflowToRunID := ""
	if len(c.Args()) < 1 {
		log.Errorln("No workfow specified!")
	} else {
		workflowToRunID = c.Args()[0]
	}

	if workflowToRunID == "" {
		// no workflow specified
		//  list all the available ones and then exit
		printAvailableWorkflows(bitriseConfig)
	}
	if strings.HasPrefix(workflowToRunID, "_") {
		// util workflow specified
		//  print about util workflows and then exit
		printAboutUtilityWorkflows()
	}

	// Run selected configuration
	buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	sendAnonymizedAnalytics(buildRunResults)
}
