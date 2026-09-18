package local

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmd returns the `bitrise local` parent command.
func NewCmd() *cobra.Command {
	localCommand := &cobra.Command{
		Use:   "local",
		Short: "Run and manage Bitrise workflows on the local host.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}

	localCommand.AddCommand(
		NewRunCommand(),
		NewInitCommand(),
		NewSetupCommand(),
		NewToolsCommand(),
		NewWorkflowListCommand(),
		// Deprecated, kept for backward compatibility but hidden (see trigger.go).
		cmdutil.AsHidden(NewTriggerCommand()),
	)

	return localCommand
}
