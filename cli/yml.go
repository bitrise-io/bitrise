package cli

import "github.com/spf13/cobra"

func newYMLCommand() *cobra.Command {
	ymlCommand := &cobra.Command{
		Use:   "yml",
		Short: "Work with bitrise.yml files.",
		RunE:  requireKnownSubcommand,
	}

	ymlCommand.AddCommand(
		newValidateCommand(),
		newMergeCommand(),
	)

	return ymlCommand
}
