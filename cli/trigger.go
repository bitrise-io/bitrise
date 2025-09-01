package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/urfave/cli"
)

var triggerCommand = cli.Command{
	Name:    "trigger",
	Aliases: []string{"t"},
	Usage:   "Triggers a specified Workflow.",
	Action:  trigger,
	Flags: []cli.Flag{
		// cli params
		cli.StringFlag{Name: PatternKey, Usage: "trigger pattern."},
		cli.StringFlag{Name: ConfigKey + ", " + configShortKey, Usage: "Path where the workflow config file is located."},
		cli.StringFlag{Name: InventoryKey + ", " + inventoryShortKey, Usage: "Path of the inventory file."},
		cli.BoolFlag{Name: secretFilteringFlag, Usage: "Hide secret values from the log.", EnvVar: configs.IsSecretFilteringKey},

		cli.StringFlag{Name: PushBranchKey, Usage: "Git push branch name."},
		cli.StringFlag{Name: PRSourceBranchKey, Usage: "Git pull request source branch name."},
		cli.StringFlag{Name: PRTargetBranchKey, Usage: "Git pull request target branch name."},
		cli.StringFlag{Name: PRReadyStateKey, Usage: "Git pull request ready state. Options: ready_for_review draft converted_to_ready_for_review"},
		cli.StringFlag{Name: TagKey, Usage: "Git tag name."},

		// cli params used in CI mode
		cli.StringFlag{Name: JSONParamsKey, Usage: "Specify command flags with json string-string hash."},
		cli.StringFlag{Name: JSONParamsBase64Key, Usage: "Specify command flags with base64 encoded json string-string hash."},

		// should deprecate
		cli.StringFlag{Name: ConfigBase64Key, Usage: "base64 encoded config data."},
		cli.StringFlag{Name: InventoryBase64Key, Usage: "base64 encoded inventory data."},
	},
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

func trigger(c *cli.Context) error {
	logCommandParameters(c)

	// Expand cli.Context
	var prGlobalFlagPtr *bool
	if c.GlobalIsSet(PRKey) {
		prGlobalFlagPtr = pointers.NewBoolPtr(c.GlobalBool(PRKey))
	}

	var ciGlobalFlagPtr *bool
	if c.GlobalIsSet(CIKey) {
		ciGlobalFlagPtr = pointers.NewBoolPtr(c.GlobalBool(CIKey))
	}

	var secretFiltering *bool
	if c.IsSet(secretFilteringFlag) {
		secretFiltering = pointers.NewBoolPtr(c.Bool(secretFilteringFlag))
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

	triggerPattern := c.String(PatternKey)
	if triggerPattern == "" && len(c.Args()) > 0 {
		triggerPattern = c.Args()[0]
	}

	pushBranch := c.String(PushBranchKey)
	prSourceBranch := c.String(PRSourceBranchKey)
	prTargetBranch := c.String(PRTargetBranchKey)
	prReadyState := models.PullRequestReadyState(c.String(PRReadyStateKey))
	tag := c.String(TagKey)

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	bitriseConfigPath := c.String(ConfigKey)

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	triggerParams, err := parseTriggerParams(
		triggerPattern,
		pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("Failed to parse trigger command params, error: %s", err)
	}

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.InventoryBase64Data, triggerParams.InventoryPath)
	if err != nil {
		failf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath)
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
		failf("Failed to process agent config: %w", err)
	}

	runner := NewWorkflowRunner(runConfig, agentConfig)
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
