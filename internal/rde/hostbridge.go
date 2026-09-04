package rde

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// ActionOpenVNC opens a VNC viewer on the user's local machine, pointed at this
// session's desktop. It is the first action the host bridge exposes; the path
// segment doubles as the route the in-session skill calls.
const ActionOpenVNC = "open-vnc"

// ActionDownload and ActionUpload move files between the session and the user's
// local machine. Unlike open-vnc they apply to every session (not just ones
// that expose a VNC endpoint). The path segment doubles as the route the
// in-session skill calls.
const (
	ActionDownload = "download"
	ActionUpload   = "upload"
)

// Fixed remote paths. They are trusted constants with shell-safe characters and
// a leading ~, so they are written into remote commands UNQUOTED — quoting them
// would suppress the shell's tilde expansion.
const (
	hostBridgeControlDir  = "~/.config/rde"
	hostBridgeControlFile = hostBridgeControlDir + "/host-bridge.json"
	hostBridgeSkillDir    = "~/.claude/skills/rde-host"
	hostBridgeSkillFile   = hostBridgeSkillDir + "/SKILL.md"
)

const (
	// bridgeStartTimeout bounds the initial dial+listen+write so a session that
	// denies remote forwarding (or is slow) degrades quickly instead of delaying
	// the user reaching Claude.
	bridgeStartTimeout = 10 * time.Second
	// bridgeActionTimeout caps how long a single host action may run.
	bridgeActionTimeout = 30 * time.Second
	// bridgeReconnectBackoff paces reconnect attempts after a dropped connection.
	bridgeReconnectBackoff = 3 * time.Second
	// bridgeMaxRequestBody caps request bodies. Current actions take no body;
	// the cap is defense for when later actions do.
	bridgeMaxRequestBody = 64 << 10
)

// hostBridgeSkillHeader is the shared preamble of the provisioned skill: the
// frontmatter plus the bridge mechanics (read the control file, POST with the
// token). Each registered action appends its own section, so the skill always
// describes exactly the actions available in this session — and nothing else.
//
//go:embed hostbridge_skill.md
var hostBridgeSkillHeader string

// HostAction is one entry in the bridge's allowlist. It is a struct, not a bare
// func, so an action can carry metadata alongside its handler without reshaping
// the allowlist or every existing action.
type HostAction struct {
	// Handle runs the action and returns a JSON-marshalable result. ctx is
	// bounded by bridgeActionTimeout; r exposes the request for actions that take
	// parameters (open-vnc takes none).
	Handle func(ctx context.Context, r *http.Request) (any, error)

	// SkillSection is the Markdown section appended to the skill header to tell
	// Claude when and how to use this action. Only sections for registered
	// actions are written, so the skill never advertises a capability the
	// session lacks. (Reserved for future metadata too, e.g. a confirmation
	// policy for side-effecting actions like a file download.)
	SkillSection string

	// Timeout caps how long this action's Handle may run. Zero means use
	// bridgeActionTimeout (the 30s default suited to quick actions like
	// open-vnc); file transfers set a much larger value since a single archive
	// can take minutes to move through cloud storage.
	Timeout time.Duration
}

// HostBridge exposes a fixed allowlist of "host actions" to the Claude Code
// instance running inside an RDE session. It opens a loopback listener on the
// session over an SSH reverse forward and serves HTTP on it; the in-session
// Claude reaches it by reading a control file (URL + token) and calling the
// endpoint. The reverse forward is only a control channel — actions execute
// locally (open-vnc launches the viewer on the user's machine and the VNC
// password never leaves the local side).
//
// Everything is best-effort and isolated from the foreground Claude session: if
// the session sshd denies remote forwarding, Start returns an error and the
// caller simply runs without the bridge.
type HostBridge struct {
	Service     *Service
	WorkspaceID string
	SessionID   string

	// Actions is the allowlist, keyed by the path segment a request targets
	// (e.g. "open-vnc"). It is built by the caller so this layer never depends on
	// the cmd packages.
	Actions map[string]HostAction

	// Debug, if set, receives diagnostics about degraded or failed bridge
	// activity. The bridge never disrupts the foreground session, so problems
	// surface only here.
	Debug func(format string, args ...any)

	credential string // per-session bearer secret (random hex)
	srv        *http.Server

	mu     sync.Mutex
	client *sshClient
	ln     net.Listener
}

