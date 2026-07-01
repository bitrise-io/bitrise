package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareCommand() *cobra.Command {
	shareCommand := &cobra.Command{
		Use:   "share",
		Short: "Publish your step.",
		RunE:  share,
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
	logCommandParameters(cmd)

	if err := tools.StepmanShare(); err != nil {
		failf("Bitrise share failed, error: %s", err)
	}

	return nil
}
