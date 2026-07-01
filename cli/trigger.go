package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/spf13/cobra"
)

var triggerCommand = &cobra.Command{
	Use:     "trigger",
	Aliases: []string{"t"},
	Short:   "Triggers a specified Workflow.",
	RunE:    trigger,
}

func init() {
	flags := triggerCommand.Flags()
	addTriggerFilterFlags(flags)
	addConfigAndInventoryFlags(flags)
	addSecretFilteringFlag(flags)
	addJSONParamsFlags(flags)
}

func printAvailableTriggerFilters(triggerMap []models.TriggerMapItemModel) {
	log.Info("The following trigger filters are available:")

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

func trigger(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	prGlobalFlagPtr, err := resolveBoolFlagOrEnv(cmd.Root().PersistentFlags(), PRKey)
	if err != nil {
		failf(err.Error())
	}
	ciGlobalFlagPtr, err := resolveBoolFlagOrEnv(cmd.Root().PersistentFlags(), CIKey)
	if err != nil {
		failf(err.Error())
	}
	secretFiltering, err := resolveBoolFlagOrEnv(cmd.Flags(), SecretFilteringKey)
	if err != nil {
		failf(err.Error())
	}
	secretEnvsFiltering, err := resolveBoolEnv(configs.IsSecretEnvsFilteringKey)
	if err != nil {
		failf("%s", err)
	}

	triggerPattern, _ := cmd.Flags().GetString(PatternKey)
	if triggerPattern == "" && len(args) > 0 {
		triggerPattern = args[0]
	}

	pushBranch, _ := cmd.Flags().GetString(PushBranchKey)
	prSourceBranch, _ := cmd.Flags().GetString(PRSourceBranchKey)
	prTargetBranch, _ := cmd.Flags().GetString(PRTargetBranchKey)
	prReadyStateStr, _ := cmd.Flags().GetString(PRReadyStateKey)
	prReadyState := models.PullRequestReadyState(prReadyStateStr)
	tag, _ := cmd.Flags().GetString(TagKey)

	bitriseConfigBase64Data, _ := cmd.Flags().GetString(ConfigBase64Key)
	bitriseConfigPath, _ := cmd.Flags().GetString(ConfigKey)

	inventoryBase64Data, _ := cmd.Flags().GetString(InventoryBase64Key)
	inventoryPath, _ := cmd.Flags().GetString(InventoryKey)

	jsonParams, _ := cmd.Flags().GetString(JSONParamsKey)
	jsonParamsBase64, _ := cmd.Flags().GetString(JSONParamsBase64Key)

	triggerParams, err := parseTriggerParams(
		triggerPattern,
		pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("failed to parse trigger command params, error: %s", err)
	}

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.InventoryBase64Data, triggerParams.InventoryPath)
	if err != nil {
		failf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath, bitrise.ValidationTypeMinimal)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		failf("Failed to create bitrise config, error: %s", err)
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
	isSecretFilteringMode, err := isSecretFiltering(secretFiltering, inventoryEnvironments)
	if err != nil {
		failf("Failed to check Secret Filtering mode, error: %s", err)
	}

	isSecretEnvsFilteringMode, err := isSecretEnvsFiltering(secretEnvsFiltering, inventoryEnvironments)
	if err != nil {
		failf("Failed to check Secret Envs Filtering mode, error: %s", err)
	}

	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		failf("Failed to check  PR mode, error: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		failf("Failed to check  CI mode, error: %s", err)
	}

	_, workflowToRunID, err := getPipelineAndWorkflowIDByParamsInCompatibleMode(bitriseConfig.TriggerMap, triggerParams, isPRMode)
	if err != nil {
		log.Errorf("Failed to get workflow id by pattern, error: %s", err)
		if strings.Contains(err.Error(), "no matching workflow found with trigger params:") {
			printAvailableTriggerFilters(bitriseConfig.TriggerMap)
		}
		os.Exit(1)
	}

	runConfig := RunConfig{
		Modes: models.WorkflowRunModes{
			CIMode:                  isCIMode,
			PRMode:                  isPRMode,
			DebugMode:               configs.IsDebugMode,
			SecretFilteringMode:     isSecretFilteringMode,
			SecretEnvsFilteringMode: isSecretEnvsFilteringMode,
			NoOutputTimeout:         0,
		},
		Config:   bitriseConfig,
		Workflow: workflowToRunID,
		Secrets:  inventoryEnvironments,
	}
	agentConfig, err := setupAgentConfig()
	if err != nil {
		failf("Failed to process agent config: %s", err)
	}

	runner := NewWorkflowRunner(runConfig, agentConfig, globalTracker)
	exitCode, err := runner.RunWorkflowsWithSetupAndCheckForUpdate()
	if err != nil {
		if err == errWorkflowRunFailed {
			msg := createWorkflowRunStatusMessage(exitCode)
			printWorkflowRunStatusMessage(msg)
			os.Exit(exitCode)
		}

		failf(err.Error())
	}

	msg := createWorkflowRunStatusMessage(0)
	printWorkflowRunStatusMessage(msg)
	os.Exit(0)

	return nil
}
