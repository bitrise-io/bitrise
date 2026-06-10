package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Validates the step collection.",
	RunE:  runShareAudit,
}

func runShareAudit(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	if err := tools.StepmanShareAudit(); err != nil {
		failf("Bitrise share audit failed, error: %s", err)
	}

	return nil
}
