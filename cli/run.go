package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/urfave/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"
	OutputFormatKey        = "output-format"

	depManagerBrew      = "brew"
	depManagerTryCheck  = "_"
	secretFilteringFlag = "secret-filtering"
)

var runCommand = cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Runs a specified Workflow.",
	Action:  run,
	Flags: []cli.Flag{
		// cli params
		cli.StringFlag{Name: WorkflowKey, Usage: "workflow id to run."},
		cli.StringFlag{Name: ConfigKey + ", " + configShortKey, Usage: "Path where the workflow config file is located."},
		cli.StringFlag{Name: InventoryKey + ", " + inventoryShortKey, Usage: "Path of the inventory file."},
		cli.BoolFlag{Name: secretFilteringFlag, Usage: "Hide secret values from the log."},

		// cli params used in CI mode
		cli.StringFlag{Name: JSONParamsKey, Usage: "Specify command flags with json string-string hash."},
		cli.StringFlag{Name: JSONParamsBase64Key, Usage: "Specify command flags with base64 encoded json string-string hash."},
		cli.StringFlag{Name: OutputFormatKey, Usage: "Log format. Available values: json, console"},

		// deprecated
		flPath,

		// should deprecate
		cli.StringFlag{Name: ConfigBase64Key, Usage: "base64 encoded config data."},
		cli.StringFlag{Name: InventoryBase64Key, Usage: "base64 encoded inventory data."},
	},
}

func printAboutUtilityWorkflowsText() {
	log.Print("Note about utility workflows:")
	log.Print(" Utility workflow names start with '_' (example: _my_utility_workflow).")
	log.Print(" These workflows can't be triggered directly, but can be used by other workflows")
	log.Print(" in the before_run and after_run lists.")
}

func printAvailableWorkflows(config models.BitriseDataModel) {
	workflowNames := []string{}
	utilityWorkflowNames := []string{}

	for wfName := range config.Workflows {
		if strings.HasPrefix(wfName, "_") {
			utilityWorkflowNames = append(utilityWorkflowNames, wfName)
		} else {
			workflowNames = append(workflowNames, wfName)
		}
	}
	sort.Strings(workflowNames)
	sort.Strings(utilityWorkflowNames)

	if len(workflowNames) > 0 {
		log.Print("The following workflows are available:")
		for _, wfName := range workflowNames {
			log.Print(" * " + wfName)
		}

		log.Print()
		log.Print("You can run a selected workflow with:")
		log.Print("$ bitrise run WORKFLOW-ID")
		log.Print()
	} else {
		log.Print("No workflows are available!")
	}

	if len(utilityWorkflowNames) > 0 {
		log.Print()
		log.Print("The following utility workflows are defined:")
		for _, wfName := range utilityWorkflowNames {
			log.Print(" * " + wfName)
		}

		log.Print()
		printAboutUtilityWorkflowsText()
		log.Print()
	}
}

func runAndExit(bitriseConfig models.BitriseDataModel, inventoryEnvironments []envmanModels.EnvironmentItemModel, workflowToRunID string, tracker analytics.Tracker) {
	if workflowToRunID == "" {
		failf("No workflow id specified")
	}

	if err := bitrise.RunSetupIfNeeded(version.VERSION, false); err != nil {
		failf("Setup failed, error: %s", err)
	}

	startTime := time.Now()

	// Run selected configuration
	if buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments, tracker); err != nil {
		tracker.Wait()
		logExit(1)
		failf("Failed to run workflow, error: %s", err)
	} else if buildRunResults.IsBuildFailed() {
		tracker.Wait()
		exitCode := buildRunResults.ExitCode()
		logExit(exitCode)
		os.Exit(exitCode)
	}
	if err := checkUpdate(); err != nil {
		log.Warnf("failed to check for update, error: %s", err)
	}

	tracker.Wait()
	logExit(0)
	os.Exit(0)
}

func logExit(exitCode int) {
	var message string
	var colorMessage string
	if exitCode == 0 {
		message = "Bitrise build successful"
		colorMessage = colorstring.Green(message)
	} else {
		message = fmt.Sprintf("Bitrise build failed (exit code: %d)", exitCode)
		colorMessage = colorstring.Red(message)
	}
	analytics.LogMessage("info", "bitrise-cli", "exit", map[string]interface{}{"build_slug": os.Getenv("BITRISE_BUILD_SLUG")}, message)
	log.Print()
	log.Print(colorMessage)
	log.Print()
}

