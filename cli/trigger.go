package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
)

var triggerOpts struct {
	pattern          string
	config           string
	inventory        string
	secretFiltering  bool
	pushBranch       string
	prSourceBranch   string
	prTargetBranch   string
	prReadyState     string
	tag              string
	jsonParams       string
	jsonParamsBase64 string
	configBase64     string
	inventoryBase64  string
}

var triggerCmd = &cobra.Command{
	Use:     "trigger",
	Aliases: []string{"t"},
	Short:   "Triggers a specified Workflow.",
	RunE:    trigger,
}

func init() {
	triggerCmd.Flags().StringVar(&triggerOpts.pattern, PatternKey, "", "trigger pattern.")
	triggerCmd.Flags().StringVarP(&triggerOpts.config, ConfigKey, configShortKey, "", "Path where the workflow config file is located.")
	triggerCmd.Flags().StringVarP(&triggerOpts.inventory, InventoryKey, inventoryShortKey, "", "Path of the inventory file.")
	triggerCmd.Flags().BoolVar(&triggerOpts.secretFiltering, secretFilteringFlag, false, "Hide secret values from the log.")

	triggerCmd.Flags().StringVar(&triggerOpts.pushBranch, PushBranchKey, "", "Git push branch name.")
	triggerCmd.Flags().StringVar(&triggerOpts.prSourceBranch, PRSourceBranchKey, "", "Git pull request source branch name.")
	triggerCmd.Flags().StringVar(&triggerOpts.prTargetBranch, PRTargetBranchKey, "", "Git pull request target branch name.")
	triggerCmd.Flags().StringVar(&triggerOpts.prReadyState, PRReadyStateKey, "", "Git pull request ready state. Options: ready_for_review draft converted_to_ready_for_review")
	triggerCmd.Flags().StringVar(&triggerOpts.tag, TagKey, "", "Git tag name.")

	triggerCmd.Flags().StringVar(&triggerOpts.jsonParams, JSONParamsKey, "", "Specify command flags with json string-string hash.")
	triggerCmd.Flags().StringVar(&triggerOpts.jsonParamsBase64, JSONParamsBase64Key, "", "Specify command flags with base64 encoded json string-string hash.")

	triggerCmd.Flags().StringVar(&triggerOpts.configBase64, ConfigBase64Key, "", "base64 encoded config data.")
	triggerCmd.Flags().StringVar(&triggerOpts.inventoryBase64, InventoryBase64Key, "", "base64 encoded inventory data.")
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

	// Expand flags
	var prGlobalFlagPtr *bool
	if cmd.Root().PersistentFlags().Changed(PRKey) {
		prGlobalFlagPtr = pointers.NewBoolPtr(prMode)
	}

	var ciGlobalFlagPtr *bool
	if cmd.Root().PersistentFlags().Changed(CIKey) {
		ciGlobalFlagPtr = pointers.NewBoolPtr(ciMode)
	}

	var secretFiltering *bool
	if cmd.Flags().Changed(secretFilteringFlag) {
		secretFiltering = pointers.NewBoolPtr(triggerOpts.secretFiltering)
	} else if os.Getenv(configs.IsSecretFilteringKey) == "true" {
		secretFiltering = pointers.NewBoolPtr(true)
	} else if os.Getenv(configs.IsSecretFilteringKey) == "false" {
		secretFiltering = pointers.NewBoolPtr(false)
	}

	var secretEnvsFiltering *bool
	if os.Getenv(configs.IsSecretEnvsFilteringKey) == "true" {
		secretEnvsFiltering = pointers.NewBoolPtr(true)
	} else if os.Getenv(configs.IsSecretEnvsFilteringKey) == "false" {
		secretEnvsFiltering = pointers.NewBoolPtr(false)
	}

	triggerPattern := triggerOpts.pattern
	if triggerPattern == "" && len(args) > 0 {
		triggerPattern = args[0]
	}

	pushBranch := triggerOpts.pushBranch
	prSourceBranch := triggerOpts.prSourceBranch
	prTargetBranch := triggerOpts.prTargetBranch
	prReadyState := models.PullRequestReadyState(triggerOpts.prReadyState)
	tag := triggerOpts.tag

	bitriseConfigBase64Data := triggerOpts.configBase64
	bitriseConfigPath := triggerOpts.config

	inventoryBase64Data := triggerOpts.inventoryBase64
	inventoryPath := triggerOpts.inventory

	jsonParams := triggerOpts.jsonParams
	jsonParamsBase64 := triggerOpts.jsonParamsBase64

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
