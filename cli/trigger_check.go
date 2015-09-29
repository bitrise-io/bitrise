package cli

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
	"github.com/ryanuber/go-glob"
)

func registerFatal(errorMsg, format string) {
	msg := map[string]string{
		"error": errorMsg,
	}

	if format == OutputFormatRaw {
		log.Fatal(msg["error"])
	} else {
		bytes, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Faild to parse error model, err: %s", err)
		}

		fmt.Println(string(bytes))
		os.Exit(1)
	}
}

// GetWorkflowIDByPattern ...
func GetWorkflowIDByPattern(config models.BitriseDataModel, pattern string) (string, error) {
	// Check for workflow ID in trigger map
	for _, item := range config.TriggerMap {
		if glob.Glob(item.Pattern, pattern) {
			if !item.IsPullRequestAllowed && IsPullRequestMode {
				return "", fmt.Errorf("Trigger pattern (%s) match found, but pull request is not enabled", pattern)
			}
			return item.WorkflowID, nil
		}
	}
	return "", fmt.Errorf("Run triggered by pattern: (%s), but no matching workflow found", pattern)
}

func triggerCheck(c *cli.Context) {
	format := c.String(OuputFormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	// Config validation
	bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create config, err: %s", err), format)
	}

	// Trigger filter validation
	triggerPattern := ""
	if len(c.Args()) < 1 {
		registerFatal("No trigger pattern specified", format)
	} else {
		triggerPattern = c.Args()[0]
	}

	if triggerPattern == "" {
		registerFatal("No trigger pattern specified", format)
	}

	workflowToRunID, err := GetWorkflowIDByPattern(bitriseConfig, triggerPattern)
	if err != nil {
		registerFatal(fmt.Sprintf("Faild to select workflow by pattern (%s), err: %s", triggerPattern, err), format)
	}

	switch format {
	case OutputFormatRaw:
		fmt.Printf("%s -> %s\n", triggerPattern, colorstring.Blue(workflowToRunID))
		break
	case OutputFormatJSON:
		triggerModel := map[string]string{
			"pattern":  triggerPattern,
			"workflow": workflowToRunID,
		}
		bytes, err := json.Marshal(triggerModel)
		if err != nil {
			registerFatal(fmt.Sprintf("Faild to parse trigger model, err: %s", err), format)
		}

		fmt.Println(string(bytes))
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

}
