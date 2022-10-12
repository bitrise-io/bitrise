package cli

import (
	"errors"
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
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/version"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	coreanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/gofrs/uuid"
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

var workflowNotSpecifiedErr = errors.New("workflow not specified")
var utilityWorkflowSpecifiedErr = errors.New("utility workflow specified")
var workflowRunFailedErr = errors.New("workflow run failed")

type WorkflowRunModes struct {
	CIMode                  bool
	PRMode                  bool
	DebugMode               bool
	SecretFilteringMode     bool
	SecretEnvsFilteringMode bool
	NoOutputTimeout         time.Duration
}

type RunConfig struct {
	Modes    WorkflowRunModes
	Config   models.BitriseDataModel
	Workflow string
	Secrets  []envmanModels.EnvironmentItemModel
}

type StepExecutionPlan struct {
	UUID   string `json:"uuid"`
	StepID string `json:"step_id"`
}

type WorkflowExecutionPlan struct {
	UUID       string              `json:"uuid"`
	WorkflowID string              `json:"workflow_id"`
	Steps      []StepExecutionPlan `json:"steps"`
}

type WorkflowRunDescriptor struct {
	Version          string `json:"version"`
	LogFormatVersion string `json:"log_format_version"`

	CIMode                  bool `json:"ci_mode"`
	PRMode                  bool `json:"pr_mode"`
	DebugMode               bool `json:"debug_mode"`
	NoOutputTimeoutMode     bool `json:"no_output_timeout_mode"`
	SecretFilteringMode     bool `json:"secret_filtering_mode"`
	SecretEnvsFilteringMode bool `json:"secret_envs_filtering_mode"`

	ExecutionPlan []WorkflowExecutionPlan `json:"execution_plan"`
}

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

func run(c *cli.Context) error {
	config, err := processArgs(c)
	if err != nil {
		if err == workflowNotSpecifiedErr {
			printAvailableWorkflows(config.Config)
			failf("No workflow specified")
		} else if err == utilityWorkflowSpecifiedErr {
			printAboutUtilityWorkflowsText()
			failf("Utility workflows can't be triggered directly")
		}
		failf("Failed to process arguments: %s", err)
	}

	tracker := analytics.NewDefaultTracker()
	defer func() {
		tracker.Wait()
	}()

	runner := newWorkflowRunner(tracker)
	exitCode, err := runner.RunWorkflow(*config)
	if err != nil {
		if err == workflowRunFailedErr {
			msg := createWorkflowRunStatusMessage(exitCode)
			printWorkflowRunStatusMessage(msg)
			analytics.LogMessage("info", "bitrise-cli", "exit", map[string]interface{}{"build_slug": os.Getenv("BITRISE_BUILD_SLUG")}, msg)
			os.Exit(exitCode)
		}

		failf(err.Error())
	}

	msg := createWorkflowRunStatusMessage(0)
	printWorkflowRunStatusMessage(msg)
	analytics.LogMessage("info", "bitrise-cli", "exit", map[string]interface{}{"build_slug": os.Getenv("BITRISE_BUILD_SLUG")}, msg)
	os.Exit(0)

	return nil
}

type workflowRunner struct {
	tracker analytics.Tracker
}

func newWorkflowRunner(tracker analytics.Tracker) *workflowRunner {
	return &workflowRunner{tracker: tracker}
}

func (r *workflowRunner) RunWorkflow(config RunConfig) (int, error) {
	PrintBitriseHeaderASCIIArt(version.VERSION)

	if err := registerRunModes(config.Modes); err != nil {
		return 1, fmt.Errorf("failed to register workflow run modes: %s", err)
	}

	printRunningWorkflow(config.Config, config.Workflow)

	return runWorkflows(config.Config, config.Secrets, config.Workflow, r.tracker)
}

func processArgs(c *cli.Context) (*RunConfig, error) {
	workflowToRunID := c.String(WorkflowKey)
	if workflowToRunID == "" && len(c.Args()) > 0 {
		workflowToRunID = c.Args()[0]
	}

	if workflowToRunID == "" {
		return nil, workflowNotSpecifiedErr
	}
	if strings.HasPrefix(workflowToRunID, "_") {
		return nil, utilityWorkflowSpecifiedErr
	}

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
		return nil, fmt.Errorf("failed to parse command params: %s", err)
	}

	inventoryEnvironments, err := CreateInventoryFromCLIParams(runParams.InventoryBase64Data, runParams.InventoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %s", err)
	}

	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(runParams.BitriseConfigBase64Data, runParams.BitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create bitrise config: %s", err)
	}

	isPRMode, err := isPRMode(prGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		return nil, fmt.Errorf("failed to check PR mode: %s", err)
	}

	isCIMode, err := isCIMode(ciGlobalFlagPtr, inventoryEnvironments)
	if err != nil {
		return nil, fmt.Errorf("failed to check CI mode: %s", err)
	}

	enabledFiltering, err := isSecretFiltering(secretFiltering, inventoryEnvironments)
	if err != nil {
		return nil, fmt.Errorf("failed to check Secret Filtering mode: %s", err)
	}

	enabledEnvsFiltering, err := isSecretEnvsFiltering(secretEnvsFiltering, inventoryEnvironments)
	if err != nil {
		return nil, fmt.Errorf("failed to check Secret Envs Filtering mode: %s", err)
	}

	noOutputTimeout := readNoOutputTimoutConfiguration(inventoryEnvironments)

	return &RunConfig{
		Modes: WorkflowRunModes{
			CIMode:                  isCIMode,
			PRMode:                  isPRMode,
			DebugMode:               configs.IsDebugMode,
			NoOutputTimeout:         noOutputTimeout,
			SecretFilteringMode:     enabledFiltering,
			SecretEnvsFilteringMode: enabledEnvsFiltering,
		},
		Config:   bitriseConfig,
		Workflow: workflowToRunID,
		Secrets:  inventoryEnvironments,
	}, nil
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

func printAboutUtilityWorkflowsText() {
	log.Print("Note about utility workflows:")
	log.Print(" Utility workflow names start with '_' (example: _my_utility_workflow).")
	log.Print(" These workflows can't be triggered directly, but can be used by other workflows")
	log.Print(" in the before_run and after_run lists.")
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

func registerRunModes(modes WorkflowRunModes) error {
	if err := registerSecretFiltering(modes.SecretFilteringMode); err != nil {
		return fmt.Errorf("failed to register Secret Filtering mode: %s", err)
	}

	if err := registerSecretEnvsFiltering(modes.SecretEnvsFilteringMode); err != nil {
		return fmt.Errorf("failed to register Secret Envs Filtering mode: %s", err)
	}

	if err := registerPrMode(modes.PRMode); err != nil {
		return fmt.Errorf("failed to register PR mode: %s", err)
	}

	if err := registerCIMode(modes.CIMode); err != nil {
		return fmt.Errorf("failed to register CI mode: %s", err)
	}

	registerNoOutputTimeout(modes.NoOutputTimeout)
	return nil
}

func createWorkflowRunDescriptor(targetWorkflow string, workflows map[string]models.WorkflowModel, uuidProvider func() string) WorkflowRunDescriptor {
	var executionPlan []WorkflowExecutionPlan
	workflowList := walkWorkflows(targetWorkflow, workflows, nil)
	for _, workflowID := range workflowList {
		workflow := workflows[workflowID]

		var stepPlan []StepExecutionPlan
		for _, stepItem := range workflow.Steps {
			var stepID string
			for id := range stepItem {
				stepID = id
			}

			stepPlan = append(stepPlan, StepExecutionPlan{
				UUID:   uuidProvider(),
				StepID: stepID,
			})
		}

		executionPlan = append(executionPlan, WorkflowExecutionPlan{
			UUID:       uuidProvider(),
			WorkflowID: workflowID,
			Steps:      stepPlan,
		})
	}

	return WorkflowRunDescriptor{
		Version:                 version.VERSION,
		LogFormatVersion:        "1",
		CIMode:                  configs.IsCIMode,
		PRMode:                  configs.IsPullRequestMode,
		DebugMode:               configs.IsDebugMode,
		NoOutputTimeoutMode:     configs.NoOutputTimeout > 0,
		SecretFilteringMode:     configs.IsSecretFiltering,
		SecretEnvsFilteringMode: configs.IsSecretEnvsFiltering,
		ExecutionPlan:           executionPlan,
	}
}

func runWorkflows(bitriseConfig models.BitriseDataModel, inventoryEnvironments []envmanModels.EnvironmentItemModel, workflowToRunID string, tracker analytics.Tracker) (int, error) {
	if workflowToRunID == "" {
		return 1, workflowNotSpecifiedErr
	}

	if err := bitrise.RunSetupIfNeeded(version.VERSION, false); err != nil {
		return 1, fmt.Errorf("setup failed: %s", err)
	}

	startTime := time.Now()

	// Run selected configuration
	if buildRunResults, err := runWorkflowWithConfiguration(startTime, workflowToRunID, bitriseConfig, inventoryEnvironments, tracker); err != nil {
		return 1, fmt.Errorf("failed to run workflow: %s", err)
	} else if buildRunResults.IsBuildFailed() {
		return buildRunResults.ExitCode(), workflowRunFailedErr
	}
	if err := checkUpdate(); err != nil {
		log.Warnf("failed to check for update, error: %s", err)
	}

	return 0, nil
}

// RunWorkflowWithConfiguration ...
func runWorkflowWithConfiguration(
	startTime time.Time,
	workflowToRunID string,
	bitriseConfig models.BitriseDataModel,
	secretEnvironments []envmanModels.EnvironmentItemModel,
	tracker analytics.Tracker) (models.BuildRunResultsModel, error) {

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunID]
	if !exist {
		return models.BuildRunResultsModel{}, fmt.Errorf("Specified Workflow (%s) does not exist", workflowToRunID)
	}

	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunID
	}

	// Envman setup
	if err := os.Setenv(configs.EnvstorePathEnvKey, configs.OutputEnvstorePath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := os.Setenv(configs.FormattedOutputPathEnvKey, configs.FormattedOutputPath); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to add env, err: %s", err)
	}

	if err := tools.EnvmanInit(configs.OutputEnvstorePath, false); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to run envman init: %s", err)
	}

	// App level environment
	environments := append(secretEnvironments, bitriseConfig.App.Environments...)

	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_ID", workflowToRunID); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_ID env: %s", err)
	}
	if err := os.Setenv("BITRISE_TRIGGERED_WORKFLOW_TITLE", workflowToRun.Title); err != nil {
		return models.BuildRunResultsModel{}, fmt.Errorf("Failed to set BITRISE_TRIGGERED_WORKFLOW_TITLE env: %s", err)
	}

	environments = append(environments, workflowToRun.Environments...)

	// Bootstrap Toolkits
	for _, aToolkit := range toolkits.AllSupportedToolkits() {
		toolkitName := aToolkit.ToolkitName()
		if !aToolkit.IsToolAvailableInPATH() {
			// don't bootstrap if any preinstalled version is available,
			// the toolkit's `PrepareForStepRun` can bootstrap for itself later if required
			// or if the system installed version is not sufficient
			if err := aToolkit.Bootstrap(); err != nil {
				return models.BuildRunResultsModel{}, fmt.Errorf("Failed to bootstrap the required toolkit for the step (%s), error: %s",
					toolkitName, err)
			}
		}
	}

	// Trigger WillStartRun
	buildRunStartModel := models.BuildRunStartModel{
		EventName:   string(plugins.WillStartRun),
		StartTime:   startTime,
		ProjectType: bitriseConfig.ProjectType,
	}
	if err := plugins.TriggerEvent(plugins.WillStartRun, buildRunStartModel); err != nil {
		log.Warnf("Failed to trigger WillStartRun, error: %s", err)
	}

	//
	buildRunResults := models.BuildRunResultsModel{
		StartTime:      startTime,
		StepmanUpdates: map[string]int{},
		ProjectType:    bitriseConfig.ProjectType,
	}

	descriptor := createWorkflowRunDescriptor(workflowToRunID, bitriseConfig.Workflows, func() string { return uuid.Must(uuid.NewV4()).String() })
	if len(descriptor.ExecutionPlan) < 1 {
		return models.BuildRunResultsModel{}, fmt.Errorf("execution plan doesn't have any workflow to run")
	}
	lastWorkflowID := descriptor.ExecutionPlan[len(descriptor.ExecutionPlan)-1].WorkflowID

	buildIDProperties := coreanalytics.Properties{analytics.BuildExecutionID: uuid.Must(uuid.NewV4()).String()}
	buildRunResults, err := activateAndRunWorkflow(
		descriptor,
		workflowToRunID, workflowToRun, bitriseConfig,
		buildRunResults,
		&environments, secretEnvironments,
		lastWorkflowID, tracker, buildIDProperties)
	if err != nil {
		return buildRunResults, errors.New("[BITRISE_CLI] - Failed to activate and run workflow " + workflowToRunID)
	}

	// Build finished
	bitrise.PrintSummary(buildRunResults)

	// Trigger WorkflowRunDidFinish
	buildRunResults.EventName = string(plugins.DidFinishRun)
	if err := plugins.TriggerEvent(plugins.DidFinishRun, buildRunResults); err != nil {
		log.Warnf("Failed to trigger WorkflowRunDidFinish, error: %s", err)
	}

	return buildRunResults, nil
}

