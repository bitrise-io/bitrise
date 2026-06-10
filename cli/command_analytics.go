package cli

import (
	"fmt"
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

	parts := strings.Split(cmd.CommandPath(), " ")
	if len(parts) > 1 {
		commandName = parts[1]
		if len(parts) > 2 {
			subcommandName = parts[2]
		}
	}

	flags := collectFlags(cmd)
	sendCommandInfo(commandName, subcommandName, flags)
}

func collectFlags(cmd *cobra.Command) []string {
	var flags []string

	cmd.Root().PersistentFlags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})

	return flags
}

func sendCommandInfo(command, subcommand string, flags []string) {
	globalTracker.SendCommandInfo(command, subcommand, flags)
}
