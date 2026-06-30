package cli

import "github.com/spf13/cobra"

// newStepCommand merges the old `steps` cache commands and the `share` command
// under a single `step` parent.
func newStepCommand() *cobra.Command {
	stepCommand := &cobra.Command{
		Use:   "step",
		Short: "Manage steps.",
		RunE:  requireKnownSubcommand,
	}

	stepCommand.AddCommand(
		newListCachedStepsCommand(),
		newPreloadStepsCommand(),
		newShareCommand(),
	)

	return stepCommand
}

// newLegacyStepsCommand builds the deprecated top-level `steps` parent, kept as a
// hidden alias for the step cache commands now grouped under `step`.
func newLegacyStepsCommand() *cobra.Command {
	stepsCommand := &cobra.Command{
		Use:    "steps",
		Short:  "Manage Steps cache.",
		Hidden: true,
		RunE:   requireKnownSubcommand,
	}

	stepsCommand.AddCommand(
		newListCachedStepsCommand(),
		newPreloadStepsCommand(),
	)

	return stepsCommand
}
