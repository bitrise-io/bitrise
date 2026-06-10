package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Publish your step.",
	RunE:  runShare,
}

func runShare(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	if err := tools.StepmanShare(); err != nil {
		failf("Bitrise share failed, error: %s", err)
	}

	return nil
}

func init() {
	shareCmd.AddCommand(shareStartCmd, shareCreateCmd, shareAuditCmd, shareFinishCmd)
}
