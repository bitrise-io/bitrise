package session

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

// followRetryInterval is how long --follow waits before retrying while the
// stage hasn't produced any logs yet. A package var so tests can shorten it;
// not exposed as a flag to keep the command surface small.
var followRetryInterval = 3 * time.Second

// snapshotIdleTimeout is how long the default (non-follow) mode waits for new
// content before deciding the replayed log has drained and exiting. The
// backend never signals end-of-log, so this idle gap is how "print what's
// there, then stop" terminates. A package var so tests can shorten it; not a
// flag, to keep the command surface small.
var snapshotIdleTimeout = 2 * time.Second

func newLogsCmd() *cobra.Command {
	var (
		stage  string
		follow bool
		format string
	)

	c := &cobra.Command{
		Use:   "logs SESSION_ID --stage warmup|startup",
		Short: "Print a session's warmup or startup logs",
		Long: `Print the warmup or startup script logs for a session — useful for debugging a
session stuck provisioning or one that came up failed. The log replays from the
start every time you connect.

By default this prints the log captured so far and exits, without waiting for
more output. Pass --follow to keep streaming new output live until you stop it
with Ctrl-C (the backend does not signal end-of-log, so --follow runs until
interrupted).

  --stage    which script's logs to show: warmup or startup (required). warmup
             runs once at session creation; startup runs on every session
             start/restart.
  --follow   keep streaming new output until Ctrl-C, and wait for the stage to
             start if it hasn't produced any logs yet (instead of erroring).

--format is rejected — logs stream as raw text, not a single object. Pipe or
redirect as needed; diagnostics go to stderr so a redirect captures only log
text.`,
		Example: `  bitrise rde session logs SESSION_ID --stage startup
  bitrise rde session logs SESSION_ID --stage warmup
  bitrise rde session logs SESSION_ID --stage startup --follow
  bitrise rde session logs SESSION_ID --stage startup > session.log`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}
			// Reject any machine-readable format before any network round-trip
			// — the feed is plain text, not a single object.
			if output.Format != output.FormatRaw {
				return fmt.Errorf("session logs cannot be combined with --format %s (the feed is plain text, not a single object)", output.Format)
			}
			// Validate --stage up front so a typo errors before the stream
			// header prints (mirrors notifications' --order validation).
			switch stage {
			case internalrde.LogStageWarmup, internalrde.LogStageStartup:
			default:
				return fmt.Errorf("--stage must be %s or %s", internalrde.LogStageWarmup, internalrde.LogStageStartup)
			}

			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)
			sessionID, err := svc.ResolveSessionID(cmd.Context(), workspaceID, args[0])
			if err != nil {
				return err
			}

			// Stop cleanly on Ctrl-C: cancelling the context ends the stream
			// and StreamSessionLogs returns nil, so the command exits 0.
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer cancel()

			out := cmd.OutOrStdout()
			emit := func(chunk string) error {
				_, werr := io.WriteString(out, chunk)
				return werr
			}

			if follow {
				return runFollow(ctx, cmd, svc, workspaceID, sessionID, stage, emit)
			}
			return runSnapshot(ctx, cmd, svc, workspaceID, sessionID, stage, emit)
		},
	}

	c.Flags().StringVar(&stage, "stage", "", "which logs to show: warmup or startup (required)")
	c.Flags().BoolVarP(&follow, "follow", "f", false, "keep streaming until Ctrl-C, waiting for the stage to start if needed")
	_ = c.MarkFlagRequired("stage")
	// logs already owns -f for --follow, so --format registers with no
	// shorthand here.
	c.Flags().StringVar(&format, cmdutil.FormatKey, "", "Output format. Accepted: raw (default), json, yml")

	_ = c.RegisterFlagCompletionFunc("stage", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			internalrde.LogStageWarmup + "\twarmup script (once at creation)",
			internalrde.LogStageStartup + "\tstartup script (every start/restart)",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	return c
}

// runSnapshot prints the stage log captured so far and exits, without waiting
// for new output. The backend never sends EOF, so it relies on an idle gap
// (snapshotIdleTimeout): the replayed backlog arrives in a burst, then the
// stream goes quiet and the command returns. A pre-stream 404 ("logs not
// available yet") is turned into a friendly, actionable message and a non-zero
// exit rather than a raw API error.
func runSnapshot(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, workspaceID, sessionID, stage string, emit func(string) error) error {
	err := svc.StreamSessionLogs(ctx, workspaceID, sessionID, stage, snapshotIdleTimeout, emit)
	if internalrde.IsNotFound(err) {
		cmdutil.SilenceRootErrors(cmd)
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"No %s logs available yet — the session may still be provisioning. Retry shortly, or pass --follow to wait.\n", stage)
		return err
	}
	return err
}

// runFollow streams the stage log live until Ctrl-C. If the stage hasn't
// started producing logs yet (404), it waits followRetryInterval and retries
// — the dashboard does the same. The retry is safe because a 404 only ever
// comes back pre-stream (nothing has been printed yet), so it can't duplicate
// already-streamed output. idleTimeout is 0 here: follow never self-terminates.
func runFollow(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, workspaceID, sessionID, stage string, emit func(string) error) error {
	if !cmdutil.IsQuiet(cmd) {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Streaming %s logs for session %s — Ctrl-C to stop…\n", stage, sessionID)
	}
	for {
		err := svc.StreamSessionLogs(ctx, workspaceID, sessionID, stage, 0, emit)
		if err == nil {
			return nil
		}
		if internalrde.IsNotFound(err) && ctx.Err() == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(followRetryInterval):
				continue
			}
		}
		return err
	}
}
