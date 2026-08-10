package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

// NewCmd returns the `config` parent command and its subcommands.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration (defaults persisted to a YAML file)",
		Long: fmt.Sprintf(`Manage persistent CLI configuration.

Storage:
  Global file: ~/.config/bitrise/cli/config.yml
               (honors $XDG_CONFIG_HOME instead of ~/.config)
  Per-dir:     .bitrise-cli.yml in the current directory or any ancestor

Precedence, highest to lowest: per-dir file > global file > built-in default.
web_base_url is additionally overridable via $%s, which takes precedence
over all of the above.

Recognized keys: %s

'get'/'set'/'unset'/'list' only read and write the global file — per-dir
files must be edited by hand.

To manage your access token, use 'bitrise auth login/logout/status'.`,
			cmdutil.EnvWebBaseURL, strings.Join(internalconfig.Keys, ", "),
		),
		RunE: cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(
		NewPathCommand(),
		NewListCommand(),
		NewGetCommand(),
		NewSetCommand(),
		NewUnsetCommand(),
	)
	return c
}
