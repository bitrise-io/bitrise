package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
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
			Usage: "Output format. Accepted: raw, json, yml.",
		},
		cli.BoolFlag{
			Name:  "minimal",
			Usage: "Print only workflow summary.",
		},
	},
}

func printWorkflList(workflowList map[string]map[string]string, format string, minimal bool) error {
	printRawWorkflowMap := func(name string, workflow map[string]string) {
		fmt.Printf("⚡️ %s\n", colorstring.Green(name))
		fmt.Printf("  %s: %s\n", colorstring.Yellow("Title"), workflow["title"])
		fmt.Printf("  %s: %s\n", colorstring.Yellow("Summary"), workflow["summary"])
		if !minimal {
			fmt.Printf("  %s: %s\n", colorstring.Yellow("Description"), workflow["description"])
		}
		fmt.Printf("  %s: bitrise run %s\n", colorstring.Yellow("Run with"), name)
		fmt.Println()
	}

	switch format {
	case output.FormatRaw:
		workflowNames := []string{}
		utilityWorkflowNames := []string{}

		for wfName := range workflowList {
			if strings.HasPrefix(wfName, "_") {
				utilityWorkflowNames = append(utilityWorkflowNames, wfName)
			} else {
				workflowNames = append(workflowNames, wfName)
			}
		}
		sort.Strings(workflowNames)
		sort.Strings(utilityWorkflowNames)

		fmt.Println()

		if len(workflowNames) > 0 {
			fmt.Printf("%s\n", "Workflows")
			fmt.Printf("%s\n", "---------")
			for _, name := range workflowNames {
				workflow := workflowList[name]
				printRawWorkflowMap(name, workflow)
			}
			fmt.Println()
		}

		if len(utilityWorkflowNames) > 0 {
			fmt.Printf("%s\n", "Util Workflows")
			fmt.Printf("%s\n", "--------------")
			for _, name := range utilityWorkflowNames {
				workflow := workflowList[name]
				printRawWorkflowMap(name, workflow)
			}
			fmt.Println()
		}

		if len(workflowNames) == 0 && len(utilityWorkflowNames) == 0 {
			fmt.Printf("Config doesn't contain any workflow")
		}

	case output.FormatJSON:
		bytes, err := json.Marshal(workflowList)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))

	case output.FormatYML:
		bytes, err := yaml.Marshal(workflowList)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	default:
		return fmt.Errorf("Invalid output format: %s", format)
	}
	return nil

}

func workflowList(c *cli.Context) error {
	warnings := []string{}

	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	format := c.String(OuputFormatKey)

	minimal := c.Bool(MinimalModeKey)
	workflowIDOnly := c.Bool(WorkFlowIDOnlyKey)

	// Input validation
	if format == "" {
		format = output.FormatRaw
	} else if !(format == output.FormatRaw || format == output.FormatJSON || format == output.FormatYML) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	// Config validation
	bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	warnings = append(warnings, warns...)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to create bitrise config, err: %s", err), warnings, output.FormatJSON)
	}

	if len(bitriseConfig.Workflows) > 0 {
		if workflowIDOnly {
			workflowIDs := []string{}
			for workflowID := range bitriseConfig.Workflows {
				workflowIDs = append(workflowIDs, workflowID)
			}
			fmt.Println(strings.Join(workflowIDs, " "))
			return nil
		}

		workflowList := map[string]map[string]string{}
		for workflowID, workflow := range bitriseConfig.Workflows {
			workflowMap := map[string]string{}
			workflowMap["title"] = workflow.Title
			workflowMap["summary"] = workflow.Summary
			if !minimal {
				workflowMap["description"] = workflow.Description
			}
			workflowList[workflowID] = workflowMap
		}

		if err := printWorkflList(workflowList, format, minimal); err != nil {
			registerFatal(fmt.Sprintf("Failed to print workflows, err: %s", err), warnings, output.FormatJSON)
		}
	}

	return nil
}
