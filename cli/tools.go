package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/bitrise-io/colorstring"
	"github.com/urfave/cli"
)

const (
	outputFormatPlaintext = "plaintext"
	outputFormatJSON      = "json"
	outputFormatBash      = "bash"

	toolsSetupSubcommandName   = "setup"
	toolsInstallSubcommandName = "install"
	toolsLatestSubcommandName  = "latest"
	toolsInfoCommandName       = "info"

	toolsConfigKey      = "config"
	toolsConfigShortKey = "c"

	toolsWorkflowKey = "workflow"

	toolsOutputFormatKey      = "format"
	toolsOutputFormatShortKey = "f"

	toolsInstalledKey      = "installed"
	toolsInstalledShortKey = "i"

	toolsProviderKey      = "provider"
	toolsProviderShortKey = "p"

	toolsActiveKey      = "active"
	toolsActiveShortKey = "a"
)

var (
	flToolsProvider = cli.StringFlag{
		Name:  toolsProviderKey + ", " + toolsProviderShortKey,
		Usage: `Tool provider to use (asdf/mise). If not specified, uses the default.`,
	}

	flToolsOutputFormat = cli.StringFlag{
		Name:  toolsOutputFormatKey + ", " + toolsOutputFormatShortKey,
		Usage: `Output format of the env vars that activate installed tools. Options: plaintext, json, bash`,
		Value: outputFormatPlaintext,
	}

	flToolsInstalled = cli.BoolFlag{
		Name:  toolsInstalledKey + ", " + toolsInstalledShortKey,
		Usage: `Install the latest already installed version instead of the latest available release`,
	}

	flToolsActive = cli.BoolFlag{
		Name:  toolsActiveKey + ", " + toolsActiveShortKey,
		Usage: `Show only currently active tools in the shell context (based on config files in current directory)`,
	}

	flToolsConfig = cli.StringSliceFlag{
		Name: toolsConfigKey + ", " + toolsConfigShortKey,
		Usage: `Config or version file paths to install tools from. Can be specified multiple times. If not provided, detects files in the working directory. Supported file names and formats:
	- .tool-versions (asdf/mise style): multiple tools, one "<tool> <version>" per line
	- .<tool>-version (e.g. .node-version, .ruby-version): single tool, version string only
	- bitrise.yml: tools defined in the "tools" section`,
		TakesFile: true,
	}

	flToolsWorkflow = cli.StringFlag{
		Name:  toolsWorkflowKey + ", w",
		Usage: "Workflow ID to use when installing from bitrise.yml (optional, uses global tools if not specified)",
	}
)