// Start dials the session, opens the reverse forward, writes the control file
// and the skill, and prepares the HTTP server. It runs under a short bounded
// context so a denied/slow forward fails fast. On any error the caller should
// degrade unconditionally (continue without the bridge) — the forward-denied
// error is not a typed sentinel. The skill is written only on success, so a
// failed forward never advertises a capability that cannot work.
//
// The skill must be in place before Claude launches: Claude only watches
// ~/.claude/skills if that directory exists at startup, and on a fresh VM it
// does not — Start creates it.
func (b *HostBridge) Start(ctx context.Context) error {
	if b.Service == nil || b.Service.client == nil {
		return errClient()
	}
	if len(b.Actions) == 0 {
		return fmt.Errorf("host bridge: no actions registered")
	}
	if err := b.ensureCredential(); err != nil {
		return err
	}

	if err := b.connect(ctx); err != nil {
		return err
	}
	skillCtx, cancel := context.WithTimeout(ctx, bridgeStartTimeout)
	defer cancel()
	if err := b.writeSkill(skillCtx); err != nil {
		b.closeConn()
		return err
	}

	b.srv = &http.Server{
		Handler:           b.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		// No write deadline: each action bounds its own duration via
		// serveAction's per-action timeout, and a file transfer can legitimately
		// run for minutes. The read side stays bounded (bodies are tiny JSON read
		// immediately), and the listener is loopback-only and token-gated, so a
		// slow-write attack is not a concern.
		WriteTimeout: 0,
	}
	return nil
}

// Serve runs the bridge until ctx is cancelled. On a dropped connection it
// re-dials, re-listens, and rewrites the control file with the new port, so the
// bridge survives the same network blips the interactive attach reconnects
// through. It is best-effort and never returns an error: failures are logged
// via Debug and retried. Serve is a no-op if Start was not called or failed.
func (b *HostBridge) Serve(ctx context.Context) {
	if b.srv == nil {
		return
	}
	// On cancellation, closing the server unblocks Serve with ErrServerClosed and
	// closes the active listener; closing the SSH client stops its forwarded
	// channel handlers. The stopped channel gives the watcher a second exit, so
	// it doesn't outlive a Serve that returned on its own (failed reconnect, no
	// connection, or an already-closed server).
	stopped := make(chan struct{})
	defer close(stopped)
	go func() {
		select {
		case <-ctx.Done():
			b.Close()
		case <-stopped:
		}
	}()

	for {
		c, ln := b.currentConn()
		if c == nil || ln == nil {
			return
		}
		// keepAlive closes the client on the first dead probe, which unblocks
		// Accept — so a silently half-open connection is detected in seconds
		// rather than waiting out the OS TCP timeout.
		done := make(chan struct{})
		go c.keepAlive(done)

		err := b.srv.Serve(ln)
		close(done)

		// ErrServerClosed means the server itself was shut down, which no
		// reconnect can undo — treat it as intentional even when ctx is still
		// live, or the loop spins re-listening on a dead server. (A caller can
		// legitimately Close() before cancelling: `rde claude` defers both, and
		// LIFO runs Close first.)
		if ctx.Err() != nil || errors.Is(err, http.ErrServerClosed) {
			return // intentional shutdown
		}
		// The connection dropped. Close the dead client+listener (closing the
		// client also stops its forwarded-channel goroutines, so reconnecting
		// does not leak them) and re-establish.
		b.debugf("host bridge connection lost (%v); reconnecting", err)
		b.closeConn()
		if !b.reconnect(ctx) {
			return
		}
	}
}

