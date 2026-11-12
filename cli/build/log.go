package build

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
)

// NewLogCommand returns the `build log` subcommand.
func NewLogCommand() *cobra.Command {
	var (
		wait     bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "log BUILD_ID",
		Short: "Print the build log",
		Long: `Print the log output for a single build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

--wait waits for the build to finish before printing the log — useful when
the build is still in-progress. Ctrl-C detaches without affecting the
running build, and exits non-zero because no log was printed.

Output is always raw text — this command has no --format flag.`,
		Example: `  bitrise build log abc123 --app my-app-id
  bitrise build log abc123 --app my-app-id --wait
  bitrise build log abc123 --app my-app-id --wait --interval 10s
  bitrise build log abc123 --app my-app-id > build.log`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

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

			if wait {
				b, err := svc.View(cmd.Context(), appSlug, buildSlug)
				if err != nil {
					return err
				}
				if b.Status == "in-progress" {
					if _, err := waitForBuild(cmd, svc, b, interval, "build log --wait "+buildSlug); err != nil {
						return err
					}
				}
			}

			return svc.Log(cmd.Context(), appSlug, buildSlug, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the build to finish before printing the log")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval when --wait is active")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")

	return cmd
}
