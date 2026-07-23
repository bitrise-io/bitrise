package auth

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "auth",
		Short: "Manage the Bitrise access token",
		Long: `Manage the Bitrise access token used for API requests.

Both Personal Access Tokens (PAT) and Workspace API Tokens (WAT) work the
same way on the wire — paste either kind here.

Storage:
  YAML file at $XDG_CONFIG_HOME/bitrise/cli/auth.yaml (or
  ~/.config/bitrise/cli/auth.yaml). Written with 0600 permissions, separate
  from preferences in config.yml.

Env override:
  BITRISE_TOKEN takes precedence over the saved token; useful for CI.`,
		Example: `  bitrise auth status
  bitrise auth login
  bitrise auth logout`,
		RunE: cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(
		NewLoginCommand(),
		NewLogoutCommand(),
		NewStatusCommand(),
	)
	return c
}
