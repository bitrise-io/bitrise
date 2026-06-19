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
	"github.com/spf13/pflag"
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

// commandTokenIndex returns the index of the first argument that is not a global
// flag — the command/plugin/positional token. Global flags before this boundary
// configure bitrise; everything from it onward belongs to the command (and, for
// plugins and envman, is forwarded verbatim), so it must not be scanned for or
// stripped of global flags.
func commandTokenIndex(args []string) int {
	for i, a := range args {
		if !isGlobalFlagArg(a) {
			return i
		}
	}
	return len(args)
}

// applyGlobalFlagsFromArgs sets the global flags on the plugin/envman dispatch
// paths, where cobra does not parse them. Only the leading args (before the
// command token) are bitrise globals; anything after belongs to the passthrough.
func applyGlobalFlagsFromArgs(root *cobra.Command, args []string) {
	for _, a := range args[:commandTokenIndex(args)] {
		for _, name := range []string{DebugModeKey, CIKey, PRKey} {
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
