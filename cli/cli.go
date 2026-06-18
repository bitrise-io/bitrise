package cli

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	rootCmd.SetArgs(normalizeLegacyArgs(rawArgs, rootCmd))
	if err := rootCmd.Execute(); err != nil {
		failf(err.Error())
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

// globalBoolFlag returns the value of a persistent (global) flag, read from the
// root command so it is also available on the plugin dispatch path.
func globalBoolFlag(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool(name)
	return v
}

// globalFlagChanged reports whether a persistent (global) flag was explicitly
// set on the command line (so an explicit value takes precedence over env vars).
func globalFlagChanged(cmd *cobra.Command, name string) bool {
	return cmd.Root().PersistentFlags().Changed(name)
}

// ciModeFlagOverride mirrors the --ci flag's CI env var binding: the override is
// set (non-nil) when --ci was passed or the CI env var is present, with the value
// taken from the flag, otherwise the parsed env value. A non-nil result takes
// precedence over inventory-based CI detection.
func ciModeFlagOverride(cmd *cobra.Command) *bool {
	if globalFlagChanged(cmd, CIKey) {
		return pointers.NewBoolPtr(globalBoolFlag(cmd, CIKey))
	}
	if val, ok := os.LookupEnv(configs.CIModeEnvKey); ok {
		parsed, _ := strconv.ParseBool(val)
		return pointers.NewBoolPtr(parsed)
	}
	return nil
}

// secretFilteringFlagOverride mirrors the trigger command's secret-filtering
// flag, which was bound to the BITRISE_SECRET_FILTERING env var: the value comes
// from the flag (if set) or the parsed env var (if present), aborting on a
// non-bool env value (an empty value is treated as false), exactly as before.
// Returns nil when neither is set, so detection falls back to the inventory.
func secretFilteringFlagOverride(cmd *cobra.Command) *bool {
	// The env var was bound to the flag and validated when the flag set was
	// built, before the command-line value was applied, so a non-bool env value
	// aborts even when --secret-filtering is also passed.
	envVal, envSet := os.LookupEnv(configs.IsSecretFilteringKey)
	if envSet && envVal != "" {
		if _, err := strconv.ParseBool(envVal); err != nil {
			failf("could not parse %q as bool value for $%s", envVal, configs.IsSecretFilteringKey)
		}
	}

	if cmd.Flags().Changed(secretFilteringFlag) {
		v, _ := cmd.Flags().GetBool(secretFilteringFlag)
		return pointers.NewBoolPtr(v)
	}
	if !envSet {
		return nil
	}
	if envVal == "" {
		return pointers.NewBoolPtr(false)
	}
	parsed, _ := strconv.ParseBool(envVal)
	return pointers.NewBoolPtr(parsed)
}

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

// detectPlugin decides plugin dispatch: it only happens when the first
// non-global-flag token is not a known command, so e.g. `bitrise run a:b` stays
// a run invocation rather than being treated as a plugin.
func detectPlugin(root *cobra.Command, rawArgs []string) (string, []string, bool) {
	cmdArgs := stripGlobalFlags(rawArgs)
	if len(cmdArgs) == 0 {
		return "", nil, false
	}
	if isKnownCommand(root, cmdArgs[0]) {
		return "", nil, false
	}
	return plugins.ParseArgs(cmdArgs)
}

// envmanPassthrough reports whether the invocation targets the envman command
// (the first non-global-flag token is "envman") and, if so, returns the args
// that follow it, to be forwarded verbatim.
func envmanPassthrough(rawArgs []string) ([]string, bool) {
	for i, a := range rawArgs {
		if isGlobalFlagArg(a) {
			continue
		}
		if a == envmanCommand.Name() {
			return rawArgs[i+1:], true
		}
		return nil, false
	}
	return nil, false
}

func runEnvman(root *cobra.Command, rawArgs []string, envmanArgs []string) {
	applyGlobalFlagsFromArgs(root, rawArgs)
	if err := before(root, nil); err != nil {
		failf(err.Error())
	}

	logCommandParameters(envmanCommand)

	if err := runCommandWith("envman", envmanArgs); err != nil {
		failf("Command failed, error: %s", err)
	}
}

func runPlugin(root *cobra.Command, rawArgs []string, pluginName string, pluginArgs []string) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())

	applyGlobalFlagsFromArgs(root, rawArgs)
	if err := before(root, nil); err != nil {
		failf(err.Error())
	}

	logPluginCommandParameters(pluginName, pluginArgs)

	plugin, found, err := plugins.LoadPlugin(pluginName)
	if err != nil {
		failf("failed to get plugin (%s), error: %s", pluginName, err)
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
}

func applyGlobalFlagsFromArgs(root *cobra.Command, args []string) {
	for _, name := range []string{DebugModeKey, CIKey, PRKey} {
		for _, a := range args {
			switch {
			case a == "--"+name || a == "-"+name:
				_ = root.PersistentFlags().Set(name, "true")
			case strings.HasPrefix(a, "--"+name+"="):
				_ = root.PersistentFlags().Set(name, strings.TrimPrefix(a, "--"+name+"="))
			case strings.HasPrefix(a, "-"+name+"="):
				_ = root.PersistentFlags().Set(name, strings.TrimPrefix(a, "-"+name+"="))
			}
		}
	}
}

func stripGlobalFlags(args []string) []string {
	out := []string{}
	for _, a := range args {
		if isGlobalFlagArg(a) {
			continue
		}
		out = append(out, a)
	}
	return out
}

func isGlobalFlagArg(a string) bool {
	for _, name := range []string{DebugModeKey, CIKey, PRKey} {
		if a == "--"+name || a == "-"+name ||
			strings.HasPrefix(a, "--"+name+"=") || strings.HasPrefix(a, "-"+name+"=") {
			return true
		}
	}
	return false
}

func isKnownCommand(root *cobra.Command, name string) bool {
	if name == "help" {
		return true
	}
	for _, c := range root.Commands() {
		if c.Name() == name {
			return true
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}

// normalizeLegacyArgs rewrites single-dash long flags (e.g. `-config`) to their
// double-dash form, so both spellings are accepted. Passthrough args are left
// untouched: everything after `--` and everything from the envman command
// onwards (envman forwards its args verbatim and uses single-dash long flags of
// its own).
func normalizeLegacyArgs(args []string, root *cobra.Command) []string {
	known := knownLongFlagNames(root)
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" || a == "envman" {
			out = append(out, args[i:]...)
			break
		}
		out = append(out, normalizeArg(a, known))
	}
	return out
}

func normalizeArg(arg string, known map[string]bool) string {
	if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
		return arg
	}
	name := strings.TrimPrefix(arg, "-")
	if eq := strings.IndexByte(name, '='); eq >= 0 {
		name = name[:eq]
	}
	if len(name) >= 2 && known[name] {
		return "-" + arg
	}
	return arg
}

func knownLongFlagNames(root *cobra.Command) map[string]bool {
	names := map[string]bool{"help": true, "version": true}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		c.PersistentFlags().VisitAll(func(f *pflag.Flag) { names[f.Name] = true })
		c.Flags().VisitAll(func(f *pflag.Flag) { names[f.Name] = true })
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	return names
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	globalTracker.Wait()
	os.Exit(1)
}
