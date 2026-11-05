// Package quarantinedtests implements the `bitrise app quarantined-tests`
// command group: list, view, add, update and remove an app's quarantined
// tests.
package quarantinedtests

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
)

// NewCmd returns the `bitrise app quarantined-tests` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "quarantined-tests",
		Short: "List and manage the tests quarantined on an app",
		Long: `List and manage the tests quarantined on an app. Quarantined tests are
excluded from the test runs of the builds the quarantine applies to, so a flaky
test does not fail them.

Changing quarantined tests requires permission to run builds on the app and a
plan that includes test quarantine.`,
		RunE: cmdutil.RequireKnownSubcommand,
	}
	c.AddCommand(NewListCommand(), NewViewCommand(), NewAddCommand(), NewUpdateCommand(), NewRemoveCommand())
	return c
}
