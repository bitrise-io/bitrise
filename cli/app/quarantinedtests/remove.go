package quarantinedtests

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
)

// NewRemoveCommand returns the `app quarantined-tests remove` subcommand.
func NewRemoveCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove ID",
		Short: "Remove a test from quarantine",
		Long: `Remove a test from quarantine, so it is included in test runs again.

Asks for confirmation unless --yes is passed. Without a terminal to ask on,
--yes is required.`,
		Example: `  bitrise app quarantined-tests remove 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id
  bitrise app quarantined-tests remove 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if !yes && !cmdutil.IsTerminal(cmd.InOrStdin()) {
				return errors.New("stdin is not a terminal to confirm on; pass --yes to remove the test without confirmation")
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			appSlug, err := cmdutil.ResolveAndLookupAppSlug(cmd, client)
			if err != nil {
				return err
			}
			svc := internalapp.NewService(client)
			id := args[0]

			if !yes {
				t, err := svc.GetQuarantinedTest(cmd.Context(), appSlug, id)
				if err != nil {
					return fmt.Errorf("getting quarantined test failed: %w", err)
				}
				if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Remove %s from quarantine? [y/N]: ", testPath(t)); err != nil {
					return err
				}
				answer, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "", true)
				if err != nil {
					return err
				}
				if answer != "y" && answer != "Y" && answer != "yes" {
					return errors.New("aborted")
				}
			}

			if err := svc.RemoveQuarantinedTest(cmd.Context(), appSlug, id); err != nil {
				return fmt.Errorf("removing quarantined test failed: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed quarantined test %s\n", id)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "remove without asking for confirmation")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")

	return cmd
}
