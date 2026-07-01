package local

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/spf13/cobra"
)

// NewTriggerCommand ...
func NewTriggerCommand() *cobra.Command {
	triggerCommand := &cobra.Command{
		Use:     "trigger",
		Aliases: []string{"t"},
		Short:   "Triggers a specified Workflow.",
		RunE:    trigger,
	}

	flags := triggerCommand.Flags()
	cmdutil.AddTriggerFilterFlags(flags)
	cmdutil.AddConfigAndInventoryFlags(flags)
	cmdutil.AddSecretFilteringFlag(flags)
	cmdutil.AddJSONParamsFlags(flags)

	return triggerCommand
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
	cmdutil.LogCommandParameters(cmd)

	prGlobalFlagPtr, err := cmdutil.ResolveBoolFlagOrEnv(cmd.Root().PersistentFlags(), cmdutil.PRKey)
	if err != nil {
		failf("%s", err)
	}
	ciGlobalFlagPtr, err := cmdutil.ResolveBoolFlagOrEnv(cmd.Root().PersistentFlags(), cmdutil.CIKey)
	if err != nil {
		failf("%s", err)
	}
	secretFiltering, err := cmdutil.ResolveBoolFlagOrEnv(cmd.Flags(), cmdutil.SecretFilteringKey)
	if err != nil {
		failf("%s", err)
	}
	secretEnvsFiltering, err := cmdutil.ResolveBoolEnv(configs.IsSecretEnvsFilteringKey)
	if err != nil {
		cmdutil.Failf("%s", err)
	}

	triggerPattern, _ := cmd.Flags().GetString(cmdutil.PatternKey)
	if triggerPattern == "" && len(args) > 0 {
		triggerPattern = args[0]
	}

	pushBranch, _ := cmd.Flags().GetString(cmdutil.PushBranchKey)
	prSourceBranch, _ := cmd.Flags().GetString(cmdutil.PRSourceBranchKey)
	prTargetBranch, _ := cmd.Flags().GetString(cmdutil.PRTargetBranchKey)
	prReadyStateStr, _ := cmd.Flags().GetString(cmdutil.PRReadyStateKey)
	prReadyState := models.PullRequestReadyState(prReadyStateStr)
	tag, _ := cmd.Flags().GetString(cmdutil.TagKey)

	bitriseConfigBase64Data, _ := cmd.Flags().GetString(cmdutil.ConfigBase64Key)
	bitriseConfigPath, _ := cmd.Flags().GetString(cmdutil.ConfigKey)

	inventoryBase64Data, _ := cmd.Flags().GetString(cmdutil.InventoryBase64Key)
	inventoryPath, _ := cmd.Flags().GetString(cmdutil.InventoryKey)

	jsonParams, _ := cmd.Flags().GetString(cmdutil.JSONParamsKey)
	jsonParamsBase64, _ := cmd.Flags().GetString(cmdutil.JSONParamsBase64Key)

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
	inventoryEnvironments, err := cmdutil.CreateInventoryFromCLIParams(triggerParams.InventoryBase64Data, triggerParams.InventoryPath)
	if err != nil {
		cmdutil.Failf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := cmdutil.CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath, bitrise.ValidationTypeMinimal)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		cmdutil.Failf("Failed to create bitrise config, error: %s", err)
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
	isSecretFilteringMode, err := cmdutil.IsSecretFiltering(secretFiltering, inventoryEnvironments)
	if err != nil {
		cmdutil.Failf("Failed to check Secret Filtering mode, error: %s", err)
	}

	isSecretEnvsFilteringMode, err := cmdutil.IsSecretEnvsFiltering(secretEnvsFiltering, inventoryEnvironments)
	if err != nil {
		cmdutil.Failf("Failed to check Secret Envs Filtering mode, error: %s", err)
	}

	isPRMode, err := cmdutil.IsPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		cmdutil.Failf("Failed to check  PR mode, error: %s", err)
	}

	isCIMode, err := cmdutil.IsCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		cmdutil.Failf("Failed to check  CI mode, error: %s", err)
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
		cmdutil.Failf("Failed to process agent config: %s", err)
	}

	runner := NewWorkflowRunner(runConfig, agentConfig, cmdutil.Tracker())
	exitCode, err := runner.RunWorkflowsWithSetupAndCheckForUpdate()
	if err != nil {
		if err == errWorkflowRunFailed {
			msg := createWorkflowRunStatusMessage(exitCode)
			printWorkflowRunStatusMessage(msg)
			os.Exit(exitCode)
		}

		failf("%s", err)
	}

	msg := createWorkflowRunStatusMessage(0)
	printWorkflowRunStatusMessage(msg)
	os.Exit(0)

	return nil
}

func migratePatternToParams(params RunAndTriggerParamsModel, isPullRequestMode bool) RunAndTriggerParamsModel {
	if isPullRequestMode {
		params.PushBranch = ""
		params.PRSourceBranch = params.TriggerPattern
		params.PRTargetBranch = ""
		params.Tag = ""
	} else {
		params.PushBranch = params.TriggerPattern
		params.PRSourceBranch = ""
		params.PRTargetBranch = ""
		params.Tag = ""
	}

	params.TriggerPattern = ""

	return params
}

// migrates deprecated params.TriggerPattern to params.PushBranch or params.PRSourceBranch based on isPullRequestMode
// and returns the triggered workflow id
func getPipelineAndWorkflowIDByParamsInCompatibleMode(triggerMap models.TriggerMapModel, params RunAndTriggerParamsModel, isPullRequestMode bool) (string, string, error) {
	if params.TriggerPattern != "" {
		params = migratePatternToParams(params, isPullRequestMode)
	}

	return triggerMap.FirstMatchingTarget(params.PushBranch, params.PRSourceBranch, params.PRTargetBranch, params.PRReadyState, params.Tag)
}
