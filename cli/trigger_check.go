package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
)

var triggerCheckOpts struct {
	pattern           string
	pushBranch        string
	prSourceBranch    string
	prTargetBranch    string
	prReadyState      string
	tag               string
	configBase64      string
	config            string
	inventoryBase64   string
	inventory         string
	jsonParams        string
	jsonParamsBase64  string
	format            string
}

var triggerCheckCmd = &cobra.Command{
	Use:   "trigger-check",
	Short: "Prints out which workflow will triggered by specified pattern.",
	RunE:  triggerCheck,
}

func init() {
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.pattern, PatternKey, "", "trigger pattern.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.pushBranch, PushBranchKey, "", "Git push branch name.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.prSourceBranch, PRSourceBranchKey, "", "Git pull request source branch name.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.prTargetBranch, PRTargetBranchKey, "", "Git pull request target branch name.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.prReadyState, PRReadyStateKey, "", "Git pull request ready state. Options: ready_for_review draft converted_to_ready_for_review")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.tag, TagKey, "", "Git tag name.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.configBase64, ConfigBase64Key, "", "base64 encoded config data.")
	triggerCheckCmd.Flags().StringVarP(&triggerCheckOpts.config, ConfigKey, "c", "", "Path where the workflow config file is located.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.inventoryBase64, InventoryBase64Key, "", "base64 encoded inventory data.")
	triggerCheckCmd.Flags().StringVarP(&triggerCheckOpts.inventory, InventoryKey, "i", "", "Path of the inventory file.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.jsonParams, JSONParamsKey, "", "Specify command flags with json string-string hash.")
	triggerCheckCmd.Flags().StringVar(&triggerCheckOpts.jsonParamsBase64, JSONParamsBase64Key, "", "Specify command flags with base64 encoded json string-string hash.")
	triggerCheckCmd.Flags().StringVarP(&triggerCheckOpts.format, OuputFormatKey, "f", "", "Output format. Accepted: raw, json.")
}

// --------------------
// Utility
// --------------------

func registerFatal(errorMsg string, warnings []string, format string) {
	message := ValidationItemModel{
		IsValid:  (len(errorMsg) > 0),
		Error:    errorMsg,
		Warnings: warnings,
	}

	if format == output.FormatRaw {
		for _, warning := range message.Warnings {
			log.Warnf("warning: %s", warning)
		}
		failf(message.Error)
	} else {
		bytes, err := json.Marshal(message)
		if err != nil {
			failf("Failed to parse error model, error: %s", err)
		}

		log.Print(string(bytes))
		os.Exit(1)
	}
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

// --------------------
// CLI command
// --------------------

func triggerCheck(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	warnings := []string{}

	//
	// Expand flags
	var prGlobalFlagPtr *bool
	if cmd.Root().PersistentFlags().Changed(PRKey) {
		prGlobalFlagPtr = pointers.NewBoolPtr(prMode)
	}

	triggerPattern := triggerCheckOpts.pattern
	if triggerPattern == "" && len(args) > 0 {
		triggerPattern = args[0]
	}

	pushBranch := triggerCheckOpts.pushBranch
	prSourceBranch := triggerCheckOpts.prSourceBranch
	prTargetBranch := triggerCheckOpts.prTargetBranch
	prReadyState := models.PullRequestReadyState(triggerCheckOpts.prReadyState)
	tag := triggerCheckOpts.tag

	bitriseConfigBase64Data := triggerCheckOpts.configBase64
	bitriseConfigPath := triggerCheckOpts.config

	inventoryBase64Data := triggerCheckOpts.inventoryBase64
	inventoryPath := triggerCheckOpts.inventory

	jsonParams := triggerCheckOpts.jsonParams
	jsonParamsBase64 := triggerCheckOpts.jsonParamsBase64

	format := triggerCheckOpts.format

	triggerParams, err := parseTriggerCheckParams(
		triggerPattern,
		pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag,
		format,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to parse trigger check params, err: %s", err), warnings, triggerParams.Format)
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(triggerParams.InventoryBase64Data, triggerParams.InventoryPath)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create inventory, err: %s", err), warnings, triggerParams.Format)
	}

	// Config validation
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath, bitrise.ValidationTypeFull)
	warnings = append(warnings, warns...)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create config, err: %s", err), warnings, triggerParams.Format)
	}

	// Format validation
	if triggerParams.Format == "" {
		triggerParams.Format = output.FormatRaw
	} else if triggerParams.Format != output.FormatRaw && triggerParams.Format != output.FormatJSON {
		registerFatal(fmt.Sprintf("Invalid format: %s", triggerParams.Format), warnings, output.FormatJSON)
	}

	// Trigger filter validation
	if triggerParams.TriggerPattern == "" &&
		triggerParams.PushBranch == "" && triggerParams.PRSourceBranch == "" && triggerParams.PRTargetBranch == "" && triggerParams.Tag == "" {
		registerFatal("No trigger pattern nor trigger params specified", warnings, triggerParams.Format)
	}
	//

	//
	// Main
	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to check  PR mode, err: %s", err), warnings, triggerParams.Format)
	}

	pipelineToRunID, workflowToRunID, err := getPipelineAndWorkflowIDByParamsInCompatibleMode(bitriseConfig.TriggerMap, triggerParams, isPRMode)
	if err != nil {
		registerFatal(err.Error(), warnings, triggerParams.Format)
	}

	triggerModel := map[string]string{}
	if pipelineToRunID != "" {
		triggerModel["pipeline"] = pipelineToRunID
	}
	if workflowToRunID != "" {
		triggerModel["workflow"] = workflowToRunID
	}

	if triggerParams.TriggerPattern != "" {
		triggerModel["pattern"] = triggerParams.TriggerPattern
	} else {
		if triggerParams.PushBranch != "" {
			triggerModel["push-branch"] = triggerParams.PushBranch
		} else if triggerParams.PRSourceBranch != "" || triggerParams.PRTargetBranch != "" {
			if triggerParams.PRSourceBranch != "" {
				triggerModel["pr-source-branch"] = triggerParams.PRSourceBranch
			}
			if triggerParams.PRTargetBranch != "" {
				triggerModel["pr-target-branch"] = triggerParams.PRTargetBranch
			}
		} else if triggerParams.Tag != "" {
			triggerModel["tag"] = triggerParams.Tag
		}
	}

	switch triggerParams.Format {
	case output.FormatRaw:
		msg := ""
		for key, value := range triggerModel {
			if key == "pipeline" || key == "workflow" {
				msg = msg + fmt.Sprintf("-> %s", colorstring.Blue(value))
			} else {
				msg = fmt.Sprintf("%s: %s ", key, value) + msg
			}
		}
		log.Print(msg)
	case output.FormatJSON:
		bytes, err := json.Marshal(triggerModel)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to parse trigger model, err: %s", err), warnings, triggerParams.Format)
		}

		log.Print(string(bytes))
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", triggerParams.Format), warnings, output.FormatJSON)
	}
	//

	return nil
}
