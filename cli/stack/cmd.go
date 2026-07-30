package stack

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "stack",
		Short: "Manage build stacks.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewListCommand())
	return c
}
