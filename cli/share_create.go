package cli

import (
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

var shareCreateCommand = &cobra.Command{
	Use:   "create",
	Short: "Create your change - add it to your own copy of the collection.",
	RunE:  create,
}

func init() {
	shareCreateCommand.Flags().String(TagKey, "", "Git (version) tag.")
	shareCreateCommand.Flags().String(GitKey, "", "Git clone url of the step repository.")
	shareCreateCommand.Flags().String(StepIDKey, "", "ID of the step.")
}

func create(cmd *cobra.Command, _ []string) error {
	logCommandParameters(cmd)

	// Input validation
	tag, _ := cmd.Flags().GetString(TagKey)
	if tag == "" {
		failf("No step tag specified")
	}

	gitURI, _ := cmd.Flags().GetString(GitKey)
	if gitURI == "" {
		failf("No step url specified")
	}

	stepID, _ := cmd.Flags().GetString(StepIDKey)

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		failf("Bitrise share create failed, error: %s", err)
	}

	return nil
}
