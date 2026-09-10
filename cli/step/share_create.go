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

	shareCreateCommand.Flags().String(cmdutil.TagKey, "", "Git (version) tag. (required)")
	shareCreateCommand.Flags().String(cmdutil.GitKey, "", "Git clone url of the step repository. (required)")
	shareCreateCommand.Flags().String(cmdutil.StepIDKey, "", "ID of the step (default: derived from --git's repository name).")
	_ = shareCreateCommand.MarkFlagRequired(cmdutil.TagKey)
	_ = shareCreateCommand.MarkFlagRequired(cmdutil.GitKey)

	return shareCreateCommand
}

func create(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	tag, _ := cmd.Flags().GetString(cmdutil.TagKey)
	gitURI, _ := cmd.Flags().GetString(cmdutil.GitKey)
	stepID, _ := cmd.Flags().GetString(cmdutil.StepIDKey)

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		cmdutil.Failf("Bitrise share create failed, error: %s", err)
	}

	return nil
}
