package cli

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

// Run is the entry point for the Bitrise CLI.
func Run() {
	// Normalize single-dash long flags (e.g. -ci, -debug) that urfave/cli v1 accepted
	// but pflag requires as double-dash. Also maps -v → --version.
	cliArgs := normalizeLegacyArgs(os.Args[1:])

	// In the case of --output-format=json flag is set for the run command, all the logs are expected in JSON format.
	// Because logs might be printed before processing the run command args,
	// we need to manually parse the logger configuration.
	isRunCommand, logFormat, isDebugMode := loggerParameters(cliArgs)

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
		if err := os.Setenv(configs.DebugModeEnvKey, "true"); err != nil {
			failf("Failed to set DEBUG env, error: %s", err)
		}
		if err := os.Setenv("LOGLEVEL", "debug"); err != nil {
			failf("Failed to set LOGLEVEL env, error: %s", err)
		}
		configs.IsDebugMode = true
	}

	// This is needed for the getPluginsList func used in the help function,
	// which is evaluated before executing the PersistentPreRunE.
	if err := plugins.InitPaths(); err != nil {
		failf("Failed to initialize plugin path, error: %s", err)
	}

	rootCmd := &cobra.Command{
		Use:               path.Base(os.Args[0]),
		Short:             "Bitrise Automations Workflow Runner",
		Version:           version.VERSION,
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: before,
	}

	rootCmd.PersistentFlags().BoolVar(&debugMode, DebugModeKey, false, "If true it enables DEBUG mode.")
	rootCmd.PersistentFlags().BoolVar(&ciMode, CIKey, false, "If true it indicates that we're used by another tool so don't require any user input!")
	rootCmd.PersistentFlags().BoolVar(&prMode, PRKey, false, "If true bitrise runs in pull request mode.")

	// Capture cobra's default help before overriding, so subcommands can use it.
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd != cmd.Root() {
			defaultHelp(cmd, args)
			return
		}
		fmt.Printf("NAME: %s - %s\n\n", cmd.Root().Name(), cmd.Root().Short)
		fmt.Printf("USAGE: %s [OPTIONS] COMMAND/PLUGIN [arg...]\n\n", cmd.Root().Name())
		fmt.Printf("VERSION: %s\n\n", cmd.Root().Version)
		fmt.Printf("GLOBAL OPTIONS:\n")
		cmd.Root().PersistentFlags().PrintDefaults()
		fmt.Printf("\nCOMMANDS:\n")
		for _, c := range cmd.Root().Commands() {
			if !c.Hidden {
				fmt.Printf("  %-15s %s\n", c.Name(), c.Short)
			}
		}
		fmt.Printf("\n%s\n", getPluginsList())
		fmt.Printf("COMMAND HELP: %s COMMAND --help/-h\n", cmd.Root().Name())
	})

	registerCommands(rootCmd)
	rootCmd.SetArgs(cliArgs)

	globalTracker = analytics.NewDefaultTracker()
	defer func() {
		globalTracker.Wait()
	}()

	if err := rootCmd.Execute(); err != nil {
		// cobra returns an error for unknown commands; try plugin routing
		if strings.Contains(err.Error(), "unknown command") {
			pluginName, pluginArgs, isPlugin := plugins.ParseArgs(cliArgs)
			if isPlugin {
				logPluginCommandParameters(pluginName, pluginArgs)

				// cobra errored before routing so PersistentPreRunE never fired and
				// no flags were parsed. Apply global flags from cliArgs via pflag.Set()
				// so that before() can correctly use Changed() to distinguish explicit
				// flag values from defaults (important for e.g. --ci=false overriding
				// a CI_MODE=true env var).
				for _, arg := range cliArgs {
					for _, key := range []string{DebugModeKey, CIKey, PRKey} {
						if arg == "--"+key {
							_ = rootCmd.PersistentFlags().Set(key, "true")
						} else if val, ok := strings.CutPrefix(arg, "--"+key+"="); ok {
							_ = rootCmd.PersistentFlags().Set(key, val)
						}
					}
				}

				if err := before(rootCmd, nil); err != nil {
					failf("Setup failed, error: %s", err)
				}

				logger := log.NewLogger(log.GetGlobalLoggerOpts())
				plugin, found, pluginErr := plugins.LoadPlugin(pluginName)
				if pluginErr != nil {
					failf("failed to get plugin (%s), error: %s", pluginName, pluginErr)
				}
				if !found {
					failf("plugin (%s) not installed", pluginName)
				}

				if err := bitrise.RunSetupIfNeeded(logger); err != nil {
					failf("Setup failed, error: %s", err)
				}

				if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
					failf("failed to run plugin (%s), error: %s", pluginName, err)
				}
			} else {
				if helpErr := rootCmd.Help(); helpErr != nil {
					failf("failed to show help: %s", helpErr)
				}
				failf(err.Error())
			}
		} else {
			failf(err.Error())
		}
	}
}