// Close tears the bridge down. Safe to call repeatedly and even if Start failed.
func (b *HostBridge) Close() {
	if b.srv != nil {
		_ = b.srv.Close() //nolint:errcheck // teardown; nothing actionable on close failure
	}
	b.closeConn()
}

// connect dials the session, opens the reverse-forward listener, and writes the
// control file with the freshly-allocated remote port. Used by both Start and
// the reconnect loop.
func (b *HostBridge) connect(ctx context.Context) error {
	// Bound each attempt so a slow dial/write fails fast; Start surfaces the
	// error and the reconnect loop just retries after a backoff.
	ctx, cancel := context.WithTimeout(ctx, bridgeStartTimeout)
	defer cancel()

	sess, err := b.Service.GetSession(ctx, b.WorkspaceID, b.SessionID)
	if err != nil {
		return fmt.Errorf("fetch session: %w", err)
	}
	target, err := sshTargetForSession(sess)
	if err != nil {
		return err
	}
	client, err := dialSSHWithRetry(ctx, target)
	if err != nil {
		return err
	}
	ln, err := client.listenRemote("127.0.0.1:0")
	if err != nil {
		_ = client.Close() //nolint:errcheck // dial cleanup; the listen error is what matters
		return fmt.Errorf("open reverse forward (the session may not permit remote port forwarding): %w", err)
	}
	port, err := listenerPort(ln)
	if err != nil {
		_ = ln.Close()     //nolint:errcheck // cleanup
		_ = client.Close() //nolint:errcheck // cleanup
		return err
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := writeControlFile(ctx, client, url, b.credential); err != nil {
		_ = ln.Close()     //nolint:errcheck // cleanup
		_ = client.Close() //nolint:errcheck // cleanup
		return err
	}
	b.setConn(client, ln)
	return nil
}

// reconnect retries connect with backoff until it succeeds or ctx is cancelled.
func (b *HostBridge) reconnect(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(bridgeReconnectBackoff):
		}
		if err := b.connect(ctx); err != nil {
			b.debugf("host bridge reconnect failed: %v", err)
			continue
		}
		return true
	}
}

func (b *HostBridge) ensureCredential() error {
	if b.credential != "" {
		return nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("generate host bridge credential: %w", err)
	}
	b.credential = hex.EncodeToString(buf)
	return nil
}

func (b *HostBridge) setConn(c *sshClient, ln net.Listener) {
	b.mu.Lock()
	b.client, b.ln = c, ln
	b.mu.Unlock()
}

func (b *HostBridge) currentConn() (*sshClient, net.Listener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.client, b.ln
}

func (b *HostBridge) closeConn() {
	b.mu.Lock()
	c, ln := b.client, b.ln
	b.client, b.ln = nil, nil
	b.mu.Unlock()
	if ln != nil {
		_ = ln.Close() //nolint:errcheck // teardown
	}
	if c != nil {
		_ = c.Close() //nolint:errcheck // teardown
	}
}

func (b *HostBridge) writeSkill(ctx context.Context) error {
	c, _ := b.currentConn()
	if c == nil {
		return fmt.Errorf("host bridge: no connection")
	}
	return remoteWriteFile(ctx, c, hostBridgeSkillDir, hostBridgeSkillFile, b.buildSkill(), false)
}

// buildSkill assembles the SKILL.md for this session: the shared header plus the
// section of each registered action (ordered by action name for a stable
// result). Because only registered actions contribute, the skill describes
// exactly what this session can do — e.g. a Linux session without VNC never
// gets the open-vnc section.
func (b *HostBridge) buildSkill() string {
	names := make([]string, 0, len(b.Actions))
	for name := range b.Actions {
		names = append(names, name)
	}
	sort.Strings(names)

	var sb strings.Builder
	sb.WriteString(strings.TrimRight(hostBridgeSkillHeader, "\n"))
	for _, name := range names {
		section := strings.TrimSpace(b.Actions[name].SkillSection)
		if section == "" {
			continue
		}
		sb.WriteString("\n\n")
		sb.WriteString(section)
	}
	sb.WriteString("\n")
	return sb.String()
}

