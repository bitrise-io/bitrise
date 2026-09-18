package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmd merges the old `steps` cache commands and the `share` command
// under a single `step` parent.
func NewCmd() *cobra.Command {
	stepCommand := &cobra.Command{
		Use:   "step",
		Short: "Manage steps.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}

	stepCommand.AddCommand(
		NewListCachedStepsCommand(),
		NewPreloadStepsCommand(),
		NewShareCommand(),
		NewSearchCommand(),
		NewInputsCommand(),
	)

	return stepCommand
}

// NewLegacyStepsCommand builds the deprecated top-level `steps` parent, kept as a
// hidden alias for the step cache commands now grouped under `step`.
func NewLegacyStepsCommand() *cobra.Command {
	stepsCommand := &cobra.Command{
		Use:    "steps",
		Short:  "Manage Steps cache.",
		Hidden: true,
		RunE:   cmdutil.RequireKnownSubcommand,
	}

	stepsCommand.AddCommand(
		NewListCachedStepsCommand(),
		NewPreloadStepsCommand(),
	)

	return stepsCommand
}
