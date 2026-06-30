package cli

import "github.com/spf13/cobra"

func newLocalCommand() *cobra.Command {
	localCommand := &cobra.Command{
		Use:   "local",
		Short: "Run and manage Bitrise workflows on the local host.",
		RunE:  requireKnownSubcommand,
	}

	localCommand.AddCommand(
		newRunCommand(),
		newInitCommand(),
		newSetupCommand(),
		newToolsCommand(),
		newWorkflowListCommand(),
	)

	return localCommand
}
