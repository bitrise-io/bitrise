package user

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// NewCmd returns the `bitrise user` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "user",
		Short: "Create and manage your Bitrise account.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewCreateCommand(), NewMeCommand())
	return c
}
