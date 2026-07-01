package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/tools"
	"github.com/spf13/cobra"
)

func newShareCreateCommand() *cobra.Command {
	shareCreateCommand := &cobra.Command{
		Use:   "create",
		Short: "Create your change - add it to your own copy of the collection.",
		RunE:  create,
	}

	shareCreateCommand.Flags().String(cmdutil.TagKey, "", "Git (version) tag.")
	shareCreateCommand.Flags().String(cmdutil.GitKey, "", "Git clone url of the step repository.")
	shareCreateCommand.Flags().String(cmdutil.StepIDKey, "", "ID of the step.")

	return shareCreateCommand
}

func create(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	// Input validation
	tag, _ := cmd.Flags().GetString(cmdutil.TagKey)
	if tag == "" {
		cmdutil.Failf("No step tag specified")
	}

	gitURI, _ := cmd.Flags().GetString(cmdutil.GitKey)
	if gitURI == "" {
		cmdutil.Failf("No step url specified")
	}

	stepID, _ := cmd.Flags().GetString(cmdutil.StepIDKey)

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		cmdutil.Failf("Bitrise share create failed, error: %s", err)
	}

	return nil
}
