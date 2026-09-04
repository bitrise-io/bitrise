package template

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// NewCmd returns the `bitrise rde template` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "template",
		Short: "List and inspect RDE templates",
		Long: `List and inspect RDE templates.

Commands that take a TEMPLATE_ID also accept a template name — it's resolved
to an ID for you. Names aren't unique, so if more than one template shares the
name the command errors and lists the candidate IDs to pick from.`,
		Args: cobra.NoArgs,
		RunE: cmdutil.DelegateToList,
	}
	c.AddCommand(
		newListCmd(),
		newViewCmd(),
		newCreateCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
	)
	return c
}
