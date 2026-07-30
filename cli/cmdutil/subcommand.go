package cmdutil

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/spf13/cobra"
)

// RequireKnownSubcommand is the RunE for parent commands that only dispatch to
// subcommands (they have no action of their own): show help when invoked bare,
// and error on an unknown subcommand instead of silently succeeding.
func RequireKnownSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
}

// AsHidden marks cmd as hidden and returns it.
func AsHidden(cmd *cobra.Command) *cobra.Command {
	cmd.Hidden = true
	return cmd
}

// ShowSubcommandHelp prints cmd's help output.
func ShowSubcommandHelp(cmd *cobra.Command) {
	if err := cmd.Help(); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}
