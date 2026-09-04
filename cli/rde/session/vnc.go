package session

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newVNCCmd() *cobra.Command {
	var forwardPort int
	var format string
	c := &cobra.Command{
		Use:   "vnc SESSION_ID",
		Short: "Print VNC connection details, or forward the endpoint to a local port",
		Long: `Print the VNC connection details (address, host, port, username, password,
and a ready-to-use vnc:// URL) for a session.

The VNC password is ephemeral and tied to this session. Avoid pasting the
output into chat or sharing it — anyone with the URL can connect to the
session. ` + "`rde session view`" + ` and other commands intentionally hide it.

In raw mode the URL is the only thing on stdout, so it's safe to pipe:

  open "$(bitrise rde session vnc SESSION_ID)"

In --format json/yml mode a fully-decomposed {address, host, port, username,
password, url} object is emitted — host and port are always discrete fields,
so a caller building its own connection never has to parse the address or URL.

Pass --forward PORT to open an SSH tunnel and expose the session's VNC endpoint
on a local port, then block until Ctrl-C (use 0 to auto-pick a free port):

  bitrise rde session vnc SESSION_ID --forward 0      # auto-pick a local port
  bitrise rde session vnc SESSION_ID --forward 5901   # bind localhost:5901

A native VNC client (macOS Screen Sharing, Remmina, …) can then connect to the
printed localhost address. The tunnel rides the same SSH connection the CLI
already uses, so no direct network route to the session is required and no
credentials are embedded in a URL handed to the OS. Prefer ` + "`rde session open-vnc`" + `
when you just want to launch your viewer against a directly-reachable endpoint.`,
		Example: `  bitrise rde session vnc SESSION_ID
  bitrise rde session vnc SESSION_ID --format json
  bitrise rde session vnc SESSION_ID --forward 5901
  open "$(bitrise rde session vnc SESSION_ID)"`,
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
			// Reject the unsupported flag combo before resolving the name — no
			// point in a lookup round-trip for a request we won't serve
			// (mirrors `view --watch`). --forward runs a long-lived tunnel, not
			// a single-object result, so it can't satisfy the machine-readable
			// contract.
			forwarding := cmd.Flags().Changed("forward")
			if forwarding && output.Format != output.FormatRaw {
				return fmt.Errorf("--forward cannot be combined with --format %s (it runs a long-lived tunnel, not a single-object result)", output.Format)
			}

			svc := internalrde.NewService(client)
			sessionID, err := svc.ResolveSessionID(cmd.Context(), workspaceID, args[0])
			if err != nil {
				return err
			}

			if forwarding {
				return runVNCForward(cmd, svc, workspaceID, sessionID, forwardPort)
			}

			creds, err := svc.GetSessionVNC(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, creds, renderVNCCredentials)
		},
	}
	c.Flags().IntVar(&forwardPort, "forward", 0,
		"forward the session's VNC endpoint to this local port, then block until Ctrl-C; use 0 to auto-pick a free port")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

// runVNCForward opens the local-port tunnel and blocks until Ctrl-C. It prints
// the ready-to-use local vnc:// URL on stdout (one line, same as the non-forward
// path) and a human status on stderr.
func runVNCForward(cmd *cobra.Command, svc *internalrde.Service, workspaceID, sessionID string, localPort int) error {
	// Ctrl-C cancels the tunnel and returns cleanly rather than hard-killing.
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	// onReady fires once the local listener is up; ForwardVNC hands back the
	// session's credentials (fetched exactly once, service-side) so the printed
	// URL points at the local address.
	onReady := func(localAddr string, creds internalrde.VNCCredentials) {
		line := localAddr
		if host, portStr, splitErr := net.SplitHostPort(localAddr); splitErr == nil {
			if p, atoiErr := strconv.Atoi(portStr); atoiErr == nil {
				line = internalrde.FormatVNCURL(host, p, creds.Username, creds.Password)
			}
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
		if !cmdutil.IsQuiet(cmd) {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
				"Forwarding the session's VNC endpoint to %s — connect your VNC client there. Press Ctrl-C to stop.\n", localAddr)
		}
	}

	if err := svc.ForwardVNC(ctx, workspaceID, sessionID, localPort, onReady); err != nil {
		return err
	}
	if !cmdutil.IsQuiet(cmd) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Stopped forwarding.")
	}
	return nil
}

func renderVNCCredentials(w io.Writer, creds internalrde.VNCCredentials) error {
	_, err := fmt.Fprintln(w, creds.URL)
	return err
}
