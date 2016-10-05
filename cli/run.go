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
	"github.com/urfave/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

	depManagerBrew     = "brew"
	depManagerTryCheck = "_"
)

// --------------------
// Utility
// --------------------

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
}

func runAndExit(bitriseConfig models.BitriseDataModel, inventoryEnvironments []envmanModels.EnvironmentItemModel, workflowToRunID string) {
	if workflowToRunID == "" {
		log.Fatal("No workflow id specified")
	}

	if !configs.CheckIsSetupWasDoneForVersion(version.VERSION) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(version.VERSION, false); err != nil {
			log.Fatalf("Setup failed, error: %s", err)
		}
	}

	startTime := time.Now()

	// Run selected configuration
	if buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments); err != nil {
		log.Fatalf("Failed to run workflow, error: %s", err)
	} else if buildRunResults.IsBuildFailed() {
		os.Exit(1)
	}
	os.Exit(0)
}

// --------------------
// CLI command
// --------------------

func run(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	//
	// Expand cli.Context
	var prGlobalFlagPtr *bool
	if c.IsSet(PRKey) {
		*prGlobalFlagPtr = c.Bool(PRKey)
	}

	var ciGlobalFlagPtr *bool
	if c.IsSet(CIKey) {
		*ciGlobalFlagPtr = c.Bool(CIKey)
	}

	workflowToRunID := c.String(WorkflowKey)
	if workflowToRunID == "" && len(c.Args()) > 0 {
		workflowToRunID = c.Args()[0]
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

	runParams, err := parseRunParams(
		workflowToRunID,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("Failed to parse command params, error: %s", err)
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(runParams.InventoryBase64Data, runParams.InventoryPath)
	if err != nil {
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(runParams.BitriseConfigBase64Data, runParams.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Workflow id validation
	if runParams.WorkflowToRunID == "" {
		// no workflow specified
		//  list all the available ones and then exit
		log.Error("No workfow specified!")
		printAvailableWorkflows(bitriseConfig)
		os.Exit(1)
	}
	if strings.HasPrefix(runParams.WorkflowToRunID, "_") {
		// util workflow specified
		//  print about util workflows and then exit
		printAboutUtilityWorkflows()
		os.Exit(1)
	}
	//

	//
	// Main
	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  PR mode, error: %s", err)
	}

	if err := registerPrMode(isPRMode); err != nil {
		log.Fatalf("Failed to register  PR mode, error: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  CI mode, error: %s", err)
	}

	if err := registerCIMode(isCIMode); err != nil {
		log.Fatalf("Failed to register  CI mode, error: %s", err)
	}

	log.Infoln(colorstring.Green("Running workflow:"), runParams.WorkflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, runParams.WorkflowToRunID)
	//

	return nil
}
