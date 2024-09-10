package cli

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/urfave/cli"
)

var globalTracker analytics.Tracker

func logPluginCommandParameters(name string, _ []string) {
	// Plugin command parameters are routed into the function but are not processed yet because it is complex to correctly
	// parse the arguments without knowing the structure. If we notice that our users do use plugins, then we can add
	// plugin specific argument parsers.
	sendCommandInfo(fmt.Sprintf(":%s", name), "", []string{})
}

func logSubcommandParameters(c *cli.Context) {
	if c == nil {
		return
	}

	commandName := "unknown"
	subcommandName := "unknown"

	if names := strings.Split(c.Command.FullName(), " "); len(names) == 2 {
		commandName = names[0]
		subcommandName = names[1]
	}

	flags := collectFlags(c)

	sendCommandInfo(commandName, subcommandName, flags)
}

func logCommandParameters(c *cli.Context) {
	if c == nil {
		return
	}

	flags := collectFlags(c)

	sendCommandInfo(c.Command.Name, "", flags)
}

func collectFlags(c *cli.Context) []string {
	var flags []string

	for _, flag := range c.GlobalFlagNames() {
		if isSet := c.GlobalIsSet(flag); isSet {
			flags = append(flags, flag)
		}
	}

	for _, flag := range c.FlagNames() {
		if isSet := c.IsSet(flag); isSet {
			flags = append(flags, flag)
		}
	}

	return flags
}

func sendCommandInfo(command, subcommand string, flags []string) {
	globalTracker.SendCommandInfo(command, subcommand, flags)
}
