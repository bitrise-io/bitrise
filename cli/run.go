package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	utilsLog "github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pointers"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

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

		// deprecated
		flPath,

		// should deprecate
		cli.StringFlag{Name: ConfigBase64Key, Usage: "base64 encoded config data."},
		cli.StringFlag{Name: InventoryBase64Key, Usage: "base64 encoded inventory data."},
	},
}

func printAboutUtilityWorkflowsText() {
	fmt.Println("Note about utility workflows:")
	fmt.Println(" Utility workflow names start with '_' (example: _my_utility_workflow).")
	fmt.Println(" These workflows can't be triggered directly, but can be used by other workflows")
	fmt.Println(" in the before_run and after_run lists.")
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
		fmt.Println("The following workflows are available:")
		for _, wfName := range workflowNames {
			fmt.Println(" * " + wfName)
		}

		fmt.Println()
		fmt.Println("You can run a selected workflow with:")
		fmt.Println("$ bitrise run WORKFLOW-ID")
		fmt.Println()
	} else {
		fmt.Println("No workflows are available!")
	}

	if len(utilityWorkflowNames) > 0 {
		fmt.Println()
		fmt.Println("The following utility workflows are defined:")
		for _, wfName := range utilityWorkflowNames {
			fmt.Println(" * " + wfName)
		}

		fmt.Println()
		printAboutUtilityWorkflowsText()
		fmt.Println()
	}
}

func runAndExit(bitriseConfig models.BitriseDataModel, inventoryEnvironments []envmanModels.EnvironmentItemModel, workflowToRunID string) {
	if workflowToRunID == "" {
		log.Fatal("No workflow id specified")
	}

	if err := bitrise.RunSetupIfNeeded(version.VERSION, false); err != nil {
		log.Fatalf("Setup failed, error: %s", err)
	}

	startTime := time.Now()

	// Run selected configuration
	if buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments); err != nil {
		logExit(1)
		log.Fatalf("Failed to run workflow, error: %s", err)
	} else if buildRunResults.IsBuildFailed() {
		logExit(1)
		os.Exit(1)
	}
	if err := checkUpdate(); err != nil {
		log.Warnf("failed to check for update, error: %s", err)
	}
	logExit(0)
	os.Exit(0)
}

func logExit(exitCode int) {
	var message string
	var colorMessage string
	if exitCode == 0 {
		message = "Bitrise Build Successful"
		colorMessage = colorstring.Green(message)
	} else {
		message = fmt.Sprintf("Bitrise Build Failed (exit code: %d)", exitCode)
		colorMessage = colorstring.Red(message)
	}
	utilsLog.RInfof("bitrise-cli", "exit", map[string]interface{}{"build_slug": os.Getenv("BITRISE_BUILD_SLUG")}, message)
	fmt.Println()
	fmt.Print(colorMessage)
	fmt.Println()
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
		log.Fatalf("Failed to create inventory, error: %s", err)
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(runParams.BitriseConfigBase64Data, runParams.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Workflow id validation
	if runParams.WorkflowToRunID == "" {
		// no workflow specified
		//  list all the available ones and then exit
		log.Error("No workflow specified!")
		fmt.Println()
		printAvailableWorkflows(bitriseConfig)
		os.Exit(1)
	}
	if strings.HasPrefix(runParams.WorkflowToRunID, "_") {
		// util workflow specified
		//  print about util workflows and then exit
		log.Error("Utility workflows can't be triggered directly")
		fmt.Println()
		printAboutUtilityWorkflowsText()
		os.Exit(1)
	}
	//

	//
	// Main
	enabledFiltering, err := isSecretFiltering(secretFiltering, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check Secret Filtering mode, error: %s", err)
	}

	if err := registerSecretFiltering(enabledFiltering); err != nil {
		log.Fatalf("Failed to register Secret Filtering mode, error: %s", err)
	}

	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check PR mode, error: %s", err)
	}

	if err := registerPrMode(isPRMode); err != nil {
		log.Fatalf("Failed to register PR mode, error: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		log.Fatalf("Failed to check CI mode, error: %s", err)
	}

	if err := registerCIMode(isCIMode); err != nil {
		log.Fatalf("Failed to register CI mode, error: %s", err)
	}

	printRunningWorkflow(bitriseConfig, runParams.WorkflowToRunID)

	runAndExit(bitriseConfig, inventoryEnvironments, runParams.WorkflowToRunID)
	//
	return nil
}
