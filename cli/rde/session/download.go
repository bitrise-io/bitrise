package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

func newDownloadCmd() *cobra.Command {
	var onlyContents bool
	c := &cobra.Command{
		Use:   "download SESSION_ID REMOTE_PATH LOCAL_PATH",
		Short: "Download a file or directory from a session",
		Long: `Download a file or directory from a running session to the local machine.

The remote path is tar+gzip'd server-side, served via a signed download URL,
then extracted into LOCAL_PATH locally.

When REMOTE_PATH is a directory, the directory itself is recreated inside
LOCAL_PATH by default. Pass --only-contents to drop just its contents into
LOCAL_PATH instead.`,
		Example: `  bitrise rde session download SESSION_ID /Users/vagrant/project/build ./build
  bitrise rde session download SESSION_ID /Users/vagrant/logs ./logs --only-contents`,
		Args: cmdutil.RequireArgs("SESSION_ID", "REMOTE_PATH", "LOCAL_PATH"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			sessionID, sourcePath, localDest := args[0], args[1], args[2]
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)
			sessionID, err = svc.ResolveSessionID(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Downloading %s → %s from session %s…\n", sourcePath, localDest, sessionID)
			}
			if err := svc.DownloadFile(cmd.Context(), workspaceID, sessionID, sourcePath, localDest, onlyContents); err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Download complete.\n")
			}
			return nil
		},
	}
	c.Flags().BoolVar(&onlyContents, "only-contents", false, "when REMOTE_PATH is a directory, extract only its contents (not the directory itself)")
	return c
}
