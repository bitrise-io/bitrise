package stack

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "stack",
		Short: "List available stacks",
		Long: `List the build stacks available to you.

Running "bitrise stack" with no subcommand lists stacks.`,
		Example: `  bitrise stack list
  bitrise stack list --format json
  bitrise stack list --workspace WORKSPACE_ID`,
		RunE: cmdutil.DelegateToList,
	}
	c.AddCommand(NewListCommand())
	return c
}
