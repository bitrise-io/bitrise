package rde

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/cli/rde/machinetype"
	"github.com/bitrise-io/bitrise/v2/cli/rde/session"
	"github.com/bitrise-io/bitrise/v2/cli/rde/stack"
)

// NewCmd returns the `bitrise rde` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "rde",
		Short: "Manage Bitrise Remote Dev Environments (sessions, templates, …)",
		Long: `Manage Bitrise Remote Dev Environments — sessions, templates, saved inputs,
and the machine catalog (stacks, machine types).

Workspace resolution (highest to lowest precedence):
  --workspace ID            flag on the rde command
  BITRISE_WORKSPACE_ID      environment variable
  default_workspace_id      saved with 'bitrise config set'
  auto-detect               when none of the above is set: your only workspace is used,
                             or you're prompted to pick one interactively`,
		Example: `  bitrise rde session list --workspace WORKSPACE_ID
  bitrise rde session list --format json
  bitrise rde machine-type list --stack osx-xcode-16.0.x-edge`,
		RunE: cmdutil.RequireKnownSubcommand,
	}
	c.PersistentFlags().String(cmdutil.FlagWorkspace, "", "workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)")

	c.AddCommand(
		stack.NewCmd(),
		machinetype.NewCmd(),
		session.NewCmd(),
	)
	return c
}
