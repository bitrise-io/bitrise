package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareFinishCommand = &cobra.Command{
	Use:   "finish",
	Short: "Finish up.",
	RunE:  finish,
}

func finish(cmd *cobra.Command, _ []string) error {
	logCommandParameters(cmd)

	if err := tools.StepmanShareFinish(); err != nil {
		failf("Bitrise share finish failed, error: %s", err)
	}

	return nil
}
