package user

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// NewCmd returns the `bitrise user` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "user",
		Short: "Create and manage your Bitrise account",
		Long: `Manage your own Bitrise account from the CLI.

Supports account creation and viewing the currently
authenticated user. After running "user create" you must click the link
emailed to you, then run "bitrise auth login --email <addr>" to mint and
store an access token.`,
		Example: `  bitrise user me
  bitrise user me --format json
  bitrise user create --email alice@example.com --username alice --first-name Alice --last-name L`,
		RunE: cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewCreateCommand(), NewMeCommand())
	return c
}
