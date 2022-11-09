package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

// Run ...
func Run() {
	// In the case of `--output-format=json` flag is set for the run command, all the logs are expected in JSON format.
	// Because logs might be printed before processing the run command args,
	// we need to manually parse the logger configuration.
	isRunCommand, logFormat, isDebugMode := loggerParameters(os.Args[1:])

	if !isDebugMode {
		isDebugMode = os.Getenv(configs.DebugModeEnvKey) == "true"
	}

	loggerType := log.ConsoleLogger
	if isRunCommand && string(logFormat) != "" {
		loggerType = logFormat
	}

	// Global logger needs to be initialised before using any log function
	opts := log.LoggerOpts{
		LoggerType:      loggerType,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: isDebugMode,
		Writer:          os.Stdout,
		TimeProvider:    time.Now,
	}
	log.InitGlobalLogger(opts)

	if isDebugMode {
		// set for other tools, as an ENV
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			failf("Failed to set DEBUG env, error: %s", err)

		}

		if err := os.Setenv("LOGLEVEL", "debug"); err != nil {
			failf("Failed to set LOGLEVEL env, error: %s", err)

		}

		configs.IsDebugMode = true
	}

	// This is needed for the getPluginsList func in the cli.AppHelpTemplate
	// and cli.AppHelpTemplate is evaluated before executing app.Before.
	if err := plugins.InitPaths(); err != nil {
		failf("Failed to initialize plugin path, error: %s", err)
	}

	cli.VersionPrinter = func(c *cli.Context) { log.Print(c.App.Version) }
	cli.AppHelpTemplate = fmt.Sprintf(helpTemplate, getPluginsList())

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Bitrise Automations Workflow Runner"
	app.Version = version.VERSION

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	app.Action = func(c *cli.Context) error {
		pluginName, pluginArgs, isPlugin := plugins.ParseArgs(c.Args())
		if isPlugin {
			plugin, found, err := plugins.LoadPlugin(pluginName)
			if err != nil {
				return fmt.Errorf("Failed to get plugin (%s), error: %s", pluginName, err)
			}
			if !found {
				return fmt.Errorf("Plugin (%s) not installed", pluginName)
			}

			if err := bitrise.RunSetupIfNeeded(version.VERSION, false); err != nil {
				failf("Setup failed, error: %s", err)
			}

			if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
				return fmt.Errorf("Failed to run plugin (%s), error: %s", pluginName, err)
			}
		} else {
			if err := cli.ShowAppHelp(c); err != nil {
				return fmt.Errorf("Failed to show help, error: %s", err)
			}
			return errors.New("")
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		failf(err.Error())
	}
}

func loggerParameters(arguments []string) (isRunCommand bool, outputFormat log.LoggerType, isDebug bool) {
	for i, argument := range arguments {
		if argument == "run" {
			isRunCommand = true
		}

		if argument == "--"+OutputFormatKey || argument == "-"+OutputFormatKey {
			// syntax
			// -flag
			// --flag   // double dashes are also permitted
			// -flag=x
			// -flag x  // non-boolean flags only
			// One or two dashes may be used; they are equivalent.
			// https://pkg.go.dev/flag#hdr-Command_line_flag_syntax
			if i+1 < len(arguments) {
				value := arguments[i+1]
				switch value {
				case string(log.JSONLogger):
					outputFormat = log.JSONLogger
				case string(log.ConsoleLogger):
					outputFormat = log.ConsoleLogger
				default:
					// At this point we don't care about invalid values,
					// the execution will fail when parsing the command's arguments.
				}
			}
		}
		if strings.HasPrefix(argument, "--"+OutputFormatKey+"=") || strings.HasPrefix(argument, "-"+OutputFormatKey+"=") {
			flagValue := strings.TrimPrefix(argument, "--"+OutputFormatKey+"=")
			flagValue = strings.TrimPrefix(flagValue, "-"+OutputFormatKey+"=")
			if flagValue == string(log.JSONLogger) {
				outputFormat = log.JSONLogger
			} else if flagValue == string(log.ConsoleLogger) {
				outputFormat = log.ConsoleLogger
			}
		}

		if argument == "--"+DebugModeKey || argument == "-"+DebugModeKey {
			// "-flag x" Command line flag syntax is not supported for boolean flags
			// https://pkg.go.dev/flag#hdr-Command_line_flag_syntax
			isDebug = true
		}
		if strings.HasPrefix(argument, "--"+DebugModeKey+"=") || strings.HasPrefix(argument, "-"+DebugModeKey+"=") {
			flagValue := strings.TrimPrefix(argument, "--"+DebugModeKey+"=")
			flagValue = strings.TrimPrefix(flagValue, "-"+DebugModeKey+"=")
			value, err := strconv.ParseBool(flagValue)
			if err == nil {
				isDebug = value
			}
		}
	}

	return
}

func before(c *cli.Context) error {
	/*
		return err will print app's help also,
		use log.Fatal to avoid print help.
	*/

	initHelpAndVersionFlags()

	// CI Mode check
	if c.Bool(CIKey) {
		// if CI mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
			failf("Failed to set CI env, error: %s", err)
		}
		configs.IsCIMode = true
	}

	if err := configs.InitPaths(); err != nil {
		failf("Failed to initialize required paths, error: %s", err)
	}

	// Pull Request Mode check
	if c.Bool(PRKey) {
		// if PR mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.PRModeEnvKey, "true"); err != nil {
			failf("Failed to set PR env, error: %s", err)
		}
		configs.IsPullRequestMode = true
	}

	pullReqID := os.Getenv(configs.PullRequestIDEnvKey)
	if pullReqID != "" {
		configs.IsPullRequestMode = true
	}

	IsPR := os.Getenv(configs.PRModeEnvKey)
	if IsPR == "true" {
		configs.IsPullRequestMode = true
	}

	return nil
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}
