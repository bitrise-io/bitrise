// Package savedinput wires `bitrise rde saved-input` subcommands. Saved
// inputs are USER-scoped (not workspace-scoped), so these commands don't
// read --workspace.
package savedinput

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// NewCmd returns the `bitrise rde saved-input` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "saved-input",
		Aliases: []string{"saved-inputs"},
		Short:   "Manage saved inputs (reusable credentials/values)",
		Args:    cobra.NoArgs,
		RunE:    cmdutil.DelegateToList,
	}
	c.AddCommand(
		newListCmd(),
		newViewCmd(),
		newCreateCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
	)
	return c
}
