package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareFinishCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "finish",
		Short: "Finish up.",
		RunE:  finish,
	}
}

func finish(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	if err := tools.StepmanShareFinish(); err != nil {
		cmdutil.Failf("Bitrise share finish failed, error: %s", err)
	}

	return nil
}
