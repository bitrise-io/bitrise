package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	envmanModels "github.com/bitrise-io/envman/models"
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

func printAboutUtilityWorkflowsAndExit() {
	log.Error("Utility workflows can't be triggered directly")
	fmt.Println()
	log.Infoln("Note about utility workflows:")
	log.Infoln("Utility workflow names start with '_' (example: _my_utility_workflow),")
	log.Infoln(" these can't be triggered directly but can be used by other workflows")
	log.Infoln(" in the before_run and after_run blocks.")
	os.Exit(1)
}

func printAvailableWorkflowsAndExit(config models.BitriseDataModel) {
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

func runAndExit(bitriseConfig models.BitriseDataModel, inventoryEnvironments []envmanModels.EnvironmentItemModel, workflowToRunID string) {
	if workflowToRunID == "" {
		log.Fatal("No workflow id specified")
	}

	if !configs.CheckIsSetupWasDoneForVersion(version.VERSION) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(version.VERSION, false); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	}

	startTime := time.Now()

	// Run selected configuration
	if buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments); err != nil {
		log.Fatalln("Error: ", err)
	} else if buildRunResults.IsBuildFailed() {
		os.Exit(1)
	}
	os.Exit(0)
}

func run(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	//
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

	workflowToRunID := ""
	if len(c.Args()) < 1 {
		log.Fatal("No workfow specified!")
	} else {
		workflowToRunID = c.Args()[0]
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

	// Workflow id validation
	if workflowToRunID == "" {
		// no workflow specified
		//  list all the available ones and then exit
		printAvailableWorkflowsAndExit(bitriseConfig)
	}
	if strings.HasPrefix(workflowToRunID, "_") {
		// util workflow specified
		//  print about util workflows and then exit
		printAboutUtilityWorkflowsAndExit()
	}
	//

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

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//
}
