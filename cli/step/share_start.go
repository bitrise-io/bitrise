package step

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareStartCommand() *cobra.Command {
	shareStartCommand := &cobra.Command{
		Use:   "start",
		Short: "Preparations for publishing.",
		RunE:  start,
	}

	shareStartCommand.Flags().StringP(cmdutil.CollectionKey, "c", "", "Collection of step.")
	cmdutil.SetFlagEnvVar(shareStartCommand.Flags(), cmdutil.CollectionKey, cmdutil.CollectionPathEnvKey)

	return shareStartCommand
}

func start(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	// Input validation
	collectionURI, _ := cmd.Flags().GetString(cmdutil.CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(cmdutil.CollectionPathEnvKey)
	}
	if collectionURI == "" {
		cmdutil.Failf("No step collection specified")
	}

	if err := tools.StepmanShareStart(collectionURI); err != nil {
		cmdutil.Failf("Bitrise share start failed, error: %s", err)
	}

	return nil
}
