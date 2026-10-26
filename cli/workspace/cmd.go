package workspace

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "workspace",
		Short: "List and inspect workspaces.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewListCommand())
	c.AddCommand(NewViewCommand())
	return c
}
