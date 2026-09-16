package session

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newSSHCmd() *cobra.Command {
	var (
		passwordOnly bool
		format       string
	)
	c := &cobra.Command{
		Use:   "ssh SESSION_ID",
		Short: "Print SSH connection details (command, host, port, user, password) for a session",
		Long: `Print the SSH connection details for a running session: a ready-to-run ssh
command line, the host, port and user it decomposes into, and the session's
SSH password.

The password is ephemeral and tied to this session. ` + "`rde session view`" + ` and
` + "`--format json`" + ` on other commands intentionally hide it; this command is the
opt-in way to get it — for scp, an SSH tunnel to a device session's ports, or
any tool that is not ` + "`rde session exec`" + ` (which dials for you and needs none
of this).

Human mode prints the command on the first line and the password on the
second. --password-only prints just the password, so it can be fed to an
SSH_ASKPASS helper without parsing:

  RDE_SSH_PASSWORD="$(bitrise rde session ssh SESSION_ID --password-only)"

--format json emits {address, host, port, user, password, command}.`,
		Example: `  bitrise rde session ssh SESSION_ID
  bitrise rde session ssh SESSION_ID --format json
  RDE_SSH_PASSWORD="$(bitrise rde session ssh SESSION_ID --password-only)"`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}
			if passwordOnly && output.Format == output.FormatJSON {
				return fmt.Errorf("--password-only cannot be combined with --format json (read .password from the JSON instead)")
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
			creds, err := svc.GetSessionSSH(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			if passwordOnly {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), creds.Password)
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, creds, renderSSHCredentials)
		},
	}
	c.Flags().BoolVar(&passwordOnly, "password-only", false, "print only the SSH password (for SSH_ASKPASS helpers and scripts)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderSSHCredentials(w io.Writer, creds internalrde.SSHCredentials) error {
	_, err := fmt.Fprintf(w, "%s\n%s\n", creds.Command, creds.Password)
	return err
}
