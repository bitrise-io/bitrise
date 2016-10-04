package cli

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

// --------------------
// Utility
// --------------------

func printAvailableTriggerFilters(triggerMap []models.TriggerMapItemModel) {
	log.Infoln("The following trigger filters are available:")

	for _, triggerItem := range triggerMap {
		if triggerItem.Pattern != "" {
			log.Infof(" * pattern: %s", triggerItem.Pattern)
			log.Infof("   is_pull_request_allowed: %v", triggerItem.IsPullRequestAllowed)
			log.Infof("   workflow: %s", triggerItem.WorkflowID)
		} else {
			if triggerItem.PushBranch != "" {
				log.Infof(" * push_branch: %s", triggerItem.PushBranch)
				log.Infof("   workflow: %s", triggerItem.WorkflowID)
			} else if triggerItem.PullRequestSourceBranch != "" || triggerItem.PullRequestTargetBranch != "" {
				log.Infof(" * pull_request_source_branch: %s", triggerItem.PullRequestSourceBranch)
				log.Infof("   pull_request_target_branch: %s", triggerItem.PullRequestTargetBranch)
				log.Infof("   workflow: %s", triggerItem.WorkflowID)
			} else if triggerItem.Tag != "" {
				log.Infof(" * tag: %s", triggerItem.Tag)
				log.Infof("   workflow: %s", triggerItem.WorkflowID)
			}
		}
	}
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
	tag := c.String(TagKey)

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
		pushBranch, prSourceBranch, prTargetBranch, tag,
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
	if triggerParams.TriggerPattern == "" &&
		triggerParams.PushBranch == "" && triggerParams.PRSourceBranch == "" && triggerParams.PRTargetBranch == "" && triggerParams.Tag == "" {
		log.Error("No trigger pattern nor trigger params specified")
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

	workflowToRunID, err := getWorkflowIDByParamsInCompatibleMode(bitriseConfig.TriggerMap, triggerParams, isPRMode)
	if err != nil {
		log.Errorf("Failed to get workflow id by pattern, error: %s", err)
		if strings.Contains(err.Error(), "no matching workflow found with trigger params:") {
			printAvailableTriggerFilters(bitriseConfig.TriggerMap)
		}
		os.Exit(1)
	}

	if triggerParams.TriggerPattern != "" {
		log.Infof("pattern (%s) triggered workflow (%s)", triggerParams.TriggerPattern, workflowToRunID)
	} else {
		if triggerParams.PushBranch != "" {
			log.Infof("push-branch (%s) triggered workflow (%s)", triggerParams.PushBranch, workflowToRunID)
		} else if triggerParams.PRSourceBranch != "" || triggerParams.PRTargetBranch != "" {
			log.Infof("pr-source-branch (%s) and pr-target-branch (%s) triggered workflow (%s)", triggerParams.PRSourceBranch, triggerParams.PRTargetBranch, workflowToRunID)
		} else if triggerParams.Tag != "" {
			log.Infof("tag (%s) triggered workflow (%s)", triggerParams.Tag, workflowToRunID)
		}
	}

	runAndExit(bitriseConfig, inventoryEnvironments, workflowToRunID)
	//

	return nil
}
