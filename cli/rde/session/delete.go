package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete SESSION_ID",
		Short: "Permanently delete a session in any state (running sessions are stopped and discarded)",
		Long: `Permanently delete a session in any state.

This is the way to get rid of a session you are done with. It works on running,
starting, terminating, terminated and failed sessions alike — no need to
'terminate' first. The session disappears immediately and cannot be restored.

If the VM is still running, the backend stops it and then discards it together
with its disk in the background; the command does not wait for that. Anything
still on the VM (uncommitted work, unsynced files) is lost, so push or download
what you need before deleting.

Only use 'session terminate' instead when you intend to 'session restore' the
same session later — a terminated session keeps its disk (and keeps using disk
space) until it is deleted.

The command refuses while the machine state is "unknown" (e.g. its node lost
connectivity); retry once the state settles.`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

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
			if err := svc.DeleteSession(cmd.Context(), workspaceID, sessionID); err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Deleted session %s\n", sessionID)
				return err
			}
			return nil
		},
	}
}
