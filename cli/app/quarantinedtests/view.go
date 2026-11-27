package quarantinedtests

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewViewCommand returns the `app quarantined-tests view` subcommand.
func NewViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view ID",
		Short: "Show a quarantined test",
		Example: `  bitrise app quarantined-tests view 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id
  bitrise app quarantined-tests view 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			appSlug, err := cmdutil.ResolveAndLookupAppSlug(cmd, client)
			if err != nil {
				return err
			}

			t, err := internalapp.NewService(client).GetQuarantinedTest(cmd.Context(), appSlug, args[0])
			if err != nil {
				return fmt.Errorf("getting quarantined test failed: %w", err)
			}

			return output.Render(cmd.OutOrStdout(), output.Format, t, printQuarantinedTestText)
		},
	}

	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
