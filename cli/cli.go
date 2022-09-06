package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

func before(c *cli.Context) error {
	/*
		return err will print app's help also,
		use log.Fatal to avoid print help.
	*/
	//log.SetupLogger(c.String(log.OutputFormatKey), c.Bool(DebugModeKey))

	initHelpAndVersionFlags()

	// Debug mode?
	if c.Bool(DebugModeKey) {
		// set for other tools, as an ENV
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			log.Fatalf("Failed to set DEBUG env, error: %s", err)
		}
		configs.IsDebugMode = true
		log.Warn("=> Started in DEBUG mode")
	}

	// Log level
	// If log level defined - use it
	if c.String(LogLevelKey) == "" && configs.IsDebugMode {
		// if no Log Level defined and we're in Debug Mode - set loglevel to debug
		log.Warn("=> LogLevel set to debug")
		log.SetEnableDebugLog(true)
	}

	// CI Mode check
	if c.Bool(CIKey) {
		// if CI mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
			log.Fatalf("Failed to set CI env, error: %s", err)
		}
		configs.IsCIMode = true
	}

	if err := configs.InitPaths(); err != nil {
		log.Fatalf("Failed to initialize required paths, error: %s", err)
	}

	// Pull Request Mode check
	if c.Bool(PRKey) {
		// if PR mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.PRModeEnvKey, "true"); err != nil {
			log.Fatalf("Failed to set PR env, error: %s", err)
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

func printVersion(c *cli.Context) {
	log.Println(c.App.Version)
}

//func parseParams(arguments []string) (bool, string, bool) {
//	var isRunCommand bool
//	var runCommandIndex int
//
//	for i, argument := range arguments {
//		if argument == "run" {
//			isRunCommand = true
//			runCommandIndex = i
//		}
//	}
//
//	if isRunCommand == false {
//		return false, "", false
//	}
//
//	fs := flag.NewFlagSet("run", flag.ContinueOnError)
//	outputFormat := fs.String(log.OutputFormatKey, "", "test output")
//	isDebug := fs.Bool(DebugModeKey, false, "If true it enabled DEBUG mode. If no separate Log Level is specified this will also set the loglevel to debug.")
//	fs.Parse(arguments[runCommandIndex:])
//
//	return true, *outputFormat, *isDebug
//}

func loggerParameters(arguments []string) (bool, string, bool) {
	isRunCommand := false
	outputFormat := ""
	isDebug := false

	for i, argument := range arguments {
		if argument == "run" {
			isRunCommand = true
		}

		if argument == "--"+OutputFormatKey {
			if i+1 <= len(arguments) {
				value := arguments[i+1]

				if !strings.HasPrefix(value, "--") {
					outputFormat = value
				}
			}
		}

		if argument == "--"+DebugModeKey {
			if i+1 <= len(arguments) {
				value := arguments[i+1]

				if strings.HasPrefix(value, "--") {
					isDebug = true
				} else {
					value, err := strconv.ParseBool(value)
					if err == nil {
						isDebug = value
					}
				}
			}
		}
	}

	return isRunCommand, outputFormat, isDebug
}

// Run ...
func Run() {
	isRunCommand, format, isDebug := loggerParameters(os.Args[1:])
	//isRunCommand, format, isDebug := parseParams(os.Args[1:])
	if isRunCommand && format != "" {
		log.SetOutputFormat(format)
	}

	log.SetEnableDebugLog(isDebug)

	if err := plugins.InitPaths(); err != nil {
		log.Fatalf("Failed to initialize plugin path, error: %s", err)
	}

	cli.VersionPrinter = printVersion
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
				log.Fatalf("Setup failed, error: %s", err)
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
		log.Fatal(err)
	}
}
