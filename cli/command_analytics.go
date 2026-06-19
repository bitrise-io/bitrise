package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var globalTracker analytics.Tracker

func logPluginCommandParameters(name string, _ []string) {
	// Plugin command parameters are routed into the function but are not processed yet because it is complex to correctly
	// parse the arguments without knowing the structure. If we notice that our users do use plugins, then we can add
	// plugin specific argument parsers.
	sendCommandInfo(fmt.Sprintf(":%s", name), "", []string{})
}

func logCommandParameters(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	commandName := "unknown"
	subcommandName := ""

	// CommandPath is e.g. "bitrise tools install"; drop the leading program name.
	if names := strings.Fields(cmd.CommandPath()); len(names) > 1 {
		commandName = names[1]
		if len(names) > 2 {
			subcommandName = names[2]
		}
	}

	sendCommandInfo(commandName, subcommandName, collectFlags(cmd))
}

func collectFlags(cmd *cobra.Command) []string {
	var flags []string

	persistent := cmd.Root().PersistentFlags()
	for _, name := range []string{DebugModeKey, CIKey, PRKey} {
		if f := persistent.Lookup(name); f != nil && flagIsSet(f) {
			flags = append(flags, name)
		}
	}

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if flagIsSet(f) {
			flags = append(flags, f.Name)
		}
	})

	return flags
}

func flagIsSet(f *pflag.Flag) bool {
	if f.Changed {
		return true
	}
	for _, envKey := range f.Annotations[envVarAnnotation] {
		if _, present := os.LookupEnv(envKey); present {
			return true
		}
	}
	return false
}

func sendCommandInfo(command, subcommand string, flags []string) {
	globalTracker.SendCommandInfo(command, subcommand, flags)
}

// envVarAnnotation records the environment variable a flag is bound to, so that
// analytics reports the flag as set when its value comes from the environment.
const envVarAnnotation = "bitrise_env_var"

// markEnvVar binds a flag to an environment variable for analytics reporting.
func markEnvVar(fs *pflag.FlagSet, name, envKey string) {
	_ = fs.SetAnnotation(name, envVarAnnotation, []string{envKey})
}
