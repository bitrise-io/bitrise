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

	toolsSetupCommandName = "setup"
	toolsInfoCommandName  = "info"

	toolsConfigKey      = "config"
	toolsConfigShortKey = "c"

	toolsWorkflowKey = "workflow"

	toolsOutputFormatKey      = "format"
	toolsOutputFormatShortKey = "f"
)

var toolsCommand = cli.Command{
	Name:  "tools",
	Usage: "Manage available tools from inside the workflow.",
	Subcommands: []cli.Command{
		{
			Name:        toolsSetupCommandName,
			Usage:       "Install tools from version files or bitrise config.",
			UsageText:   "bitrise tools setup [--config FILE]...",
			Description: "Install development tools from version files.",
			Action: func(c *cli.Context) error {
				logCommandParameters(c)
				if err := toolsSetup(c); err != nil {
					log.Errorf("Tool setup failed: %s", err)
					os.Exit(1)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  toolsConfigKey + ", " + toolsConfigShortKey,
					Usage: "Config or version file path(s) to install tools from. Can be specified multiple times. Auto-detects if not provided.",
				},
				cli.StringFlag{
					Name:  toolsWorkflowKey + ", w",
					Usage: "Workflow ID to use when installing from bitrise config (optional, uses global tools if not specified)",
				},
				cli.StringFlag{
					Name:  toolsOutputFormatKey + ", " + toolsOutputFormatShortKey,
					Usage: `Output format of the env vars that activate the tool. Options: plaintext (default), json, bash`,
					Value: outputFormatPlaintext,
				},
			},
		},
		{
			Name:        toolsInfoCommandName,
			Usage:       "Show information about installed tools.",
			UsageText:   "bitrise tools info [--json]",
			Description: "Display information about currently installed development tools.",
			Action: func(c *cli.Context) error {
				logCommandParameters(c)
				if err := toolsInfo(c); err != nil {
					log.Errorf("Failed to get tool info: %s", err)
					os.Exit(1)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  toolsOutputFormatKey + ", " + toolsOutputFormatShortKey,
					Usage: `Output format. Options: plaintext (default), json`,
					Value: outputFormatPlaintext,
				},
			},
		},
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
			return fmt.Errorf("file does not exist: %s", file)
		}

		if isYMLConfig(file) {
			if bitriseConfigPath != "" {
				return fmt.Errorf("multiple bitrise config files specified: %s and %s", bitriseConfigPath, file)
			}

			bitriseConfigPath = file
			continue
		}

		// Separate version files from bitrise config.
		versionFilePaths = append(versionFilePaths, file)
	}

	if bitriseConfigPath != "" {
		// Setting up from bitrise config.
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

func isYMLConfig(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml")
}

func convertToOutputFormat(envs []provider.EnvironmentActivation, format string, exposedWithEnvman bool) (string, error) {
	// TODO: is this valid for all formats?
	if len(envs) == 0 {
		return "", nil
	}

	envMap := toolprovider.ConvertToEnvMap(envs)

	var builder strings.Builder
	switch format {
	case outputFormatPlaintext:
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

func exposeEnvsWithEnvman(activations []provider.EnvironmentActivation, silent bool) bool {
	envs := toolprovider.ConvertToEnvmanEnvs(activations)
	err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, envs)
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

	tools, err := toolprovider.ListInstalledTools("mise")
	if err != nil {
		return err
	}

	if len(tools) == 0 {
		log.Infof("No tools installed")
		return nil
	}

	if format == outputFormatJSON {
		data, err := json.MarshalIndent(tools, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	log.Infof("Installed tools:")
	log.Printf("")

	maxNameLen := 0
	for _, tool := range tools {
		if len(tool.Name) > maxNameLen {
			maxNameLen = len(tool.Name)
		}
	}

	for _, tool := range tools {
		padding := strings.Repeat(" ", maxNameLen-len(tool.Name)+2)

		if tool.ActiveVersion != "" {
			log.Printf("  %s%s%s (active)", tool.Name, padding, tool.ActiveVersion)
		} else if len(tool.InstalledVersions) > 0 {
			log.Printf("  %s%s%s", tool.Name, padding, tool.InstalledVersions[0])
			for i := 1; i < len(tool.InstalledVersions); i++ {
				log.Printf("  %s%s%s", strings.Repeat(" ", len(tool.Name)), padding, tool.InstalledVersions[i])
			}
		} else {
			log.Printf("  %s%s(no versions installed)", tool.Name, padding)
		}
	}

	log.Printf("")
	return nil
}
