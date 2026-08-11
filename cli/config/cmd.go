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
  Per-dir:     .bitrise-cli.yml in the current directory or any ancestor

Recognized keys: %s

api_base_url and web_base_url resolve as: global file > built-in default.
web_base_url is additionally overridable via $%s, which
wins over both. They deliberately ignore the per-directory .bitrise-cli.yml
— each names a host that receives credentials, and a repo you merely clone
and run 'bitrise' inside of must not be able to silently redirect them.

app_id resolves as: --app flag > $BITRISE_APP_ID > per-directory file >
global file. default_workspace_id resolves the same way, via --workspace and
$BITRISE_WORKSPACE_ID. Both honor the per-directory file precisely so a repo
can pin which app and workspace its checkout belongs to; they're identifiers,
not credentials. 'bitrise app create' writes app_id to the global file, and
falls back to default_workspace_id when --workspace is omitted.

'get'/'set'/'unset'/'list' only read and write the global file — per-dir
files must be edited by hand.

To manage your access token, use 'bitrise auth login/logout/status'.`,
			strings.Join(internalconfig.Keys, ", "), cmdutil.EnvWebBaseURL,
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
