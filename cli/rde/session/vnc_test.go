package session

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestVNCCmd_HumanPrintsOnlyURL(t *testing.T) {
	// Raw mode emits a single line (the vnc:// URL) so callers can pipe
	// it into `open` or another launcher without parsing.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"host.example:5901","vncUsername":"vagrant","vncPassword":"pw@with:special"
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := strings.TrimSpace(stdout)
	want := "vnc://vagrant:pw%40with%3Aspecial@host.example:5901"
	if got != want {
		t.Errorf("stdout = %q, want exactly %q", got, want)
	}
}

func TestVNCCmd_JSONIncludesPassword(t *testing.T) {
	// The dedicated vnc subcommand exists *because* the password is needed.
	// Its JSON shape includes it. The stable view-JSON contract does NOT —
	// that's covered by TestViewCmd_JSONOmitsVNCPassword below.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"vnc://host.example:5900","vncUsername":"vagrant","vncPassword":"hunter2"
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON, Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got internalrde.VNCCredentials
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got.Username != "vagrant" || got.Password != "hunter2" {
		t.Errorf("credentials = %+v, want vagrant/hunter2", got)
	}
	if !strings.Contains(got.URL, "hunter2") {
		t.Errorf("url %q should embed the password", got.URL)
	}
}

func TestVNCCmd_JSONIncludesHostPort(t *testing.T) {
	// The decomposed host/port fields let a caller build its own connection
	// without ever parsing `address` or the URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"host.example:5901","vncUsername":"vagrant","vncPassword":"hunter2"
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON, Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got internalrde.VNCCredentials
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got.Host != "host.example" || got.Port != 5901 {
		t.Errorf("host/port = %q/%d, want host.example/5901", got.Host, got.Port)
	}
}

func TestVNCCmd_ForwardRejectsMachineReadableOutput(t *testing.T) {
	// --forward is a long-lived tunnel; it can't satisfy the single-object
	// contract, so the combination is rejected before any network call —
	// including the session-name lookup. A bare name argument (not a UUID)
	// would otherwise resolve via ListSessions first, so exercising it here
	// pins the rejection ahead of name resolution. Covers both machine
	// formats, not just json.
	for _, format := range []string{output.FormatJSON, output.FormatYML} {
		for _, arg := range []string{uuidSession, "my-session"} {
			t.Run(format+"/"+arg, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					t.Error("no HTTP call should be made when --forward is rejected for a machine-readable format")
				}))
				defer srv.Close()

				_, _, err := cmdtest.Run(t, newVNCCmd(), cmdtest.Opts{
					RDEAPIBaseURL:      srv.URL,
					DefaultWorkspaceID: "ws-1",
					Format:             format,
					Args:               []string{arg, "--forward", "5901"},
				})
				if err == nil || !strings.Contains(err.Error(), "--forward cannot be combined with --format") {
					t.Errorf("err = %v, want --forward/--format rejection", err)
				}
			})
		}
	}
}

func TestVNCCmd_NoEndpointError(t *testing.T) {
	// A session with no vncAddress (e.g. a Linux template that doesn't
	// expose VNC, or one still provisioning) must error rather than print
	// an empty/half-baked URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err == nil || !strings.Contains(err.Error(), "VNC") {
		t.Errorf("err = %v, want a VNC-not-available error", err)
	}
}

func TestViewCmd_JSONOmitsVNCPassword(t *testing.T) {
	// Companion to view's SSH-password omission check. The view JSON
	// contract must never leak the VNC password — users who need it call
	// `session vnc` explicitly.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"h:5900","vncUsername":"u","vncPassword":"vnchunter2"
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON, Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "vnchunter2") || strings.Contains(stdout, "vnc_password") {
		t.Errorf("VNC password leaked into view JSON:\n%s", stdout)
	}
	if !strings.Contains(stdout, "vnc_address") {
		t.Errorf("vnc_address missing from view JSON:\n%s", stdout)
	}
}

func TestOpenVNCCmd_HandsURLToOpener(t *testing.T) {
	// Capture what we'd pass to the OS handler without actually shelling
	// out. The URL must carry the password (so the OS handler can pass it
	// to the VNC client); stdout must stay empty in raw mode so the
	// password never lands in shell history / scrollback.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"host.example:5900","vncUsername":"vagrant","vncPassword":"hunter2"
		}}`)
	}))
	defer srv.Close()

	var gotURL string
	prev := urlOpener
	urlOpener = func(_ context.Context, url string) error { gotURL = url; return nil }
	defer func() { urlOpener = prev }()

	stdout, stderr, err := cmdtest.Run(t, newOpenVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	wantURL := "vnc://vagrant:hunter2@host.example:5900"
	if gotURL != wantURL {
		t.Errorf("opener got URL %q", gotURL)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("raw stdout should be empty (no password leakage), got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Opened VNC viewer") {
		t.Errorf("stderr missing confirmation line:\n%s", stderr)
	}
}

func TestOpenVNCCmd_JSONOmitsPassword(t *testing.T) {
	// JSON mode is the same: confirmation envelope only, no password.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"h:5900","vncUsername":"u","vncPassword":"hunter2"
		}}`)
	}))
	defer srv.Close()

	prev := urlOpener
	urlOpener = func(_ context.Context, _ string) error { return nil }
	defer func() { urlOpener = prev }()

	stdout, _, err := cmdtest.Run(t, newOpenVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON, Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "hunter2") || strings.Contains(stdout, "password") {
		t.Errorf("password leaked into open-vnc JSON:\n%s", stdout)
	}
	var got openVNCResult
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if !got.Opened || got.Address != "h:5900" || got.Username != "u" {
		t.Errorf("unexpected JSON: %+v", got)
	}
}

func TestOpenVNCCmd_OpenerErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"h:5900","vncUsername":"u","vncPassword":"pw"
		}}`)
	}))
	defer srv.Close()

	prev := urlOpener
	urlOpener = func(_ context.Context, _ string) error { return errors.New("xdg-open not found") }
	defer func() { urlOpener = prev }()

	_, _, err := cmdtest.Run(t, newOpenVNCCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err == nil || !strings.Contains(err.Error(), "xdg-open") {
		t.Errorf("err = %v, want opener error to surface", err)
	}
}
