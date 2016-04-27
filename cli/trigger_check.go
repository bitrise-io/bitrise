package cli

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
	"github.com/ryanuber/go-glob"
)

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
			log.Fatalf("Failed to parse error model, err: %s", err)
		}

		fmt.Println(string(bytes))
		os.Exit(1)
	}
}

// GetWorkflowIDByPattern ...
func GetWorkflowIDByPattern(triggerMap []models.TriggerMapItemModel, pattern string, isPullRequestMode bool) (string, error) {
	matchFoundButPullRequestModeNotAllowed := false
	for _, item := range triggerMap {
		if glob.Glob(item.Pattern, pattern) {
			if isPullRequestMode && !item.IsPullRequestAllowed {
				matchFoundButPullRequestModeNotAllowed = true
				continue
			}
			return item.WorkflowID, nil
		}

	}
	if matchFoundButPullRequestModeNotAllowed {
		return "", fmt.Errorf("Run triggered by pattern: (%s) in pull request mode, but matching workflow disabled in pull request mode", pattern)
	}
	return "", fmt.Errorf("Run triggered by pattern: (%s), but no matching workflow found", pattern)
}

func triggerCheck(c *cli.Context) {
	warnings := []string{}

	//
	// Expand cli.Context
	prGlobalFlag := c.GlobalBool(PRKey)

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	format := c.String(OuputFormatKey)

	triggerPattern := ""
	if len(c.Args()) < 1 {
		registerFatal("No trigger pattern specified", warnings, format)
	} else {
		triggerPattern = c.Args()[0]
	}
	//

	//
	// Input validation
	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create inventory, err: %s", err), warnings, format)
	}

	// Config validation
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	warnings = append(warnings, warns...)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create config, err: %s", err), warnings, format)
	}

	// Format validation
	if format == "" {
		format = output.FormatRaw
	} else if !(format == output.FormatRaw || format == output.FormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	// Trigger filter validation
	if triggerPattern == "" {
		registerFatal("No trigger pattern specified", warnings, format)
	}
	//

	//
	// Main
	isPRMode, err := isPRMode(prGlobalFlag, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check  PR mode, err: %s", err)
	}

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig.TriggerMap, triggerPattern, isPRMode)
	if err != nil {
		registerFatal(err.Error(), warnings, format)
	}

	switch format {
	case output.FormatRaw:
		fmt.Printf("%s -> %s\n", triggerPattern, colorstring.Blue(workflowToRunID))
		break
	case output.FormatJSON:
		triggerModel := map[string]string{
			"pattern":  triggerPattern,
			"workflow": workflowToRunID,
		}
		bytes, err := json.Marshal(triggerModel)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to parse trigger model, err: %s", err), warnings, format)
		}

		fmt.Println(string(bytes))
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}
	//
}
