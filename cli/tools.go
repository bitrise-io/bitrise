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
	Description: `Show installed tool versions. Use --active to show only tools currently active in the shell context.

EXAMPLES:
   bitrise tools info
   bitrise tools info --active
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
	toolInstallSubcommandUsageText = "bitrise tools install [--provider PROVIDER] [--format FORMAT] <TOOL> <VERSION>[:SUFFIX]"
	toolsInstallSubcommand         = cli.Command{
		Name:      toolsInstallSubcommandName,
		Usage:     "Install a specific tool version",
		UsageText: toolInstallSubcommandUsageText,
		Description: `Install a specific version of a tool using the configured tool provider.

TOOL: tool name (e.g., nodejs, ruby, python, go, etc.)
VERSION: specific version (20.10.0), prefix (22), latest, or installed.

EXAMPLES:
   bitrise tools install nodejs 20.10.0
   bitrise tools install nodejs 22:latest
   bitrise tools install ruby installed
   eval "$(bitrise tools install ruby 3.2.0 --format bash)"  # activate in shell`,
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
	toolsLatestSubcommandUsageText = "bitrise tools latest [--provider PROVIDER] [--format FORMAT] <TOOL> [VERSION[:SUFFIX]]"
	toolsLatestSubcommand          = cli.Command{
		Name:      toolsLatestSubcommandName,
		Usage:     "Query the latest version of a tool",
		UsageText: toolsLatestSubcommandUsageText,
		Description: `Query the latest version of a tool, optionally matching a version prefix.

By default, queries latest available release. Use :installed suffix for latest installed version.

EXAMPLES:
   bitrise tools latest nodejs
   bitrise tools latest nodejs 20
   bitrise tools latest python 3.12:installed
   bitrise tools latest ruby installed`,
		Action: func(c *cli.Context) error {
			logCommandParameters(c)
			if err := toolsLatest(c); err != nil {
				log.Errorf("Tool latest failed: %s", err)
				os.Exit(1)
			}
			return nil
		},
		Flags: []cli.Flag{
			flToolsOutputFormat,
			flToolsProvider,
		},
	}
)

var toolsSetupSubcommand = cli.Command{
	Name:      toolsSetupSubcommandName,
	Usage:     "Install tools from version files or bitrise.yml",
	UsageText: "bitrise tools setup [--config FILE]...",
	Description: `Install tools from version files (e.g. .tool-versions, .node-version, etc.) or bitrise.yml.

EXAMPLES:
   bitrise tools setup --config .tool-versions
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
		// Valid formats.
		silent = true
	case outputFormatPlaintext:
		// Valid format.
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

	// Check if envstore exists - it should be initialized by the workflow runner.
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
	silent := false

	switch format {
	case outputFormatJSON:
		silent = true
	case outputFormatPlaintext:
		// Valid format.
	default:
		return fmt.Errorf("invalid --format: %s", format)
	}

	tools, err := toolprovider.ListInstalledTools("mise", activeOnly, silent)
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

// parseToolCommand parses the common tool command parameters for install and latest subcommands.
func parseToolCommand(c *cli.Context, isInstall bool) (request provider.ToolRequest, providerID, format, toolName string, silent bool, err error) {
	args := c.Args()
	providerID = c.String(toolsProviderKey)
	format = c.String(toolsOutputFormatKey)
	silent = false
	request = provider.ToolRequest{}

	if isInstall {
		// Install needs version.
		if len(args) != 2 {
			err = fmt.Errorf("requires 2 arguments")
			return
		}
	} else {
		if len(args) < 1 || len(args) > 2 {
			err = fmt.Errorf("requires 1 or 2 arguments")
			return
		}
	}

	toolName = args[0]
	if toolName == "" {
		err = fmt.Errorf("tool name cannot be empty")
		return
	}

	versionString := ""
	if len(args) >= 2 {
		versionString = args[1]
	}

	switch format {
	case outputFormatJSON:
		silent = true
	case outputFormatPlaintext:
		// Valid format.
	case outputFormatBash:
		// Install allows bash format for activation in shell.
		if isInstall {
			silent = true
			break
		}
		// If not install, fallthrough to error.
		fallthrough
	default:
		err = fmt.Errorf("invalid --format: %s", format)
		return
	}

	version, resolutionStrategy, parseErr := toolprovider.ParseVersionString(versionString)
	if parseErr != nil {
		err = fmt.Errorf("parse version string: %w", parseErr)
		return
	}

	request = provider.ToolRequest{
		ToolName:           provider.ToolID(toolName),
		UnparsedVersion:    version,
		ResolutionStrategy: resolutionStrategy,
		PluginURL:          nil,
	}

	if providerID == "" {
		providerID = "mise"
	}

	if providerID != "asdf" && providerID != "mise" {
		err = fmt.Errorf("invalid provider: %s (must be 'asdf' or 'mise')", providerID)
		return
	}

	return
}

func toolsLatest(c *cli.Context) error {
	toolRequest, providerID, format, toolName, silent, err := parseToolCommand(c, false)
	if err != nil {
		return err
	}

	useFastInstall := true
	resultVersion, err := toolprovider.GetLatestVersion(toolRequest, providerID, useFastInstall, silent)
	if err != nil {
		return err
	}

	// Output the version in the requested format.
	switch format {
	case outputFormatJSON:
		data := map[string]string{
			"tool":    toolName,
			"version": resultVersion,
		}
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case outputFormatPlaintext:
		// For plaintext, just output the version string.
		fmt.Println(resultVersion)
	}

	return nil
}

func toolsInstall(c *cli.Context) error {
	toolRequest, providerID, format, _, silent, err := parseToolCommand(c, true)
	if err != nil {
		return err
	}

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
