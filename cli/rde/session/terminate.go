package session

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newTerminateCmd() *cobra.Command {
	var (
		wait        bool
		waitTimeout time.Duration
		format      string
	)
	c := &cobra.Command{
		Use:   "terminate SESSION_ID",
		Short: "Terminate a running session but keep it for a later restore",
		Long: `Terminate a running session but keep it for a later restore.

The VM is stopped and its disk is preserved: the session stays in the list as
"terminated" and can be brought back with 'session restore'. Use this when you
intend to come back to this exact session (e.g. to keep uncommitted work or an
expensive warm state). A terminated session keeps using disk space until it is
deleted.

If you are simply done with the session, use 'session delete' instead — it
works on running sessions directly and frees the disk; no terminate needed.

Terminate is asynchronous: by default the command returns while the session
is still "terminating". Pass --wait to block until the session settles into a
terminal state ("terminated" or "failed"), e.g. before restoring it or when
you want the disk snapshot to be complete before moving on.`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
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
			sess, err := svc.TerminateSession(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}

			if wait {
				waitCtx, cancel := context.WithTimeout(cmd.Context(), waitTimeout)
				defer cancel()
				if !cmdutil.IsQuiet(cmd) && output.Format == output.FormatRaw {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for session %s to terminate (timeout %s)…\n", sess.ID, waitTimeout)
				}
				settled, waitErr := svc.WaitForTerminated(waitCtx, workspaceID, sess.ID, 0)
				if waitErr != nil {
					return fmt.Errorf("waiting for session to terminate: %w", waitErr)
				}
				sess = settled
			}

			return output.Render(cmd.OutOrStdout(), output.Format, sess, renderSessionDetail)
		},
	}
	c.Flags().BoolVar(&wait, "wait", false, "block until the session settles into a terminal state (terminated/failed) before returning")
	c.Flags().DurationVar(&waitTimeout, "wait-timeout", 10*time.Minute, "max time to wait when --wait is set (Go duration syntax: 30s, 5m, 1h)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}
