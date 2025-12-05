package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/urfave/cli"
)

const (
	outputFormatPlaintext = "plaintext"
	outputFormatJSON      = "json"
	outputFormatBash      = "bash"
)

var toolsCommand = cli.Command{
	Name:  "tools",
	Usage: "Manage development tools.",
	Subcommands: []cli.Command{
		{
			Name:        "setup",
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
					Name:  ConfigKey + ", " + configShortKey,
					Usage: "Config or version file path(s) to install tools from. Can be specified multiple times. Auto-detects if not provided.",
				},
				cli.StringFlag{
					Name:  "provider",
					Usage: "Tool provider to use (asdf, mise). Default: mise",
					Value: "mise",
				},
				cli.StringFlag{
					Name:  WorkflowKey + ", w",
					Usage: "Workflow ID to use when installing from bitrise config (optional, uses global tools if not specified)",
				},
				cli.BoolFlag{
					Name:  "fast-install",
					Usage: "Enable experimental fast install (currently Ruby only with mise)",
				},
				cli.StringFlag{
					Name:  "format, f",
					Usage: `Output format of the env vars that activate the tool. Options: plaintext (default), json, bash`,
					Value: outputFormatPlaintext,
				},
			},
		},
		{
			Name:        "info",
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
					Name:  "format, f",
					Usage: `Output format. Options: plaintext (default), json`,
					Value: outputFormatPlaintext,
				},
				cli.StringFlag{
					Name:  "provider",
					Usage: "Tool provider to query (asdf, mise). Default: mise",
					Value: "mise",
				},
			},
		},
	},
}

func toolsSetup(c *cli.Context) error {
	configFiles := c.StringSlice(ConfigKey)
	provider := c.String("provider")
	fastInstall := c.Bool("fast-install")
	workflowID := c.String(WorkflowKey)

	format := c.String("format")
	switch format {
	case outputFormatPlaintext, outputFormatJSON, outputFormatBash:
		// valid formats
	default:
		return fmt.Errorf("invalid --format: %s", format)
	}
	silent := format == outputFormatJSON || format == outputFormatBash

	// Check if any file looks like a bitrise config.
	var bitriseConfigPath string
	for _, file := range configFiles {
		if isBitriseConfig(file) {
			bitriseConfigPath = file
			break
		}
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
		envs, err := toolprovider.Run(config, tracker, false, workflowID, silent)
		if err != nil {
			return err
		}

		output, err := convertToOutputFormat(envs, format)
		if err != nil {
			return fmt.Errorf("convert to output format: %w", err)
		}
		fmt.Println(output)
	}

	opts := toolprovider.SetupOptions{
		VersionFiles:            configFiles,
		ProviderName:            provider,
		ExperimentalFastInstall: fastInstall,
	}

	tracker := analytics.NewDefaultTracker()
	envs, err := toolprovider.SetupFromVersionFiles(opts, tracker, silent)
	if err != nil {
		return err
	}

	output, err := convertToOutputFormat(envs, format)
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

func convertToOutputFormat(envs []provider.EnvironmentActivation, format string) (string, error) {
	if len(envs) == 0 {
		return "", nil
	}

	switch format {
	case outputFormatPlaintext:
		var builder strings.Builder
		builder.WriteString("Env vars to activate installed tools:\n")
		for _, env := range envs {
			for k, v := range env.ContributedEnvVars {
				builder.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
			newPaths := strings.Join(env.ContributedPaths, ":")
			builder.WriteString(fmt.Sprintf("PATH=%s:$PATH\n", newPaths))
		}
		return builder.String(), nil
	case outputFormatJSON:
		envMap := make(map[string]string)
		for _, env := range envs {
			// TODO: we should probably deduplicate keys (especially PATH), see toolprovider/env.go
			for k, v := range env.ContributedEnvVars {
				envMap[k] = v
			}
			newPaths := strings.Join(env.ContributedPaths, ":")
			envMap["PATH"] = fmt.Sprintf("%s:$PATH", newPaths)
		}
		data, err := json.MarshalIndent(envMap, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal JSON: %w", err)
		}
		return string(data), nil
	case outputFormatBash:
		var builder strings.Builder
		for _, env := range envs {
			for k, v := range env.ContributedEnvVars {
				builder.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, v))
			}
			newPaths := strings.Join(env.ContributedPaths, ":")
			value := fmt.Sprintf("%s:$PATH", newPaths)
			builder.WriteString(fmt.Sprintf("export PATH=\"%s\"\n", value))
		}
		return builder.String(), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func toolsInfo(c *cli.Context) error {
	provider := c.String("provider")
	format := c.String("format")

	tools, err := toolprovider.ListInstalledTools(provider)
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

	log.Infof("Installed tools (%s):", provider)
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
