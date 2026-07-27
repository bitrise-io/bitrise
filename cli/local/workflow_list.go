package local

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/spf13/cobra"
)

// NewWorkflowListCommand ...
func NewWorkflowListCommand() *cobra.Command {
	workflowListCommand := &cobra.Command{
		Use:   "workflows",
		Short: "List of available workflows in config.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := workflowList(cmd); err != nil {
				log.Errorf("List of available workflows in config failed, error: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}

	workflowListCommand.Flags().StringP(cmdutil.ConfigKey, "c", "", "Path where the workflow config file is located.")
	workflowListCommand.Flags().String(cmdutil.ConfigBase64Key, "", "base64 encoded config data.")
	workflowListCommand.Flags().String(cmdutil.FormatKey, "", "Output format. Accepted: raw, json.")
	workflowListCommand.Flags().Bool("minimal", false, "Print summary of workflows only.")
	workflowListCommand.Flags().Bool("id-only", false, "Print workflow ids only.")

	return workflowListCommand
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

		if !idOnly && len(info) == 0 {
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

		if !idOnly && len(info) == 0 {
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

func workflowList(cmd *cobra.Command) error {
	bitriseConfigBase64Data, _ := cmd.Flags().GetString(cmdutil.ConfigBase64Key)
	bitriseConfigPath, _ := cmd.Flags().GetString(cmdutil.ConfigKey)

	format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
	minimal, _ := cmd.Flags().GetBool("minimal")
	idOnly, _ := cmd.Flags().GetBool("id-only")

	// Input validation
	if format == "" {
		format = output.FormatRaw
	}
	if format != output.FormatRaw && format != output.FormatJSON {
		cmdutil.ShowSubcommandHelp(cmd)
		return fmt.Errorf("invalid format: %s", format)
	}

	var logger cmdutil.Logger = cmdutil.NewDefaultRawLogger()
	if format == output.FormatJSON {
		logger = cmdutil.NewDefaultJSONLogger()
	}

	if minimal && idOnly {
		logger.Print(NewErrorOutput("Either define --minimal or --id-only"))
		os.Exit(1)
	}

	warnings := []string{}

	// Config validation
	bitriseConfig, warns, err := cmdutil.CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath, bitrise.ValidationTypeFull)
	warnings = append(warnings, warns...)
	if err != nil {
		logger.Print(NewErrorOutput(fmt.Sprintf("Failed to create bitrise config: %s", err)))
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
