package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var globalTracker analytics.Tracker

// SetTracker sets the package-level tracker used by LogCommandParameters,
// LogPluginCommandParameters and Failf.
func SetTracker(t analytics.Tracker) {
	globalTracker = t
}

// Tracker returns the package-level tracker set via SetTracker.
func Tracker() analytics.Tracker {
	return globalTracker
}

// LogPluginCommandParameters ...
func LogPluginCommandParameters(name string, _ []string) {
	// Plugin command parameters are routed into the function but are not processed yet because it is complex to correctly
	// parse the arguments without knowing the structure. If we notice that our users do use plugins, then we can add
	// plugin specific argument parsers.
	SendCommandInfo(fmt.Sprintf(":%s", name), "", []string{})
}

// LogCommandParameters ...
func LogCommandParameters(cmd *cobra.Command) {
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

	SendCommandInfo(commandName, subcommandName, collectFlags(cmd))
}

func collectFlags(cmd *cobra.Command) []string {
	var flags []string

	persistent := cmd.Root().PersistentFlags()
	for _, name := range GlobalFlagNames {
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
	for _, envKey := range f.Annotations[EnvVarAnnotation] {
		if _, present := os.LookupEnv(envKey); present {
			return true
		}
	}
	return false
}

// SendCommandInfo ...
func SendCommandInfo(command, subcommand string, flags []string) {
	globalTracker.SendCommandInfo(command, subcommand, flags)
}

// EnvVarAnnotation records the environment variable a flag is bound to. The
// binding drives both analytics (a flag is reported as set when its value comes
// from the env) and the flag/env mode resolution (see ResolveBoolFlagOrEnv).
const EnvVarAnnotation = "bitrise_env_var"

// SetFlagEnvVar binds a flag to an environment variable.
func SetFlagEnvVar(fs *pflag.FlagSet, name, envKey string) {
	_ = fs.SetAnnotation(name, EnvVarAnnotation, []string{envKey})
}