func activateAndRunWorkflow(
	descriptor WorkflowRunDescriptor,
	workflowID string, workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel,
	buildRunResults models.BuildRunResultsModel,
	environments *[]envmanModels.EnvironmentItemModel, secrets []envmanModels.EnvironmentItemModel,
	lastWorkflowID string, tracker analytics.Tracker, buildIDProperties coreanalytics.Properties) (models.BuildRunResultsModel, error) {

	for _, workflowRunPlan := range descriptor.ExecutionPlan {
		buildRunResults = runWorkflow(workflowRunPlan, workflowRunPlan.WorkflowID, workflow, bitriseConfig.DefaultStepLibSource, buildRunResults, environments, secrets, workflowID == lastWorkflowID, tracker, buildIDProperties)
	}

	return buildRunResults, nil
}

func createWorkflowRunStatusMessage(exitCode int) string {
	var message string
	var colorMessage string
	if exitCode == 0 {
		message = "Bitrise build successful"
		colorMessage = colorstring.Green(message)
	} else {
		message = fmt.Sprintf("Bitrise build failed (exit code: %d)", exitCode)
		colorMessage = colorstring.Red(message)
	}
	return colorMessage
}

func printWorkflowRunStatusMessage(msg string) {
	log.Print()
	log.Print(msg)
	log.Print()
}

func walkWorkflows(workflowID string, workflows map[string]models.WorkflowModel, workflowStack []string) []string {
	workflow := workflows[workflowID]
	for _, before := range workflow.BeforeRun {
		workflowStack = walkWorkflows(before, workflows, workflowStack)
	}

	workflowStack = append(workflowStack, workflowID)

	for _, after := range workflow.AfterRun {
		workflowStack = walkWorkflows(after, workflows, workflowStack)
	}

	return workflowStack
}
