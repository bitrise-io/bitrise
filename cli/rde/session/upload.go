package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

func newUploadCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "upload SESSION_ID LOCAL_PATH REMOTE_FOLDER",
		Short: "Upload a local file or directory into a session",
		Long: `Upload a local file or directory into a running session.

The local path is tarred + gzipped, uploaded to cloud storage via a signed
URL, then extracted on the session VM inside REMOTE_FOLDER.

REMOTE_FOLDER is always a DIRECTORY (absolute path; created if missing, owned
by the session user). It is never the name of the file you are sending:

  - a directory: its contents are extracted into REMOTE_FOLDER (not the
    directory itself), overwriting files of the same name;
  - a single file: it lands as REMOTE_FOLDER/<basename>. To replace one
    remote file, upload it into the file's PARENT directory. Naming the
    file's own path as REMOTE_FOLDER is rejected by the server.

Extracted files belong to the session user (vagrant on macOS, ubuntu on
Linux); local ownership is not carried over.`,
		Example: `  bitrise rde session upload SESSION_ID ./project /Users/vagrant/project
  # replace one file: upload it into its parent directory
  bitrise rde session upload SESSION_ID ./app/AndroidManifest.xml /home/ubuntu/app
  bitrise rde session upload SESSION_ID ./build.tar.gz /Users/vagrant/artifacts`,
		Args: cmdutil.RequireArgs("SESSION_ID", "LOCAL_PATH", "REMOTE_FOLDER"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			sessionID, sourcePath, destFolder := args[0], args[1], args[2]
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
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Uploading %s → %s on session %s…\n", sourcePath, destFolder, sessionID)
			}
			if err := svc.UploadFile(cmd.Context(), workspaceID, sessionID, sourcePath, destFolder); err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Upload complete.\n")
			}
			return nil
		},
	}
	return c
}
