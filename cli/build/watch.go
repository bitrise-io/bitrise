package build

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewWatchCommand returns the `build watch` subcommand.
func NewWatchCommand() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "watch BUILD_ID",
		Short: "Stream logs for a running build",
		Long: `Wait for the build to finish, then exit with a status reflecting the
build outcome (0 = success, non-zero = failed or aborted).

With --format json/yml, or when stdout isn't a terminal, build logs stream to
it as plain text (to stderr instead, in --format json/yml, so stdout stays
pipeable and carries only the final build record). On a terminal with the
default raw format, an interactive status display is shown instead of raw
log lines.

Ctrl-C detaches the CLI without affecting the running build, and exits
non-zero.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").`,
		Example: `  bitrise build watch abc123 --app my-app-id
  bitrise build watch abc123 --app my-app-id --interval 5s
  bitrise build watch abc123 --app my-app-id --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			buildSlug := args[0]

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			appSlug, err := cmdutil.ResolveAndLookupAppSlug(cmd, client)
			if err != nil {
				return err
			}
			svc := internalbuild.NewService(client)

			b, err := svc.View(cmd.Context(), appSlug, buildSlug)
			if err != nil {
				return err
			}

			return runWatch(cmd, svc, b, interval)
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "log polling interval")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
