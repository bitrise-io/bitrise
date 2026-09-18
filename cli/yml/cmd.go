package yml

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmd returns the `bitrise yml` parent command.
func NewCmd() *cobra.Command {
	ymlCommand := &cobra.Command{
		Use:   "yml",
		Short: "Work with bitrise.yml files.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}

	ymlCommand.AddCommand(
		NewValidateCommand(),
		NewMergeCommand(),
		NewGetCommand(),
		NewUpdateCommand(),
	)

	return ymlCommand
}
