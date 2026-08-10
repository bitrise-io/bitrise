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
		Short: "Manage CLI configuration (defaults persisted to a YAML file).",
		Long: fmt.Sprintf(`Manage persistent CLI configuration.

Storage:
  Global file: ~/.config/bitrise/cli/config.yml
               (honors $XDG_CONFIG_HOME instead of ~/.config)

Precedence, highest to lowest: global file > built-in default.
web_base_url is additionally overridable via $%s, which takes precedence
over the global file.

Recognized keys: %s

These two keys deliberately ignore the per-directory .bitrise-cli.yml file
(unlike every other setting) and the legacy ~/.bitrise/config.json — both
carry credentials to whatever host they name, and a repo you merely clone
and run 'bitrise' inside of must not be able to silently redirect them.
'get'/'set'/'unset'/'list' only read and write the global file.

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
