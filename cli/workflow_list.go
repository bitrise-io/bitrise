package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

func printWorkflList(workflowList map[string]map[string]string, format string, minimal bool) error {
	printRawWorkflowMap := func(name string, workflow map[string]string) {
		fmt.Printf("⚡️ %s\n", colorstring.Green(name))

		if workflow["summary"] == "" {
			workflow["summary"] = colorstring.Yellow("No summary available.")
		}
		fmt.Printf("  Summary: %s\n", workflow["summary"])

		if !minimal {
			if workflow["description"] == "" {
				workflow["description"] = colorstring.Yellow("No description available.")
			}
			fmt.Printf("  Description: %s\n", workflow["description"])
		}

		fmt.Println()
	}

	switch format {
	case configs.OutputFormatRaw:
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

	case configs.OutputFormatJSON:
		bytes, err := json.Marshal(workflowList)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	default:
		return fmt.Errorf("Invalid output format: %s", format)
	}
	return nil
}

func workflowList(c *cli.Context) {
	// Input validation
	format := c.String(OuputFormatKey)
	if format == "" {
		format = configs.OutputFormatRaw
	} else if !(format == configs.OutputFormatRaw || format == configs.OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), configs.OutputFormatJSON)
	}

	minimal := c.Bool(MinimalModeKey)

	// Config validation
	bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
	if err != nil {
		log.Fatalf("Failed to create bitrise cofing, err: %s", err)
	}

	workflowList := map[string]map[string]string{}
	if len(bitriseConfig.Workflows) > 0 {
		for workflowID, workflow := range bitriseConfig.Workflows {
			workflowMap := map[string]string{}
			workflowMap["summary"] = workflow.Summary
			if !minimal {
				workflowMap["description"] = workflow.Description
			}

			workflowList[workflowID] = workflowMap
		}
	}

	if err := printWorkflList(workflowList, format, minimal); err != nil {
		log.Fatalf("Failed to print workflows, err: %s", err)
	}
}
