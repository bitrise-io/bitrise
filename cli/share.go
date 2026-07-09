package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareCommand = &cobra.Command{
	Use:   "share",
	Short: "Publish your step.",
	RunE:  share,
}

func init() {
	shareCommand.AddCommand(
		shareStartCommand,
		shareCreateCommand,
		shareAuditCommand,
		shareFinishCommand,
	)
}

func share(cmd *cobra.Command, _ []string) error {
	logCommandParameters(cmd)

	if err := tools.StepmanShare(); err != nil {
		failf("Bitrise share failed, error: %s", err)
	}

	return nil
}
