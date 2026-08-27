package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

// openVNCResult is the --format json/yml shape of `session open-vnc`. The
// password is intentionally omitted — `open-vnc` hands the URL to the OS
// handler, so there's no reason to also print it.
type openVNCResult struct {
	Opened   bool   `json:"opened" yaml:"opened"`
	Address  string `json:"address" yaml:"address"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
}

// urlOpener spawns the platform-appropriate URL handler. Overridable in
// tests so we can assert what we'd run without launching anything.
var urlOpener = cmdutil.OpenVNCURL

func newOpenVNCCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "open-vnc SESSION_ID",
		Short: "Open a session's VNC endpoint in the OS-default viewer",
		Long: `Hand the session's VNC URL to the operating system's default URL handler:

  - macOS:    /usr/bin/open
  - Linux:    xdg-open (must be installed; install x11-utils or similar)
  - Windows:  cmd /c start

The OS launches whatever app is registered for vnc:// (Screen Sharing on
macOS by default; Remmina/Vinagre on Linux; a third-party client on Windows).

The URL contains the ephemeral VNC password as a userinfo component. The
URL is passed as an argv element to the OS handler, so it is briefly
visible to other processes on the machine that can read this process's
argv (e.g. ` + "`ps`" + `). On a single-user dev machine this is usually fine;
on a shared host, prefer ` + "`rde session vnc`" + ` and paste the URL into your
viewer manually.`,
		Example: `  bitrise rde session open-vnc SESSION_ID
  bitrise rde session open-vnc SESSION_ID --format json`,
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
			creds, err := svc.GetSessionVNC(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			if err := urlOpener(cmd.Context(), creds.URL); err != nil {
				return fmt.Errorf("open VNC URL: %w", err)
			}
			res := openVNCResult{Opened: true, Address: creds.Address, Username: creds.Username}
			if output.Format != output.FormatRaw {
				return output.Render(cmd.OutOrStdout(), output.Format, res, nil)
			}
			if !cmdutil.IsQuiet(cmd) {
				_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Opened VNC viewer for %s\n", creds.Address)
				return err
			}
			return nil
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}