func loggerParameters(arguments []string) (isRunCommand bool, outputFormat log.LoggerType, isDebug bool) {
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

		if isFlag(DebugModeKey, argument) {
			components := strings.Split(argument, "=")
			if len(components) == 2 {
				value, err := strconv.ParseBool(components[1])
				if err == nil {
					isDebug = value
				}
			} else {
				components := strings.Split(argument, " ")

				// "-flag x" Command line flag syntax is not supported for boolean flags
				// https://pkg.go.dev/flag#hdr-Command_line_flag_syntax
				if len(components) == 1 {
					isDebug = true
				}
			}
		}
	}

	return
}

func isFlag(name, arg string) bool {
	return arg == "--"+name || arg == "-"+name ||
		strings.HasPrefix(arg, "--"+name+"=") || strings.HasPrefix(arg, "-"+name+"=")
}

func before(cmd *cobra.Command, args []string) error {
	flags := cmd.Root().PersistentFlags()

	// Apply env-var fallback only when the flag was NOT explicitly set on the
	// command line. Using Changed() rather than checking the current bool value
	// ensures that --ci=false explicitly overrides CI_MODE=true env var.
	if !flags.Changed(DebugModeKey) && os.Getenv(configs.DebugModeEnvKey) == "true" {
		debugMode = true
	}
	if !flags.Changed(CIKey) && os.Getenv(configs.CIModeEnvKey) == "true" {
		ciMode = true
	}

	if ciMode {
		if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
			failf("Failed to set CI env, error: %s", err)
		}
		configs.IsCIMode = true
	}

	if err := configs.InitPaths(); err != nil {
		failf("Failed to initialize required paths, error: %s", err)
	}

	if prMode {
		if err := os.Setenv(configs.PRModeEnvKey, "true"); err != nil {
			failf("Failed to set PR env, error: %s", err)
		}
		configs.IsPullRequestMode = true
	}

	pullReqID := os.Getenv(configs.PullRequestIDEnvKey)
	if pullReqID != "" {
		configs.IsPullRequestMode = true
	}

	isPR := os.Getenv(configs.PRModeEnvKey)
	if isPR == "true" {
		configs.IsPullRequestMode = true
	}

	isOfflineMode := isSteplibOfflineMode()
	registerSteplibOfflineMode(isOfflineMode)

	return nil
}

func failf(format string, args ...any) {
	log.Errorf(format, args...)
	globalTracker.Wait()
	os.Exit(1)
}

// normalizeLegacyArgs converts single-dash long flags that urfave/cli v1 accepted
// but pflag/cobra requires in double-dash form. Also maps -v → --version.
func normalizeLegacyArgs(args []string) []string {
	replacements := map[string]string{
		"-" + DebugModeKey: "--" + DebugModeKey,
		"-" + CIKey:        "--" + CIKey,
		"-" + PRKey:        "--" + PRKey,
		"-v":               "--version",
	}
	out := make([]string, len(args))
	copy(out, args)
	for i, arg := range out {
		for old, new := range replacements {
			if arg == old || strings.HasPrefix(arg, old+"=") {
				out[i] = new + strings.TrimPrefix(arg, old)
				break
			}
		}
	}
	return out
}
