package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

var workflowListCommand = cli.Command{
	Name:  "workflows",
	Usage: "List of available workflows in config.",
	Action: func(c *cli.Context) error {
		if err := workflowList(c); err != nil {
			log.Errorf("List of available workflows in config failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		flPath,
		flConfig,
		flConfigBase64,
		cli.StringFlag{
			Name:  "format",
			Usage: "Output format. Accepted: raw, json.",
		},
		cli.BoolFlag{
			Name:  "minimal",
			Usage: "Print summary of workflows only.",
		},
		cli.BoolFlag{
			Name:  "id-only",
			Usage: "Print workflow ids only.",
		},
	},
}

// WorkflowListOutputModel ...
type WorkflowListOutputModel struct {
	Data     map[string]map[string]string `json:"data,omitempty" yml:"data,omitempty"`
	Warnings []string                     `json:"warnings,omitempty" yml:"warnings,omitempty"`
	Error    string                       `json:"error,omitempty" yml:"error,omitempty"`
}

// NewOutput ...
func NewOutput(data map[string]map[string]string, warnings ...string) WorkflowListOutputModel {
	return WorkflowListOutputModel{
		Data:     data,
		Warnings: warnings,
	}
}

// NewErrorOutput ...
func NewErrorOutput(err string, warnings ...string) WorkflowListOutputModel {
	return WorkflowListOutputModel{
		Error:    err,
		Warnings: warnings,
	}
}

func printableRawWorkflow(id string, info map[string]string) string {
	message := ""
	message += fmt.Sprintf("⚡️ %s\n", colorstring.Green(id))
	message += fmt.Sprintf("  %s: %s\n", colorstring.Yellow("Title"), info["title"])
	message += fmt.Sprintf("  %s: %s\n", colorstring.Yellow("Summary"), info["summary"])

	if info["description"] != "" {
		message += fmt.Sprintf("  %s: %s\n", colorstring.Yellow("Description"), info["description"])
	}
	message += fmt.Sprintf("  %s: bitrise run %s\n", colorstring.Yellow("Run with"), id)
	message += "\n"
	return message
}

// String ...
func (output WorkflowListOutputModel) String() string {
	message := ""
	for _, warning := range output.Warnings {
		message += colorstring.Yellow(warning) + "\n"
	}
	if output.Error != "" {
		message += colorstring.Red(output.Error) + "\n"
		return message
	}

	workflowIDs := []string{}
	utilityWorkflowIDs := []string{}

	idOnly := false
	minimal := true
	for id, info := range output.Data {
		if strings.HasPrefix(id, "_") {
			utilityWorkflowIDs = append(utilityWorkflowIDs, id)
		} else {
			workflowIDs = append(workflowIDs, id)
		}

		if !idOnly && (info == nil || len(info) == 0) {
			idOnly = true
		}
		if minimal && info["description"] != "" {
			minimal = false
		}
	}

	sort.Strings(workflowIDs)
	sort.Strings(utilityWorkflowIDs)

	if idOnly {
		return strings.Join(workflowIDs, " ")
	}

	if len(workflowIDs) > 0 {
		message += "Workflows\n"
		message += "---------\n"
		for _, id := range workflowIDs {
			workflow := output.Data[id]
			message += printableRawWorkflow(id, workflow)
		}
	}

	if len(utilityWorkflowIDs) > 0 {
		message += "Util Workflows\n"
		message += "--------------\n"
		for _, id := range utilityWorkflowIDs {
			workflow := output.Data[id]
			message += printableRawWorkflow(id, workflow)
		}
	}

	if len(workflowIDs) == 0 && len(utilityWorkflowIDs) == 0 {
		message += colorstring.Red("Config doesn't contain any workflow")
	}

	return message
}

// JSON ...
func (output WorkflowListOutputModel) JSON() string {
	workflowIDs := []string{}
	idOnly := false
	for id, info := range output.Data {
		if !strings.HasPrefix(id, "_") {
			workflowIDs = append(workflowIDs, id)
		}

		if !idOnly && (info == nil || len(info) == 0) {
			idOnly = true
		}
	}

	var toMarshal interface{}
	if idOnly {
		toMarshal = workflowIDs
	} else {
		toMarshal = output
	}

	data, err := json.MarshalIndent(toMarshal, "", "\t")
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(data) + "\n"
}

func workflowList(c *cli.Context) error {
	// Expand cli.Context
	bitriseConfigBase64Data := c.String("config-base64")
	bitriseConfigPath := c.String("config")
	deprecatedBitriseConfigPath := c.String("path")

	format := c.String("format")
	minimal := c.Bool("minimal")
	idOnly := c.Bool("id-only")

	// Input validation
	if format == "" {
		format = output.FormatRaw
	}
	if format != output.FormatRaw && format != output.FormatJSON {
		showSubcommandHelp(c)
		return fmt.Errorf("invalid format: %s", format)
	}

	var logger Logger
	logger = NewDefaultRawLogger()
	if format == output.FormatJSON {
		logger = NewDefaultJSONLoger()
	}

	if minimal && idOnly {
		logger.Print(NewErrorOutput("Either define --minimal or --id-only"))
		os.Exit(1)
	}

	warnings := []string{}
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	// Config validation
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	warnings = append(warnings, warns...)
	if err != nil {
		logger.Print(NewErrorOutput("Either define --minimal or --id-only", warnings...))
		os.Exit(1)
	}

	if len(bitriseConfig.Workflows) > 0 {
		workflowInfoMap := map[string]map[string]string{}
		for workflowID, workflow := range bitriseConfig.Workflows {
			var workflowInfo map[string]string
			if !idOnly {
				workflowInfo = map[string]string{}
				workflowInfo["title"] = workflow.Title
				workflowInfo["summary"] = workflow.Summary
				if !minimal {
					workflowInfo["description"] = workflow.Description
				}
			}

			workflowInfoMap[workflowID] = workflowInfo
		}

		logger.Print(NewOutput(workflowInfoMap, warnings...))
	}

	return nil
}
