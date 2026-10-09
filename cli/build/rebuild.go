package build

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewRebuildCommand returns the `build rebuild` subcommand.
func NewRebuildCommand() *cobra.Command {
	var (
		remoteAccess bool
		wait         bool
		watch        bool
		interval     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "rebuild BUILD_ID",
		Short: "Start a new build with the parameters of a finished build",
		Long: `Start a new build of a finished build with the same parameters and workflow.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID.

--wait and --watch behave as in "bitrise build trigger".`,
		Example: `  bitrise build rebuild abc123 --app my-app-id
  bitrise build rebuild abc123 --app my-app-id --watch
  bitrise build rebuild abc123 --app my-app-id --remote-access
  bitrise build rebuild abc123 --app my-app-id --wait --format json`,
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

			svc := internalbuild.NewService(client)
			b, err := svc.Rebuild(cmd.Context(), internalbuild.RebuildRequest{
				AppSlug:      appSlug,
				BuildSlug:    args[0],
				RemoteAccess: remoteAccess,
			})
			if err != nil {
				return err
			}

			return runAfterTrigger(cmd, svc, b, wait, watch, interval)
		},
	}

	cmd.Flags().BoolVar(&remoteAccess, "remote-access", false, "start the new build with remote access enabled (needs remote access permission on the app)")
	cmd.Flags().BoolVar(&wait, "wait", false, "block until the build finishes without streaming logs (exit code reflects build outcome)")
	cmd.Flags().BoolVar(&watch, "watch", false, "wait for the build to finish, showing progress (exit code reflects build outcome)")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval when --wait or --watch is active")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	cmd.MarkFlagsMutuallyExclusive("wait", "watch")

	return cmd
}
