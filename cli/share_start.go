package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Preparations for publishing.",
	RunE:  runStart,
}

var shareStartOpts struct {
	collection string
}

func init() {
	shareStartCmd.Flags().StringVarP(&shareStartOpts.collection, CollectionKey, "c", "", "Collection of step.")
}

func runStart(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)
	return start(cmd, args)
}

func start(_ *cobra.Command, _ []string) error {
	collectionURI := shareStartOpts.collection
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
