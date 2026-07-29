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
		Long: `Work with bitrise.yml files.

Running "bitrise yml" with no subcommand defaults to "yml get".`,
		RunE: cmdutil.DelegateTo("get"),
	}

	ymlCommand.AddCommand(
		NewValidateCommand(),
		NewMergeCommand(),
		NewGetCommand(),
		NewUpdateCommand(),
	)

	return ymlCommand
}
