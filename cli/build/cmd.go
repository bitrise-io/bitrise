// Package build implements the `bitrise build` command group: trigger,
// list, view, log, watch, abort, and yml operations on Bitrise builds.
package build

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// NewCmd returns the `bitrise build` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "build",
		Short: "Trigger, list, and inspect builds.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewListCommand(), NewViewCommand(), NewLogCommand(), NewAbortCommand(), NewYMLCommand())
	return c
}
