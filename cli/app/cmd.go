// Package app implements the `bitrise app` command group: list, view, and
// create apps on Bitrise, and manage their quarantined tests.
package app

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/app/quarantinedtests"
	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
)

// NewCmd returns the `bitrise app` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "app",
		Short: "List, inspect, and manage apps.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewListCommand(), NewViewCommand(), NewCreateCommand(), quarantinedtests.NewCmd())
	return c
}