// writeControlFile drops the bridge URL and bearer token into a 0600 file the
// in-session skill reads before each call. The JSON is marshaled in Go (so it is
// always valid regardless of values) and shell-quoted as a whole for the remote
// write.
func writeControlFile(ctx context.Context, c *sshClient, url, credential string) error {
	payload, err := json.Marshal(map[string]string{"url": url, "token": credential})
	if err != nil {
		return fmt.Errorf("encode host bridge control file: %w", err)
	}
	return remoteWriteFile(ctx, c, hostBridgeControlDir, hostBridgeControlFile, string(payload), true)
}

// remoteWriteFile writes content to path on the session, creating dir first and
// optionally restricting the file to 0600.
func remoteWriteFile(ctx context.Context, c *sshClient, dir, path, content string, private bool) error {
	res, err := c.run(ctx, remoteWriteCommand(dir, path, content, private))
	if err != nil {
		return fmt.Errorf("write %s on session: %w", path, err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("write %s on session (exit %d): %s", path, res.ExitCode, strings.TrimSpace(res.Stderr))
	}
	return nil
}

// remoteWriteCommand builds remoteWriteFile's shell command. dir and path are
// trusted constants and are left unquoted so their leading ~ expands; only
// content is shell-quoted.
//
// A private file is created under `umask 077` rather than only chmod'd after
// the fact, so it is never briefly world-readable — the control file holds the
// bridge's bearer token. The chmod stays as well: `>` truncation preserves an
// existing file's mode, which umask cannot fix, and a reconnect rewrites that
// same file.
func remoteWriteCommand(dir, path, content string, private bool) string {
	redirect := fmt.Sprintf("printf '%%s' %s > %s", shellSingleQuote(content), path)
	if !private {
		return "mkdir -p " + dir + " && " + redirect
	}
	return "mkdir -p " + dir + " && (umask 077 && " + redirect + ") && chmod 600 " + path
}

func listenerPort(ln net.Listener) (int, error) {
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("host bridge: unexpected listener address type %T", ln.Addr())
	}
	return addr.Port, nil
}

func (b *HostBridge) handler() http.Handler {
	mux := http.NewServeMux()
	for name, action := range b.Actions {
		// Go 1.22+ method-aware routing: a matching path with the wrong method
		// yields 405, an unmatched path yields 404 (the allowlist boundary).
		mux.HandleFunc("POST /"+name, func(w http.ResponseWriter, r *http.Request) {
			b.serveAction(w, r, action)
		})
	}
	return b.authenticated(mux)
}

// authenticated gates every request on the bearer token and caps the body size.
func (b *HostBridge) authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, bridgeMaxRequestBody)
		if !b.authorized(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (b *HostBridge) authorized(r *http.Request) bool {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return false
	}
	got := []byte(strings.TrimPrefix(h, prefix))
	want := []byte(b.credential)
	// Length check first: ConstantTimeCompare returns 0 on a length mismatch but
	// is only constant-time for equal lengths, so reject mismatches cheaply here.
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

func (b *HostBridge) serveAction(w http.ResponseWriter, r *http.Request, action HostAction) {
	timeout := action.Timeout
	if timeout <= 0 {
		timeout = bridgeActionTimeout
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	result, err := action.Handle(ctx, r)
	if err != nil {
		b.debugf("host action %s: %v", r.URL.Path, err)
		b.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	b.writeJSON(w, http.StatusOK, result)
}

func (b *HostBridge) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		b.debugf("host bridge encode response: %v", err)
	}
}

func (b *HostBridge) debugf(format string, args ...any) {
	if b.Debug != nil {
		b.Debug(format, args...)
	}
}
