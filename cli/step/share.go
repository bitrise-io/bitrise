package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

// NewShareCommand ...
func NewShareCommand() *cobra.Command {
	shareCommand := &cobra.Command{
		Use:   "share",
		Short: "Publish your step.",
		Long: `Publish a step to a step library, guiding you through the full flow.

Run bare (no subcommand), this runs the whole interactive publishing wizard.
The subcommands (start, create, audit, finish) let you drive the same steps
individually instead — useful for scripting, or resuming after one step
failed.`,
		RunE: share,
	}

	shareCommand.AddCommand(
		newShareStartCommand(),
		newShareCreateCommand(),
		newShareAuditCommand(),
		newShareFinishCommand(),
	)

	return shareCommand
}

func share(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	if err := tools.StepmanShare(); err != nil {
		cmdutil.Failf("Bitrise share failed, error: %s", err)
	}

	return nil
}
