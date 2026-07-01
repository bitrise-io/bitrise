package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareAuditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "audit",
		Short: "Validates the step collection.",
		RunE:  shareAudit,
	}
}

func shareAudit(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	if err := tools.StepmanShareAudit(); err != nil {
		cmdutil.Failf("Bitrise share audit failed, error: %s", err)
	}

	return nil
}
