package cli

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
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
		if err := os.Setenv(bitrise.DebugModeEnvKey, "true"); err != nil {
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

	if err := os.Setenv(bitrise.LogLevelEnvKey, level.String()); err != nil {
		log.Fatal("Failed to set log level env:", err)
	}
	log.SetLevel(level)

	// CI Mode check
	if c.Bool(CIKey) {
		// if CI mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(bitrise.CIModeEnvKey, "true"); err != nil {
			return err
		}
		configs.IsCIMode = true
	}

	if err := bitrise.InitPaths(); err != nil {
		log.Fatalf("Failed to initialize required paths: %s", err)
	}

	if err := plugins.InitPaths(); err != nil {
		log.Fatalf("Failed to initialize required plugin paths: %s", err)
	}

	// Pull Request Mode check
	if c.Bool(PRKey) {
		// if PR mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(bitrise.PRModeEnvKey, "true"); err != nil {
			return err
		}
		configs.IsPullRequestMode = true
	}

	pullReqID := os.Getenv(bitrise.PullRequestIDEnvKey)
	if pullReqID != "" {
		configs.IsPullRequestMode = true
	}

	IsPR := os.Getenv(bitrise.PRModeEnvKey)
	if IsPR == "true" {
		configs.IsPullRequestMode = true
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
	app.Version = version.VERSION

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	app.Action = func(c *cli.Context) {
		log.Infof("Processing args: %v", c.Args())
		pluginName, pluginArgs, isPlugin := plugins.ParseArgs(c.Args())
		if isPlugin {
			log.Debugf("Try to run bitrise plugin: (%s) with args: (%v)", pluginName, pluginArgs)

			plugin, found, err := plugins.LoadPlugin(pluginName)
			if err != nil {
				log.Fatalf("Failed to get plugin (%s), err: %s", pluginName, err)
			}
			if !found {
				log.Fatalf("Plugin (%s) not installed", pluginName)
			}

			defer func() {
				if newVersion, err := plugins.CheckForNewVersion(plugin); err != nil {
					log.Fatalf("Failed to check for plugin (%s) new version", pluginName)
				} else if newVersion != "" {
					log.Warningf("New version (%s) of plugin (%s) available", newVersion, pluginName)
				} else {
					log.Debugf("No new version of plugin (%s) available", pluginName)
				}
			}()

			log.Debugf("Start plugin: (%s)", pluginName)
			if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
				log.Fatalf("Failed to run plugin (%s), err: %s", pluginName, err)
			}
		} else {
			cli.ShowAppHelp(c)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("Finished with Error:", err)
	}
}
