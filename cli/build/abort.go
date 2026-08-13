package build

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewAbortCommand returns the `build abort` subcommand.
func NewAbortCommand() *cobra.Command {
	var (
		reason              string
		abortWithSuccess    bool
		skipGitStatusReport bool
		skipNotifications   bool
	)

	cmd := &cobra.Command{
		Use:   "abort BUILD_ID",
		Short: "Abort a running or queued build",
		Long: `Abort a running or queued build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID.`,
		Example: `  bitrise build abort abc123 --app my-app-id
  bitrise build abort abc123 --app my-app-id --reason "no longer needed"
  bitrise build abort abc123 --app my-app-id --abort-with-success`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			res, err := internalbuild.NewService(client).Abort(cmd.Context(), internalbuild.AbortRequest{
				AppSlug:             appSlug,
				BuildSlug:           args[0],
				Reason:              reason,
				AbortWithSuccess:    abortWithSuccess,
				SkipGitStatusReport: skipGitStatusReport,
				SkipNotifications:   skipNotifications,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, printAbortText)
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "reason for aborting, recorded in the build log")
	cmd.Flags().BoolVar(&abortWithSuccess, "abort-with-success", false, "mark the aborted build as successful")
	cmd.Flags().BoolVar(&skipGitStatusReport, "skip-git-status-report", false, "don't report the abort to the git provider's status API")
	cmd.Flags().BoolVar(&skipNotifications, "skip-notifications", false, "don't send abort notifications")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