var toolsInfoSubcommand = cli.Command{
	Name:      toolsInfoCommandName,
	Usage:     "Show information about installed or active tools.",
	UsageText: "bitrise tools info [--active] [--format FORMAT]",
	Description: `Display information about development tools managed by the tool provider.

By default, shows all installed tool versions. Use --active to show only the tools
that are currently active in the shell context (based on your bitrise.yml, .tool-versions, mise.toml,
or other config files in the current directory).

EXAMPLES:
   Show all installed tools:
   bitrise tools info

   Show currently active tools:
   bitrise tools info --active

   Output as JSON:
   bitrise tools info --active --format json`,
	Action: func(c *cli.Context) error {
		logCommandParameters(c)
		if err := toolsInfo(c); err != nil {
			log.Errorf("Failed to get tool info: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		flToolsActive,
		flToolsOutputFormat,
	},
}

var (
	toolInstallSubcommandUsageText = "bitrise tools install [--provider PROVIDER] [--format FORMAT] <TOOL@VERSION>"
	toolsInstallSubcommand         = cli.Command{
		Name:      toolsInstallSubcommandName,
		Usage:     "Install a specific tool version",
		UsageText: toolInstallSubcommandUsageText,
		Description: `Install a specific version of a tool using the configured tool provider.

The tool specification uses the format TOOL@VERSION where:
  - TOOL is a valid tool identifier (e.g., node, ruby, python, go, java)
  - VERSION is an exact version number (e.g., 20.10.0, 3.12.1)

EXAMPLES:
   Install Node.js 20.10.0:
   bitrise tools install node@20.10.0

   Install Python 3.12.1:
   bitrise tools install python@3.12.1

   Install and activate in current shell session:
   eval "$(bitrise tools install ruby@3.2.0 --format bash)"

   Use specific provider:
   bitrise tools install go@1.21.5 --provider mise`,
		Action: func(c *cli.Context) error {
			logCommandParameters(c)
			if err := toolsInstall(c); err != nil {
				log.Errorf("Tool install failed: %s", err)
				os.Exit(1)
			}
			return nil
		},
		Flags: []cli.Flag{
			flToolsProvider,
			flToolsOutputFormat,
		},
	}
)

var (
	toolsLatestSubcommandUsageText = "bitrise tools latest [--installed] [--provider PROVIDER] [--format FORMAT] <TOOL[@VERSION]>"
	toolsLatestSubcommand          = cli.Command{
		Name:      toolsLatestSubcommandName,
		Usage:     "Query the latest version of a tool",
		UsageText: toolsLatestSubcommandUsageText,
		Description: `Query the latest version of a tool, optionally matching a version prefix.

The tool specification uses the format TOOL[@VERSION] where:
  - TOOL is a valid tool identifier (e.g., node, ruby, python, go, java)
  - VERSION is an optional version prefix (e.g., 20, 3.12)

By default, queries the latest available release. Use --installed to get the latest of the already installed versions.

EXAMPLES:
   Query latest available Node.js 20.x:
   bitrise tools latest node@20

   Query latest already installed Python 3.12.x:
   bitrise tools latest --installed python@3.12

   Query latest available Node.js (any version):
   bitrise tools latest node

   Output as JSON:
   bitrise tools latest --format json ruby@3`,
		Action: func(c *cli.Context) error {
			logCommandParameters(c)
			if err := toolsLatest(c); err != nil {
				log.Errorf("Tool latest failed: %s", err)
				os.Exit(1)
			}
			return nil
		},
		Flags: []cli.Flag{
			flToolsInstalled,
			flToolsOutputFormat,
			flToolsProvider,
		},
	}
)

var toolsSetupSubcommand = cli.Command{
	Name:      toolsSetupSubcommandName,
	Usage:     "Install tools from version files or bitrise.yml",
	UsageText: "bitrise tools setup [--config FILE]...",
	Description: `Install tools from version files (e.g. .tool-versions, .node-version, .python-version) or from the bitrise.yml.

This is meant to be called from scripts/steps running inside a workflow.

EXAMPLES:
   Setup from .tool-versions:
   bitrise tools setup --config .tool-versions

   Setup from bitrise.yml:
   bitrise tools setup --config bitrise.yml

   Setup and activate in current shell session:
   eval "$(bitrise tools setup --config .tool-versions --format bash)"`,
	Action: func(c *cli.Context) error {
		logCommandParameters(c)
		if err := toolsSetup(c); err != nil {
			log.Errorf("Tool setup failed: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		flToolsConfig,
		flToolsOutputFormat,
		flToolsWorkflow,
	},
}

var toolsCommand = cli.Command{
	Name:  "tools",
	Usage: "Manage available tools from inside the workflow.",
	Subcommands: []cli.Command{
		toolsInfoSubcommand,
		toolsSetupSubcommand,
		toolsInstallSubcommand,
		toolsLatestSubcommand,
	},
}

func toolsSetup(c *cli.Context) error {
	configFiles := c.StringSlice(toolsConfigKey)
	workflowID := c.String(toolsWorkflowKey)
	format := c.String(toolsOutputFormatKey)
	silent := false

	switch format {
	case outputFormatJSON, outputFormatBash:
		// valid formats
		silent = true
	case outputFormatPlaintext:
		// valid format
	default:
		return fmt.Errorf("invalid --format: %s", format)
	}

	var bitriseConfigPath string
	var versionFilePaths []string
	for _, file := range configFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if !silent {
				log.Warnf("file does not exist: %s", file)
			}
			continue
		}

		if isBitriseConfig(file) {
			if bitriseConfigPath != "" {
				return fmt.Errorf("multiple bitrise config files specified: %s and %s (only one bitrise.yml can be used)", bitriseConfigPath, file)
			}

			bitriseConfigPath = file
			continue
		}

		versionFilePaths = append(versionFilePaths, file)
	}

	if bitriseConfigPath != "" {
		config, warnings, err := CreateBitriseConfigFromCLIParams("", bitriseConfigPath, bitrise.ValidationTypeFull)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		for _, warning := range warnings {
			log.Warnf("Config warning: %s", warning)
		}

		tracker := analytics.NewDefaultTracker()
		envs, err := toolprovider.RunDeclarativeSetup(config, tracker, false, workflowID, silent)
		if err != nil {
			return err
		}

		exposedWithEnvman := exposeEnvsWithEnvman(envs, silent)

		output, err := convertToOutputFormat(envs, format, exposedWithEnvman)
		if err != nil {
			return fmt.Errorf("convert to output format: %w", err)
		}
		fmt.Println(output)
	}

	// Setting up from all the other version files.
	tracker := analytics.NewDefaultTracker()
	envs, err := toolprovider.RunVersionFileSetup(versionFilePaths, tracker, silent)
	if err != nil {
		return err
	}

	exposedWithEnvman := exposeEnvsWithEnvman(envs, silent)

	output, err := convertToOutputFormat(envs, format, exposedWithEnvman)
	if err != nil {
		return fmt.Errorf("convert to output format: %w", err)
	}
	fmt.Println(output)
	return nil
}

func isBitriseConfig(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml")
}

func convertToOutputFormat(envs []provider.EnvironmentActivation, format string, exposedWithEnvman bool) (string, error) {
	envMap := toolprovider.ConvertToEnvMap(envs)

	var builder strings.Builder
	switch format {
	case outputFormatPlaintext:
		if len(envs) == 0 {
			return "No new tools were installed.", nil
		}
		if exposedWithEnvman {
			builder.WriteString(colorstring.Green("✓ Tools activated for subsequent steps in the workflow"))
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf(
			"%s %s %s\n",
			colorstring.Yellow("! If you need tools in the current shell session, run"),
			colorstring.Cyan("eval \"$(bitrise tools setup --format bash ...)\""),
			colorstring.Yellow("instead."),
		))
		return builder.String(), nil
	case outputFormatJSON:
		data, err := json.MarshalIndent(envMap, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal JSON: %w", err)
		}
		return string(data), nil
	case outputFormatBash:
		if len(envs) == 0 {
			return "# No new tools were installed.", nil
		}
		// Sort K=V pairs for deterministic output (mostly for our own tests, but also generally useful).
		sortedKeys := make([]string, 0, len(envMap))
		for k := range envMap {
			sortedKeys = append(sortedKeys, k)
		}
		slices.Sort(sortedKeys)
		for _, k := range sortedKeys {
			v := envMap[k]
			builder.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, v))
		}
		message := fmt.Sprintf(
			"# %s\n# Make sure to run %s instead\n",
			colorstring.Yellow("NOTE: Tools have been installed, but they need to be activated for the current shell session."),
			colorstring.Cyan("eval \"$(bitrise tools setup --format bash ...)\""),
		)
		builder.WriteString(message)
		return builder.String(), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// exposeEnvsWithEnvman calls envman to expose the given env vars for subsequent steps in the workflow.
// Returns true if successful (since envman is not always available, e.g. in local runs).
func exposeEnvsWithEnvman(activations []provider.EnvironmentActivation, silent bool) bool {
	// When running inside a workflow step, ENVMAN_ENVSTORE_PATH will be set to OutputEnvstorePath.
	// When running locally/standalone, we fall back to InputEnvstorePath.
	envstorePath := os.Getenv(configs.EnvstorePathEnvKey)
	if envstorePath == "" {
		envstorePath = configs.InputEnvstorePath
	}

	// Check if envstore exists - it should be initialized by the workflow runner
	if _, err := os.Stat(envstorePath); err != nil {
		if !silent {
			if os.IsNotExist(err) {
				log.Warnf("! Envstore not found at %s - envman is not available to store installation paths", envstorePath)
			} else {
				log.Warnf("! Failed to access envstore at %s: %s", envstorePath, err)
			}
		}
		return false
	}

	envs := toolprovider.ConvertToEnvmanEnvs(activations)
	err := tools.EnvmanAddEnvs(envstorePath, envs)
	if err != nil {
		if !silent {
			log.Warnf("! Failed to expose tool envs with envman: %s", err)
		}
		return false
	}
	return true
}

func toolsInfo(c *cli.Context) error {
	format := c.String("format")
	activeOnly := c.Bool("active")

	tools, err := toolprovider.ListInstalledTools("mise", activeOnly, false)
	if err != nil {
		return err
	}

	if format == outputFormatJSON {
		data, err := json.MarshalIndent(tools, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(tools) == 0 {
		if activeOnly {
			log.Infof("No active tools in current context")
		} else {
			log.Infof("No tools installed")
		}
		return nil
	}

	printToolsInfo(tools, activeOnly)
	return nil
}

func printToolsInfo(tools []toolprovider.InstalledTool, activeOnly bool) {
	// Colun width calculation.
	maxNameLen := len("Tool")
	maxVersionLen := len("Version")
	for _, tool := range tools {
		if len(tool.Name) > maxNameLen {
			maxNameLen = len(tool.Name)
		}
		version := tool.ActiveVersion
		if version == "" && len(tool.InstalledVersions) > 0 {
			version = tool.InstalledVersions[0]
		}
		if len(version) > maxVersionLen {
			maxVersionLen = len(version)
		}
	}

	namePad := strings.Repeat(" ", maxNameLen+2)
	versionPad := strings.Repeat(" ", maxVersionLen+2)

	if activeOnly {
		log.Infof("Active tools:")
	} else {
		log.Infof("Installed tools:")
	}
	log.Printf("")

	// Header.
	toolHeader := colorstring.Blue("Tool")
	versionHeader := colorstring.Blue("Version")
	sourceHeader := colorstring.Blue("Source")
	log.Printf("  %s%s%s%s%s", toolHeader, namePad[:maxNameLen-len("Tool")+2], versionHeader, versionPad[:maxVersionLen-len("Version")+2], sourceHeader)

	for _, tool := range tools {
		if activeOnly {
			version := tool.ActiveVersion
			log.Printf("  %s%s%s%s%s", tool.Name, namePad[:maxNameLen-len(tool.Name)+2], colorstring.Green("%s", version), versionPad[:maxVersionLen-len(version)+2], tool.Source)
			continue
		}

		if tool.ActiveVersion != "" {
			log.Printf("  %s%s%s%s%s", tool.Name, namePad[:maxNameLen-len(tool.Name)+2], colorstring.Green("%s", tool.ActiveVersion), versionPad[:maxVersionLen-len(tool.ActiveVersion)+2], tool.Source)
			continue
		}

		if len(tool.InstalledVersions) == 0 {
			log.Printf("  %s%s(no versions installed)", tool.Name, namePad[:maxNameLen-len(tool.Name)+2])
			continue
		}

		version := tool.InstalledVersions[0]
		log.Printf("  %s%s%s%s%s", tool.Name, namePad[:maxNameLen-len(tool.Name)+2], version, versionPad[:maxVersionLen-len(version)+2], tool.Source)
		for i := 1; i < len(tool.InstalledVersions); i++ {
			log.Printf("  %s%s%s", namePad[:len(tool.Name)], namePad[:maxNameLen-len(tool.Name)+2], tool.InstalledVersions[i])
		}
	}

	log.Printf("")
}

func toolsLatest(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return fmt.Errorf("requires exactly 1 argument:\n%s", toolsLatestSubcommandUsageText)
	}

	toolSpec := args[0]
	format := c.String(toolsOutputFormatKey)
	checkInstalled := c.Bool(toolsInstalledKey)
	silent := false

	switch format {
	case outputFormatJSON:
		silent = true
	case outputFormatPlaintext:
		// valid format
	default:
		return fmt.Errorf("invalid --format: %s", format)
	}

	toolName, versionStr, err := parseToolSpec(toolSpec, false)
	if err != nil {
		return err
	}
	if versionStr == "latest" {
		return fmt.Errorf("invalid version prefix: %s (latest version is returned by default)", versionStr)
	}
	if versionStr == "installed" {
		return fmt.Errorf("invalid version prefix: %s (latest installed version is returned using --installed)", versionStr)
	}

	resolutionStrategy := provider.ResolutionStrategyLatestReleased
	if checkInstalled {
		resolutionStrategy = provider.ResolutionStrategyLatestInstalled
	}

	toolRequest := provider.ToolRequest{
		ToolName:           provider.ToolID(toolName),
		UnparsedVersion:    versionStr,
		ResolutionStrategy: resolutionStrategy,
		PluginURL:          nil,
	}

	// For tools latest, we'll use fast install regardless of the stack type
	useFastInstall := true

	// Get provider from flag, default to mise
	providerID := c.String(toolsProviderKey)
	if providerID == "" {
		providerID = "mise"
	}

	if providerID != "asdf" && providerID != "mise" {
		return fmt.Errorf("invalid provider: %s (must be 'asdf' or 'mise')", providerID)
	}

	version, err := toolprovider.GetLatestVersion(toolRequest, providerID, useFastInstall, silent)
	if err != nil {
		return err
	}

	// Output the version in the requested format
	switch format {
	case outputFormatJSON:
		data := map[string]string{
			"tool":    toolName,
			"version": version,
		}
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case outputFormatPlaintext:
		// For plaintext, just output the version string
		fmt.Println(version)
	}

	return nil
}

func toolsInstall(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return fmt.Errorf("requires exactly 1 argument:\n%s", toolInstallSubcommandUsageText)
	}

	toolSpec := args[0]
	providerID := c.String(toolsProviderKey)
	format := c.String(toolsOutputFormatKey)
	silent := false

	switch format {
	case outputFormatJSON, outputFormatBash:
		silent = true
	case outputFormatPlaintext:
		// valid format
	default:
		return fmt.Errorf("invalid --format: %s", format)
	}

	toolName, versionStr, err := parseToolSpec(toolSpec, true)
	if err != nil {
		return err
	}

	var strategy = provider.ResolutionStrategyStrict

	toolRequest := provider.ToolRequest{
		ToolName:           provider.ToolID(toolName),
		UnparsedVersion:    versionStr,
		ResolutionStrategy: strategy,
		PluginURL:          nil,
	}

	if providerID == "" {
		providerID = "mise"
	}

	if providerID != "asdf" && providerID != "mise" {
		return fmt.Errorf("invalid provider: %s (must be 'asdf' or 'mise')", providerID)
	}

	// For tools install, we'll use fast install regardless of the stack type
	useFastInstall := true

	tracker := analytics.NewDefaultTracker()
	envs, err := toolprovider.InstallSingleTool(toolRequest, providerID, useFastInstall, tracker, silent)
	if err != nil {
		return err
	}

	exposedWithEnvman := exposeEnvsWithEnvman(envs, silent)

	output, err := convertToOutputFormat(envs, format, exposedWithEnvman)
	if err != nil {
		return fmt.Errorf("convert to output format: %w", err)
	}
	fmt.Println(output)

	return nil
}

// parseToolSpec parses a tool specification in the format TOOL@VERSION or just TOOL.
// Returns toolName and version (if required and provided).
func parseToolSpec(toolSpec string, requireVersion bool) (toolName string, version string, err error) {
	parts := strings.Split(toolSpec, "@")

	if len(parts) > 2 {
		return "", "", fmt.Errorf("invalid tool specification: %s (expected TOOL@VERSION or TOOL)", toolSpec)
	}

	toolName = parts[0]
	if toolName == "" {
		return "", "", fmt.Errorf("tool name cannot be empty in: %s", toolSpec)
	}

	if len(parts) == 2 {
		version = parts[1]
	}

	if requireVersion && (len(parts) == 1 || version == "") {
		return "", "", fmt.Errorf("version cannot be empty in: %s (use format TOOL@VERSION)", toolSpec)
	}

	return toolName, version, nil
}
