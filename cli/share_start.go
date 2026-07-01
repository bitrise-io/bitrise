package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareStartCommand() *cobra.Command {
	shareStartCommand := &cobra.Command{
		Use:   "start",
		Short: "Preparations for publishing.",
		RunE:  start,
	}

	shareStartCommand.Flags().StringP(CollectionKey, "c", "", "Collection of step.")
	setFlagEnvVar(shareStartCommand.Flags(), CollectionKey, CollectionPathEnvKey)

	return shareStartCommand
}

func start(cmd *cobra.Command, _ []string) error {
	logCommandParameters(cmd)

	// Input validation
	collectionURI, _ := cmd.Flags().GetString(CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if collectionURI == "" {
		failf("No step collection specified")
	}

	if err := tools.StepmanShareStart(collectionURI); err != nil {
		failf("Bitrise share start failed, error: %s", err)
	}

	return nil
}
