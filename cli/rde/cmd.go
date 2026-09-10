package rde

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/cli/rde/claude"
	"github.com/bitrise-io/bitrise/v2/cli/rde/machinetype"
	"github.com/bitrise-io/bitrise/v2/cli/rde/savedinput"
	"github.com/bitrise-io/bitrise/v2/cli/rde/session"
	"github.com/bitrise-io/bitrise/v2/cli/rde/stack"
	"github.com/bitrise-io/bitrise/v2/cli/rde/template"
	"github.com/bitrise-io/bitrise/v2/cli/rde/usage"
)

// NewCmd returns the `bitrise rde` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "rde",
		Short: "Manage Bitrise Remote Dev Environments (sessions, templates, …)",
		Long: `Manage Bitrise Remote Dev Environments — sessions, templates, saved inputs,
and the machine catalog (stacks, machine types).

If --workspace isn't resolved from a flag, env var, or configured default,
and you belong to more than one workspace, you're prompted to pick one on a
terminal, or shown a sorted list of workspaces to choose from via --workspace
otherwise.

Saved inputs are user-scoped, though — they do not require --workspace, and
the 'saved-input' subcommand does not accept it.`,
		Example: `  bitrise rde session list --workspace WORKSPACE_ID
  bitrise rde session list --format json
  bitrise rde machine-type list --stack osx-xcode-16.0.x-edge`,
		RunE: cmdutil.RequireKnownSubcommand,
	}

	c.AddCommand(
		claude.NewCmd(),
		stack.NewCmd(),
		machinetype.NewCmd(),
		session.NewCmd(),
		template.NewCmd(),
		savedinput.NewCmd(),
		usage.NewCmd(),
	)
	return c
}
