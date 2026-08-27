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
		Short: "Permanently delete a session",
		Args:  cmdutil.RequireArgs("SESSION_ID"),
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
