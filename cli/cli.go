package cli

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/codegangsta/cli"
)

func initLogFormatter() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "15:04:05",
	})
}

func before(c *cli.Context) error {
	initLogFormatter()
	initHelpAndVersionFlags()
	initAppHelpTemplate()

	// Debug mode?
	if c.Bool(DebugModeKey) {
		// set for other tools, as an ENV
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			return err
		}
		configs.IsDebugMode = true
		log.Warn("=> Started in DEBUG mode")
	}

	// Log level
	// If log level defined - use it
	logLevelStr := c.String(LogLevelKey)
	if logLevelStr == "" && configs.IsDebugMode {
		// if no Log Level defined and we're in Debug Mode - set loglevel to debug
		logLevelStr = "debug"
		log.Warn("=> LogLevel set to debug")
	}
	if logLevelStr == "" {
		// if still empty: set the default
		logLevelStr = "info"
	}

	level, err := log.ParseLevel(logLevelStr)
	if err != nil {
		return err
	}

	if err := os.Setenv(configs.LogLevelEnvKey, level.String()); err != nil {
		log.Fatal("Failed to set log level env:", err)
	}
	log.SetLevel(level)

	// CI Mode check
	if c.Bool(CIKey) {
		// if CI mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
			return err
		}
		configs.IsCIMode = true
	}

	// Pull Request Mode check
	if c.Bool(PRKey) {
		configs.IsPullRequestMode = true
	}
	if os.Getenv(configs.PullRequestIDEnvKey) != "" {
		configs.IsPullRequestMode = true
	}
	if os.Getenv(configs.PRModeEnvKey) == "true" {
		configs.IsPullRequestMode = true
	}

	if configs.IsPullRequestMode {
		// if PR mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.PRModeEnvKey, "true"); err != nil {
			return err
		}
	}

	// Analytics
	fileConfig, err := configs.ReadConfig()
	if err != nil {
		log.Warnf("Failed to read config at: (%s)", configs.GetBitriseConfigFilePath())
		log.Fatalf("Error: %#v", err)
	}
	configs.IsAnalyticsDisabled = fileConfig.IsAnalyticsDisabled

	isAnalyticsDisabledEnvValue := os.Getenv(configs.IsAnalyticsDisabledEnvKey)
	if isAnalyticsDisabledEnvValue != "" {
		if isAnalyticsDisabledEnvValue == "true" {
			configs.IsAnalyticsDisabled = true
		} else if isAnalyticsDisabledEnvValue == "false" {
			configs.IsAnalyticsDisabled = false
		} else {
			log.Warnf("Invalid IS_ANALYTICS_DISABLED env value: (%s), should be: (true/false)", isAnalyticsDisabledEnvValue)
		}
	}

	if c.IsSet(AnalyticsDisabledKey) {
		configs.IsAnalyticsDisabled = c.Bool(AnalyticsDisabledKey)
	}

	if configs.IsAnalyticsDisabled {
		// if Analytics disabled make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.IsAnalyticsDisabledEnvKey, "true"); err != nil {
			return err
		}
	}

	if err := bitrise.InitPaths(); err != nil {
		log.Fatalf("Failed to initialize required paths: %s", err)
	}

	return nil
}

func printVersion(c *cli.Context) {
	fmt.Fprintf(c.App.Writer, "%v\n", c.App.Version)
}

// Run ...
func Run() {
	// Parse cl
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Bitrise Automations Workflow Runner"
	app.Version = "1.3.0"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	app.Action = func(c *cli.Context) {
		pluginName, pluginType, pluginArgs, isPlugin := plugins.ParseArgs(c.Args())
		if isPlugin {
			log.SetLevel(log.DebugLevel)
			log.Debugln()
			log.Debugf("Try to run bitrise plugin: (%s) (type: %s) with args: (%v)", pluginName, pluginType, pluginArgs)

			printableName := plugins.PrintableName(pluginName, pluginType)
			log.Debugf("Plugin: %v", printableName)

			plugin, err := plugins.GetPlugin(pluginName, pluginType)
			if err != nil {
				log.Fatalf("Failed to get plugin (%s), err: %s", printableName, err)
			}

			messageFromPlugin, err := plugins.RunPlugin(app.Version, plugin, pluginArgs)
			log.Debugf("message from plugin: %s", messageFromPlugin)

			if err != nil {
				log.Fatalf("Failed to run plugin (%s), err: %s", printableName, err)
			}
		} else {
			cli.ShowAppHelp(c)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("Finished with Error:", err)
	}
}
