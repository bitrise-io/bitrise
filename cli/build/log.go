package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewLogCommand returns the `build log` subcommand.
func NewLogCommand() *cobra.Command {
	var (
		wait       bool
		interval   time.Duration
		stepStatus string
		stepIDs    []string
		tail       int
	)

	cmd := &cobra.Command{
		Use:   "log BUILD_ID",
		Short: "Print the build log",
		Long: `Print the log output for a single build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

--step-status and --step print the logs of the matching steps only, each
under a header with its workflow, status and duration. --step takes the step
IDs "bitrise build steps" lists. Step logs are available once the build has
finished and its log is archived, a little after; until then only the steps
are listed, without logs.

--tail N keeps the last N lines of the log, or of each step's log.

--wait waits for the build to finish before printing the log — useful when
the build is still in-progress. With --step-status or --step it also waits
for the log to be archived. Ctrl-C detaches without affecting the running build, and
exits non-zero because no log was printed.`,
		Example: `  bitrise build log abc123 --app my-app-id
  bitrise build log abc123 --app my-app-id --tail 100
  bitrise build log abc123 --app my-app-id --step-status failed
  bitrise build log abc123 --app my-app-id --step 3f2a1c9e-... --format json
  bitrise build log abc123 --app my-app-id --wait --step-status failed --tail 50
  bitrise build log abc123 --app my-app-id > build.log`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			buildSlug := args[0]
			byStep := stepStatus != "" || len(stepIDs) > 0

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
				resumeCmd := "build log --wait " + buildSlug
				if b.Status == "in-progress" {
					if _, err := waitForBuild(cmd, svc, b, interval, resumeCmd); err != nil {
						return err
					}
				}
				if byStep {
					if err := waitForLogArchive(cmd, svc, appSlug, buildSlug, interval, resumeCmd); err != nil {
						return err
					}
				}
			}

			if byStep {
				res, err := svc.StepLogs(cmd.Context(), appSlug, buildSlug, stepStatus, stepIDs, tail)
				if err != nil {
					return err
				}
				if err := output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.StepLogsResult) error {
					return printStepLogsText(w, res, stepStatus)
				}); err != nil {
					return err
				}
				if output.Format == output.FormatRaw && !res.Complete {
					_, err := fmt.Fprintln(cmd.ErrOrStderr(), "The step logs are not available yet: the build is still running, or has just finished and its log is being archived. Only the steps that have started so far are listed. Pass --wait to wait for the logs.")
					return err
				}
				return nil
			}

			if output.Format == output.FormatRaw && tail == 0 {
				return svc.Log(cmd.Context(), appSlug, buildSlug, cmd.OutOrStdout())
			}
			res, err := svc.LogText(cmd.Context(), appSlug, buildSlug, tail)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.LogResult) error {
				_, err := io.WriteString(w, res.Log)
				return err
			})
		},
	}

	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the build to finish before printing the log")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval when --wait is active")
	cmd.Flags().StringVar(&stepStatus, "step-status", "", fmt.Sprintf("only the logs of steps with this status: %s. failed matches every status that fails the build: %s",
		strings.Join(internalbuild.StepStatuses, ", "), strings.Join(internalbuild.FailedStepStatuses, ", ")))
	cmd.Flags().StringSliceVar(&stepIDs, "step", nil, "only the log of the step with this ID, as listed by \"bitrise build steps\" (repeatable)")
	cmd.Flags().IntVar(&tail, "tail", 0, "only the last N lines of the log, or of each step's log")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("step-status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return internalbuild.StepStatuses, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func waitForLogArchive(cmd *cobra.Command, svc *internalbuild.Service, appSlug, buildSlug string, interval time.Duration, resumeCmd string) error {
	// Go's default SIGINT handling would kill the process before the detach
	// path runs.
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	err := svc.WaitForLogArchive(ctx, appSlug, buildSlug, interval)
	if errors.Is(err, context.Canceled) {
		return detach(cmd.ErrOrStderr(), resumeCmd, appSlug)
	}
	return err
}

// printStepLogsText prints each step's log under a header naming the step,
// its workflow, status and duration.
func printStepLogsText(w io.Writer, res internalbuild.StepLogsResult, status string) error {
	if len(res.Items) == 0 {
		_, err := fmt.Fprintln(w, noStepsMessage(status))
		return err
	}
	s := style.New(w)
	var b strings.Builder
	for i, st := range res.Items {
		if i > 0 && res.Items[i-1].Log != "" {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s  %s", s.Bold.Render(st.Workflow+" › "+stepName(st.Step)), s.StepStatus(st.Status).Render(st.Status))
		if d := stepDuration(st.Step); d != "" {
			fmt.Fprintf(&b, "  %s", s.Dim.Render(d))
		}
		b.WriteString("\n")
		b.WriteString(st.Log)
		if st.Log != "" && !strings.HasSuffix(st.Log, "\n") {
			b.WriteString("\n")
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
