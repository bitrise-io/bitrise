package build

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewWatchCommand returns the `build watch` subcommand.
func NewWatchCommand() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "watch BUILD_ID",
		Short: "Stream logs for a running build",
		Long: `Stream build logs until the build finishes, then exit with a status
reflecting the build outcome (0 = success, non-zero = failed or aborted).

Ctrl-C detaches the CLI without affecting the running build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID.

In --format json/yml, logs stream to stderr and the final build record is
written to stdout, so this stays pipeable.`,
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

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}
			buildSlug := args[0]

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			svc := internalbuild.NewService(client)

			b, err := svc.View(cmd.Context(), appSlug, buildSlug)
			if err != nil {
				return err
			}

			// In structured output modes, logs go to stderr so stdout carries
			// only the final build record and stays pipeable.
			logWriter := io.Writer(cmd.OutOrStdout())
			if output.Format == output.FormatJSON || output.Format == output.FormatYML {
				logWriter = cmd.ErrOrStderr()
			}
			return runWatch(cmd, svc, b, interval, logWriter, output.Format)
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "log polling interval")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
