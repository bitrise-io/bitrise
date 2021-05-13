package cli

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/urfave/cli"
)

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
		log.Fatal(message.Error)
	} else {
		bytes, err := json.Marshal(message)
		if err != nil {
			log.Fatalf("Failed to parse error model, error: %s", err)
		}

		fmt.Println(string(bytes))
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

func getPipelineAndWorkflowIDByParams(triggerMap models.TriggerMapModel, params RunAndTriggerParamsModel) (string, string, error) {
	for _, item := range triggerMap {
		match, err := item.MatchWithParams(params.PushBranch, params.PRSourceBranch, params.PRTargetBranch, params.Tag)
		if err != nil {
			return "", "", err
		}
		if match {
			return item.PipelineID, item.WorkflowID, nil
		}
	}

	return "", "", fmt.Errorf("no matching pipeline & workflow found with trigger params: push-branch: %s, pr-source-branch: %s, pr-target-branch: %s, tag: %s", params.PushBranch, params.PRSourceBranch, params.PRTargetBranch, params.Tag)
}

// migrates deprecated params.TriggerPattern to params.PushBranch or params.PRSourceBranch based on isPullRequestMode
// and returns the triggered workflow id
func getPipelineAndWorkflowIDByParamsInCompatibleMode(triggerMap models.TriggerMapModel, params RunAndTriggerParamsModel, isPullRequestMode bool) (string, string, error) {
	if params.TriggerPattern != "" {
		params = migratePatternToParams(params, isPullRequestMode)
	}

	return getPipelineAndWorkflowIDByParams(triggerMap, params)
}

// --------------------
// CLI command
// --------------------

func triggerCheck(c *cli.Context) error {
	warnings := []string{}

	//
	// Expand cli.Context
	var prGlobalFlagPtr *bool
	if c.GlobalIsSet(PRKey) {
		prGlobalFlagPtr = pointers.NewBoolPtr(c.GlobalBool(PRKey))
	}

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
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	format := c.String(OuputFormatKey)

	triggerParams, err := parseTriggerCheckParams(
		triggerPattern,
		pushBranch, prSourceBranch, prTargetBranch, tag,
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
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(triggerParams.BitriseConfigBase64Data, triggerParams.BitriseConfigPath)
	warnings = append(warnings, warns...)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create config, err: %s", err), warnings, triggerParams.Format)
	}

	// Format validation
	if triggerParams.Format == "" {
		triggerParams.Format = output.FormatRaw
	} else if !(triggerParams.Format == output.FormatRaw || triggerParams.Format == output.FormatJSON) {
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
		fmt.Println(msg)
		break
	case output.FormatJSON:
		bytes, err := json.Marshal(triggerModel)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to parse trigger model, err: %s", err), warnings, triggerParams.Format)
		}

		fmt.Println(string(bytes))
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", triggerParams.Format), warnings, output.FormatJSON)
	}
	//

	return nil
}
