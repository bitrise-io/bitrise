package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internaluser "github.com/bitrise-io/bitrise/v2/internal/user"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewMeCommand returns the `user me` subcommand.
func NewMeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Show the currently authenticated user",
		Long: `Show the profile of the user whose token is in use.

The token comes from the BITRISE_TOKEN environment variable, or from auth.yaml
as written by 'bitrise auth login' — run 'bitrise auth status' to confirm which
source is active.`,
		Example: `  bitrise user me
  bitrise user me --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			profile, err := internaluser.NewProfileService(client).Me(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetching user profile failed: %w", err)
			}

			if output.Format == output.FormatRaw {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "Username: %s\nEmail:    %s\n", profile.Username, profile.Email)
				return err
			}
			return output.Print(profile, output.Format)
		},
	}

	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}
