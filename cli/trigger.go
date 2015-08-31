package cli

import (
	"fmt"
	"os"
	"path"
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

	secretEnvironments := []envmanModels.EnvironmentItemModel{}
	if inventoryPath != "" {
		var err error
		secretEnvironments, err = bitrise.CollectEnvironmentsFromFile(inventoryPath)
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
		log.Warn("bitrise CLI version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI models's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
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

	pullRequestID := os.Getenv(bitrise.PullRequestIDEnvKey)
	workflowToRunID, err := bitriseConfig.WorkflowIDByPattern(triggerPattern, pullRequestID)
	if err != nil {
		log.Fatalf("Faild to select workflow by filter (%s), err: %s", triggerPattern, err)
	}
	log.Info("Triggered workflow id: ", workflowToRunID)

	// Run selected configuration
	if _, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, secretEnvironments); err != nil {
		log.Fatalln("Error: ", err)
	}
}
