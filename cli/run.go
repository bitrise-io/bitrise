package cli

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/versions"
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

func aboutUtilityWorkflos() {
	log.Infoln("Note about utility workflows:")
	log.Infoln("Utility workflow names start with '_' (example: _my_utility_workflow),")
	log.Infoln(" these can't be triggered directly but can be used by other workflows")
	log.Infoln(" in the before_run and after_run blocks.")
}

func printAboutUtilityWorkflos() {
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
		aboutUtilityWorkflos()
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
	inventoryPath := c.String(InventoryKey)
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = path.Join(bitrise.CurrentDir, DefaultSecretsFileName)

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			log.Fatalln("Failed to check path:", err)
		} else if !exist {
			log.Debugln("[BITRISE_CLI] - No inventory yml found")
			inventoryPath = ""
		}
	} else {
		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			log.Fatalln("Failed to check path: ", err)
		} else if !exist {
			log.Fatalln("No inventory yml found")
		}
	}

	inventoryEnvironments := []envmanModels.EnvironmentItemModel{}
	if inventoryPath != "" {
		var err error
		inventoryEnvironments, err = bitrise.CollectEnvironmentsFromFile(inventoryPath)
		if err != nil {
			log.Fatalln("Invalid invetory format: ", err)
		}
	}

	// Config validation
	bitriseConfigPath, err := GetBitriseConfigFilePath(c)
	if err != nil {
		log.Fatalf("Failed to get config (bitrise.yml) path: %s", err)
	}
	if bitriseConfigPath == "" {
		log.Fatalln("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
	}

	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to read Workflow: ", err)
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI model version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI models's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
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
		printAboutUtilityWorkflos()
	}

	// Run selected configuration
	if _, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments); err != nil {
		log.Fatalln("Error: ", err)
	}
}
