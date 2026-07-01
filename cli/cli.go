package cli

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

// Run ...
func Run() {
	rawArgs := os.Args[1:]

	initLogger(rawArgs)

	// This is needed for printInstalledPlugins in the root help output, which is
	// evaluated before executing the command's PersistentPreRunE.
	if err := plugins.InitPaths(); err != nil {
		failf("Failed to initialize plugin path, error: %s", err)
	}

	globalTracker = analytics.NewDefaultTracker()
	defer globalTracker.Wait()

	// Abort when a global bool flag's bound env var holds a non-bool value (an
	// empty value is allowed and treated as unset).
	for _, envKey := range []string{configs.CIModeEnvKey, configs.DebugModeEnvKey} {
		if _, err := resolveBoolEnv(envKey); err != nil {
			failf("%s", err)
		}
	}

	rootCmd := newRootCommand()

	if pluginName, pluginArgs, isPlugin := detectPlugin(rootCmd, rawArgs); isPlugin {
		runPlugin(rootCmd, rawArgs, pluginName, pluginArgs)
		return
	}

	// envman is a passthrough command: it must receive its args verbatim, so it
	// is dispatched before cobra to keep the global flags (which precede the
	// command) from being forwarded into the passthrough.
	if envmanArgs, isEnvman := envmanPassthrough(rawArgs); isEnvman {
		runEnvman(rootCmd, rawArgs, envmanArgs)
		return
	}

	rootCmd.SetArgs(rawArgs)
	if err := rootCmd.Execute(); err != nil {
		failf("%s", err)
	}
}

// initLogger sets up the global logger up front, before cobra parses the args,
// because log output can happen before the command itself runs.
func initLogger(arguments []string) {
	// For `--output-format=json` on the run command all logs are expected in JSON.
	// Because logs might be printed before the run command args are processed, we
	// parse the logger configuration manually here.
	isRunCommand, logFormat := loggerParameters(arguments)
	loggerType := log.ConsoleLogger
	if isRunCommand && logFormat != "" {
		loggerType = logFormat
	}

	// An explicit --debug flag wins (matching the --ci precedence); otherwise the
	// bound DEBUG env decides. cobra re-parses the same flag later for help,
	// analytics and the command itself; this early pass only feeds the logger.
	// "--flag x" syntax is not supported for bool flags, so only the bare flag and
	// the "--debug=x" form are accepted.
	debugMode := false
	debugSetByFlag := false
	for _, argument := range arguments {
		if !isFlag(DebugModeKey, argument) {
			continue
		}
		if _, raw, ok := strings.Cut(argument, "="); ok {
			if parsed, err := strconv.ParseBool(raw); err == nil {
				debugMode, debugSetByFlag = parsed, true
			}
		} else {
			debugMode, debugSetByFlag = true, true
		}
	}
	if !debugSetByFlag {
		debugMode, _ = strconv.ParseBool(os.Getenv(configs.DebugModeEnvKey))
	}

	log.InitGlobalLogger(log.LoggerOpts{
		LoggerType:      loggerType,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: debugMode,
		Writer:          os.Stdout,
		TimeProvider:    time.Now,
	})

	if debugMode {
		// propagate to other tools (and our own log level) via env
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			failf("Failed to set DEBUG env, error: %s", err)
		}
		if err := os.Setenv("LOGLEVEL", "debug"); err != nil {
			failf("Failed to set LOGLEVEL env, error: %s", err)
		}
		configs.IsDebugMode = true
	}
}

func loggerParameters(arguments []string) (isRunCommand bool, outputFormat log.LoggerType) {
	for i, argument := range arguments {
		if argument == "run" {
			isRunCommand = true
		}

		// Long flags use the double-dash form only:
		//   --output-format value
		//   --output-format=value
		if isFlag(OutputFormatKey, argument) {
			var value string
			components := strings.Split(argument, "=")

			// If the flag value was specified with an `=` mark then the second element in the array is the actual value.
			// Otherwise, the value was specified as a separate item after the flag, and we need to take the next
			// argument value.
			if len(components) == 2 {
				value = components[1]
			} else if i+1 < len(arguments) {
				value = arguments[i+1]
			}

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

	return
}

func before(cmd *cobra.Command, _ []string) error {
	root := cmd.Root()

	// CI Mode check. The --ci flag is seeded from the CI env var when not set
	// explicitly on the command line (an explicit --ci=false still wins).
	isCI, _ := root.PersistentFlags().GetBool(CIKey)
	if !root.PersistentFlags().Changed(CIKey) {
		if envCI, err := strconv.ParseBool(os.Getenv(configs.CIModeEnvKey)); err == nil {
			isCI = envCI
		}
	}
	if isCI {
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
	if isPR, _ := root.PersistentFlags().GetBool(PRKey); isPR {
		// if PR mode indicated make sure we set the related env
		//  so all other tools we use will also get it
		if err := os.Setenv(configs.PRModeEnvKey, "true"); err != nil {
			failf("Failed to set PR env, error: %s", err)
		}
		configs.IsPullRequestMode = true
	}

	if os.Getenv(configs.PullRequestIDEnvKey) != "" {
		configs.IsPullRequestMode = true
	}
	if os.Getenv(configs.PRModeEnvKey) == "true" {
		configs.IsPullRequestMode = true
	}

	// want to access this key in setup command too
	isOfflineMode := isSteplibOfflineMode()
	registerSteplibOfflineMode(isOfflineMode)

	return nil
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	globalTracker.Wait()
	os.Exit(1)
}
