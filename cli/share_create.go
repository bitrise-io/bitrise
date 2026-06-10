package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new step version.",
	RunE:  runCreate,
}

var shareCreateOpts struct {
	tag    string
	git    string
	stepID string
}

func init() {
	shareCreateCmd.Flags().StringVar(&shareCreateOpts.tag, TagKey, "", "Step tag.")
	shareCreateCmd.Flags().StringVar(&shareCreateOpts.git, GitKey, "", "Step git URI.")
	shareCreateCmd.Flags().StringVar(&shareCreateOpts.stepID, StepIDKey, "", "Step ID.")
}

func runCreate(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)
	return create(cmd, args)
}

func create(_ *cobra.Command, _ []string) error {
	tag := shareCreateOpts.tag
	if tag == "" {
		failf("No step tag specified")
	}

	gitURI := shareCreateOpts.git
	if gitURI == "" {
		failf("No step url specified")
	}

	stepID := shareCreateOpts.stepID

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		failf("Bitrise share create failed, error: %s", err)
	}

	return nil
}
