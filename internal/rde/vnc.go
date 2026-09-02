package rde

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// VNCCredentials is the credential bundle a session exposes for VNC. The
// JSON tags define the stable shape used by `rde session vnc --output json`.
// The fields mirror what the backend returns (address, username, password)
// plus a pre-built `vnc://` URL ready to hand to an OS handler.
//
// Host and Port are the address decomposed into discrete fields, so callers
// that need to build their own connection (a bridge, a native client) never
// have to parse `address` or the URL — the endpoint is always fully qualified.
type VNCCredentials struct {
	Address  string `json:"address" yaml:"address"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	URL      string `json:"url" yaml:"url"`
}

// GetSessionVNC fetches the session and returns its VNC credentials,
// erroring clearly when the session has no VNC endpoint yet (still
// provisioning, terminated, or a Linux template that doesn't expose VNC).
func (s *Service) GetSessionVNC(ctx context.Context, workspaceID, sessionID string) (VNCCredentials, error) {
	if s.client == nil {
		return VNCCredentials{}, errClient()
	}
	sess, err := s.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		return VNCCredentials{}, err
	}
	return VNCCredentialsFromSession(sess)
}

// SessionExposesVNC reports whether the session currently has a VNC endpoint.
// VNC is exposed by macOS sessions once they are running; Linux sessions have
// none. Callers use it to decide whether to offer VNC-related features for a
// session at all, rather than letting GetSessionVNC fail later.
func (s *Service) SessionExposesVNC(ctx context.Context, workspaceID, sessionID string) (bool, error) {
	if s.client == nil {
		return false, errClient()
	}
	sess, err := s.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		return false, err
	}
	return sess.VNCAddress != "", nil
}

// VNCCredentialsFromSession assembles a credentials bundle from an already
// loaded Session. Split from GetSessionVNC so callers that already hold a
// Session (e.g. `session create --wait`) can reuse it without a second GET.
func VNCCredentialsFromSession(sess Session) (VNCCredentials, error) {
	if sess.VNCAddress == "" {
		if sess.Status != "running" && sess.Status != "" {
			return VNCCredentials{}, fmt.Errorf(
				"session VNC is not available (status: %q); VNC is exposed once the session is running",
				sess.Status,
			)
		}
		return VNCCredentials{}, fmt.Errorf(
			"session has no VNC endpoint (the template may not expose VNC, or the session is still provisioning)",
		)
	}
	host, port, err := parseVNCHostPort(sess.VNCAddress)
	if err != nil {
		return VNCCredentials{}, err
	}
	creds := VNCCredentials{
		Address:  sess.VNCAddress,
		Host:     host,
		Port:     port,
		Username: sess.VNCUsername,
		Password: sess.VNCPassword,
	}
	creds.URL = FormatVNCURL(host, port, sess.VNCUsername, sess.VNCPassword)
	return creds, nil
}

// parseVNCHostPort accepts the address shapes the backend has used:
// "vnc://host:port", "host:port", or bare "host". A missing port defaults
// to 5900 (standard VNC). The host/port split is shared with the SSH address
// parser (see splitHostPort), so an IPv6 endpoint survives it intact.
func parseVNCHostPort(addr string) (string, int, error) {
	s := strings.TrimSpace(addr)
	s = strings.TrimPrefix(s, "vnc://")
	if s == "" {
		return "", 0, fmt.Errorf("empty VNC address")
	}
	host, port, err := splitHostPort(s)
	if err != nil {
		return "", 0, fmt.Errorf("VNC address %q: %w", addr, err)
	}
	if host == "" {
		return "", 0, fmt.Errorf("VNC address %q has no host", addr)
	}
	if port == 0 {
		port = 5900
	}
	return host, port, nil
}

func parsePort(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty port")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid port %q", s)
		}
		n = n*10 + int(c-'0')
		if n > 65535 {
			return 0, fmt.Errorf("port %q out of range", s)
		}
	}
	if n == 0 {
		return 0, fmt.Errorf("invalid port %q", s)
	}
	return n, nil
}

// localScreenSharingPort is where macOS Screen Sharing (ARD/RFB) listens on
// the session VM itself. The port in the backend's vnc_address is the
// externally-published relay port (e.g. …:20852), not a port on the VM, so it
// can't be the tunnel's remote target — the raw RFB server is always on the
// well-known 5900 locally.
const localScreenSharingPort = 5900

// ForwardVNC opens an SSH tunnel to the session and forwards the VM's VNC
// server to a local TCP port, blocking until ctx is cancelled. localPort 0
// auto-picks a free port; the chosen "127.0.0.1:port" is reported via onReady
// once the listener is accepting, so a caller can print connection details.
//
// The tunnel targets the session VM's loopback Screen Sharing port — the
// standard `ssh -L LOCAL:localhost:5900` recipe. The SSH connection terminates
// on the VM, so dialing 127.0.0.1:5900 there reaches the raw RFB server
// directly, bypassing the external relay. That yields a plain-RFB localhost
// endpoint any VNC client (or a websockify/noVNC bridge) can consume, with no
// credentials embedded in a handed-off URL and no direct route to the relay
// required.
func (s *Service) ForwardVNC(ctx context.Context, workspaceID, sessionID string, localPort int, onReady func(localAddr string, creds VNCCredentials)) error {
	if s.client == nil {
		return errClient()
	}
	sess, err := s.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		return err
	}
	// VNCCredentialsFromSession validates the endpoint (clear error when the
	// session exposes no VNC yet) and yields the username/password the caller
	// prints, so the session is fetched exactly once. The tunnel's remote
	// target is always the VM's loopback Screen Sharing port (see
	// localScreenSharingPort), not anything derived from these credentials.
	creds, err := VNCCredentialsFromSession(sess)
	if err != nil {
		return err
	}
	target, err := sshTargetForSession(sess)
	if err != nil {
		return err
	}
	client, err := dialSSHWithRetry(ctx, target)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck // forward errors take precedence; nothing actionable on close failure

	remoteAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(localScreenSharingPort))
	return client.forwardLocal(ctx, localPort, remoteAddr, func(localAddr string) {
		if onReady == nil {
			return
		}
		onReady(localAddr, creds)
	})
}

// FormatVNCURL emits a `vnc://[user[:pass]@]host:port` URL. Credentials are
// percent-encoded for the userinfo component, so a `@` or `:` in the password
// doesn't desync the parser on the receiving side. Exported so a caller that
// forwards the endpoint to a local port (see ForwardVNC) can present a URL
// pointing at the local address.
//
// url.Userinfo is what does the encoding: url.QueryEscape is form encoding,
// which emits `+` for a space and `%2B` for a literal `+` — both of which a
// receiving client decodes back to the wrong password.
func FormatVNCURL(host string, port int, user, pass string) string {
	u := url.URL{
		Scheme: "vnc",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	switch {
	case user != "" && pass != "":
		u.User = url.UserPassword(user, pass)
	case user != "":
		u.User = url.User(user)
	case pass != "":
		u.User = url.UserPassword("", pass)
	}
	return u.String()
}
