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

func newRestoreCmd() *cobra.Command {
	var (
		wait        bool
		waitTimeout time.Duration
		format      string
	)
	c := &cobra.Command{
		Use:   "restore SESSION_ID",
		Short: "Restore a terminated session (re-provisions its VM from the persistent disk)",
		Long: `Restore a terminated session (re-provisions its VM from the persistent disk).

Restore is asynchronous: by default the command returns while the session is
still "starting". Pass --wait to block until the session finishes provisioning
(mirrors 'session create --wait'); the command exits non-zero if the session
ends in any state other than "running". This lets an unattended caller restore
and then immediately use the session without hand-rolling a poll loop.`,
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

			// A terminated session is restored from its persistent disk, so
			// check the disk first: fail fast with a clear reason when it's
			// gone, and warn when it's about to expire (the server still
			// allows the restore, so this is the only chance to surface it).
			sess, err := svc.GetSession(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			switch sess.PersistentDiskStatus {
			case internalrde.DiskStatusUnavailable:
				return fmt.Errorf("session %s cannot be restored: its persistent disk is no longer available", sessionID)
			case internalrde.DiskStatusUnavailableSoon:
				if !cmdutil.IsQuiet(cmd) && output.Format == output.FormatRaw {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
						"Warning: this session's persistent disk will become unavailable soon (within ~1 week) — restore while you still can.\n")
				}
			}

			restored, err := svc.RestoreSession(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}

			if wait {
				waitCtx, cancel := context.WithTimeout(cmd.Context(), waitTimeout)
				defer cancel()
				if !cmdutil.IsQuiet(cmd) && output.Format == output.FormatRaw {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for session %s to become ready (timeout %s)…\n", restored.ID, waitTimeout)
				}
				ready, waitErr := svc.WaitForReady(waitCtx, workspaceID, restored.ID, 0, nil)
				if waitErr != nil {
					return fmt.Errorf("waiting for session: %w", waitErr)
				}
				restored = ready
				if ready.Status != "running" {
					if renderErr := output.Render(cmd.OutOrStdout(), output.Format, restored, renderSessionDetail); renderErr != nil {
						return renderErr
					}
					cmdutil.SilenceRootErrors(cmd)
					return fmt.Errorf("session ended provisioning with status %q (expected running)", ready.Status)
				}
			}

			return output.Render(cmd.OutOrStdout(), output.Format, restored, renderSessionDetail)
		},
	}
	c.Flags().BoolVar(&wait, "wait", false, "wait until the session leaves provisioning (running, failed, …) before returning; exits 1 if the final status isn't running")
	c.Flags().DurationVar(&waitTimeout, "wait-timeout", 10*time.Minute, "max time to wait when --wait is set (Go duration syntax: 30s, 5m, 1h)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}
