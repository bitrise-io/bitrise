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
	// In the case of `--output-format=json` flag is set for the run command, all the logs are expected in JSON format.
	// Because logs might be printed before processing the run command args,
	// we need to manually parse the logger configuration.
	isRunCommand, logFormat := loggerParameters(os.Args[1:])
	debugMode := isDebugMode(os.Args[1:])

	loggerType := log.ConsoleLogger
	if isRunCommand && string(logFormat) != "" {
		loggerType = logFormat
	}

	// Global logger needs to be initialised before using any log function
	opts := log.LoggerOpts{
		LoggerType:      loggerType,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: debugMode,
		Writer:          os.Stdout,
		TimeProvider:    time.Now,
	}
	log.InitGlobalLogger(opts)

	if debugMode {
		// set for other tools, as an ENV
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			failf("Failed to set DEBUG env, error: %s", err)

		}

		if err := os.Setenv("LOGLEVEL", "debug"); err != nil {
			failf("Failed to set LOGLEVEL env, error: %s", err)

		}

		configs.IsDebugMode = true
	}

	// This is needed for the getPluginsList func in the root help output,
	// which is evaluated before executing the command's PersistentPreRunE.
	if err := plugins.InitPaths(); err != nil {
		failf("Failed to initialize plugin path, error: %s", err)
	}

	globalTracker = analytics.NewDefaultTracker()
	defer func() {
		globalTracker.Wait()
	}()

	validateGlobalBoolEnvs()

	rootCmd := newRootCommand()

	rawArgs := os.Args[1:]
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

	normalized := normalizeLegacyArgs(rawArgs, rootCmd)

	// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
	// An unknown top-level command is not a plugin and not a known command, so
	// cobra's Find returns an error. The previous framework printed the app help
	// and exited 1 in that case.
	if _, _, err := rootCmd.Find(normalized); err != nil {
		printRootHelp(rootCmd)
		failf("")
	}

	rootCmd.SetArgs(normalized)
	if err := rootCmd.Execute(); err != nil {
		failf(err.Error())
	}
}

func loggerParameters(arguments []string) (isRunCommand bool, outputFormat log.LoggerType) {
	for i, argument := range arguments {
		if argument == "run" {
			isRunCommand = true
		}

		// syntax
		// -flag
		// --flag   // double dashes are also permitted
		// -flag=x
		// -flag x  // non-boolean flags only
		// One or two dashes may be used; they are equivalent.
		// https://pkg.go.dev/flag#hdr-Command_line_flag_syntax
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

	pullReqID := os.Getenv(configs.PullRequestIDEnvKey)
	if pullReqID != "" {
		configs.IsPullRequestMode = true
	}

	IsPR := os.Getenv(configs.PRModeEnvKey)
	if IsPR == "true" {
		configs.IsPullRequestMode = true
	}

	// want to access this key in setup command too
	isOfflineMode := isSteplibOfflineMode()
	registerSteplibOfflineMode(isOfflineMode)

	return nil
}

// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
// validateGlobalBoolEnvs aborts when a global bool flag's bound environment
// variable holds a non-bool value, matching the behaviour of the previous
// framework (an empty value is allowed and treated as false).
func validateGlobalBoolEnvs() {
	for _, envKey := range []string{configs.CIModeEnvKey, configs.DebugModeEnvKey} {
		if val, ok := os.LookupEnv(envKey); ok && val != "" {
			if _, err := strconv.ParseBool(val); err != nil {
				failf("could not parse %q as bool value for $%s", val, envKey)
			}
		}
	}
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	globalTracker.Wait()
	os.Exit(1)
}