func printRunningWorkflow(bitriseConfig models.BitriseDataModel, targetWorkflowToRunID string) {
	beforeWorkflowIDs := bitriseConfig.Workflows[targetWorkflowToRunID].BeforeRun
	afterWorkflowIDs := bitriseConfig.Workflows[targetWorkflowToRunID].AfterRun
	workflowsString := ""
	if len(beforeWorkflowIDs) == 0 && len(afterWorkflowIDs) == 0 {
		workflowsString = "Running workflow: "
	} else {
		workflowsString = "Running workflows: "
	}

	if len(beforeWorkflowIDs) != 0 {
		for _, workflowName := range beforeWorkflowIDs {
			workflowsString = workflowsString + workflowName + " --> "
		}
	}

	workflowsString = workflowsString + colorstring.Green(targetWorkflowToRunID)

	if len(afterWorkflowIDs) != 0 {
		for _, workflowName := range afterWorkflowIDs {
			workflowsString = workflowsString + " --> " + workflowName
		}
	}

	log.Infof(workflowsString)
}

func run(c *cli.Context) error {
	tracker := analytics.NewDefaultTracker()
	PrintBitriseHeaderASCIIArt(version.VERSION)

	//
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

	workflowToRunID := c.String(WorkflowKey)
	if workflowToRunID == "" && len(c.Args()) > 0 {
		workflowToRunID = c.Args()[0]
	}

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		log.Warn("'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	jsonParams := c.String(JSONParamsKey)
	jsonParamsBase64 := c.String(JSONParamsBase64Key)

	runParams, err := parseRunParams(
		workflowToRunID,
		bitriseConfigPath, bitriseConfigBase64Data,
		inventoryPath, inventoryBase64Data,
		jsonParams, jsonParamsBase64)
	if err != nil {
		return fmt.Errorf("Failed to parse command params, error: %s", err)
	}
	//

	// Inventory validation
	inventoryEnvironments, err := CreateInventoryFromCLIParams(runParams.InventoryBase64Data, runParams.InventoryPath)
	if err != nil {
		failf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(runParams.BitriseConfigBase64Data, runParams.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		failf("Failed to create bitrise config, error: %s", err)
	}

	// Workflow id validation
	if runParams.WorkflowToRunID == "" {
		// no workflow specified
		//  list all the available ones and then exit
		log.Error("No workflow specified!")
		log.Print()
		printAvailableWorkflows(bitriseConfig)
		os.Exit(1)
	}
	if strings.HasPrefix(runParams.WorkflowToRunID, "_") {
		// util workflow specified
		//  print about util workflows and then exit
		log.Error("Utility workflows can't be triggered directly")
		log.Print()
		printAboutUtilityWorkflowsText()
		os.Exit(1)
	}
	//

	//
	// Main
	enabledFiltering, err := isSecretFiltering(secretFiltering, inventoryEnvironments)
	if err != nil {
		failf("Failed to check Secret Filtering mode, error: %s", err)
	}

	if err := registerSecretFiltering(enabledFiltering); err != nil {
		failf("Failed to register Secret Filtering mode, error: %s", err)
	}

	enabledEnvsFiltering, err := isSecretEnvsFiltering(secretEnvsFiltering, inventoryEnvironments)
	if err != nil {
		failf("Failed to check Secret Envs Filtering mode, error: %s", err)
	}

	if err := registerSecretEnvsFiltering(enabledEnvsFiltering); err != nil {
		failf("Failed to register Secret Envs Filtering mode, error: %s", err)
	}

	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		failf("Failed to check PR mode, error: %s", err)
	}

	if err := registerPrMode(isPRMode); err != nil {
		failf("Failed to register PR mode, error: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		failf("Failed to check CI mode, error: %s", err)
	}

	if err := registerCIMode(isCIMode); err != nil {
		failf("Failed to register CI mode, error: %s", err)
	}

	noOutputTimeout := readNoOutputTimoutConfiguration(inventoryEnvironments)
	registerNoOutputTimeout(noOutputTimeout)

	printRunningWorkflow(bitriseConfig, runParams.WorkflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, runParams.WorkflowToRunID, tracker)
	//
	return nil
}
