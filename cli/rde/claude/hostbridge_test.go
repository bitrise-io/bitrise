package claude

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// sessionResponse is an httptest handler returning a running session whose VNC
// address is vncAddress (empty mimics a Linux session with no VNC endpoint).
func sessionResponse(vncAddress string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"vncAddress":"`+vncAddress+`","vncUsername":"vagrant","vncPassword":"hunter2"
		}}`)
	}
}

func newTestService(t *testing.T, h http.Handler) *internalrde.Service {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	client, err := rdeapi.New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("rdeapi.New: %v", err)
	}
	return internalrde.NewService(client)
}

func TestLocalHostActions_OffersOpenVNCWhenSessionHasVNC(t *testing.T) {
	svc := newTestService(t, sessionResponse("vnc://host.example:5900"))
	actions := localHostActions(context.Background(), svc, "ws-1", "s-1", "/repo")
	// Every action must ship a skill section that references its own route —
	// otherwise the composed skill would document an endpoint that doesn't exist
	// (or omit one that does).
	for _, name := range []string{internalrde.ActionOpenVNC, internalrde.ActionDownload, internalrde.ActionUpload} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("expected %q action for a VNC-capable session", name)
		}
		if !strings.Contains(action.SkillSection, name) {
			t.Errorf("%q skill section does not mention its own route", name)
		}
	}
}

func TestLocalHostActions_SkipsOpenVNCWhenNoVNC(t *testing.T) {
	// A Linux session has no VNC endpoint, so open-vnc must not be offered — but
	// download and upload work on every session and are always present.
	svc := newTestService(t, sessionResponse(""))
	actions := localHostActions(context.Background(), svc, "ws-1", "s-1", "/repo")
	if _, ok := actions[internalrde.ActionOpenVNC]; ok {
		t.Error("open-vnc must not be offered on a session without VNC")
	}
	for _, name := range []string{internalrde.ActionDownload, internalrde.ActionUpload} {
		if _, ok := actions[name]; !ok {
			t.Errorf("%q must be offered on every session", name)
		}
	}
}

func TestOpenVNCActionOpensViewer(t *testing.T) {
	// The action fetches the session's VNC credentials and hands the vnc:// URL
	// to the opener. Stub the opener to capture the URL without launching a
	// viewer, and back the service with a session that exposes VNC.
	svc := newTestService(t, sessionResponse("vnc://host.example:5900"))

	var gotURL string
	prev := openVNCURL
	openVNCURL = func(_ context.Context, url string) error { gotURL = url; return nil }
	defer func() { openVNCURL = prev }()

	action := openVNCAction(svc, "ws-1", "s-1")
	res, err := action(context.Background(), httptest.NewRequest(http.MethodPost, "/open-vnc", nil))
	if err != nil {
		t.Fatalf("action: %v", err)
	}

	want := "vnc://vagrant:hunter2@host.example:5900" // #nosec G101 -- test fixture
	if gotURL != want {
		t.Errorf("opened URL = %q, want %q", gotURL, want)
	}
	m, ok := res.(map[string]any)
	if !ok || m["opened"] != true || m["address"] != "vnc://host.example:5900" {
		t.Errorf("result = %v, want opened=true with address", res)
	}
}

func TestResolveDownloadDest(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		name      string
		localDest string
		want      string
	}{
		{"empty defaults to per-session temp dir", "", filepath.Join(os.TempDir(), "rde-claude", "s-1")},
		{"absolute used as-is", "/var/out", "/var/out"},
		{"relative joined onto launch dir", "out/logs", filepath.Join("/repo", "out", "logs")},
		{"tilde expands to home", "~/Downloads", filepath.Join(home, "Downloads")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.localDest == "~/Downloads" && home == "" {
				t.Skip("no home dir in test env")
			}
			if got := resolveDownloadDest(tt.localDest, "/repo", "s-1"); got != tt.want {
				t.Errorf("resolveDownloadDest(%q) = %q, want %q", tt.localDest, got, tt.want)
			}
		})
	}
}

func TestResolveUploadSource(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		name      string
		localPath string
		want      string
	}{
		{"absolute used as-is", "/etc/cert.p12", "/etc/cert.p12"},
		{"relative joined onto launch dir", "build/out.tar", filepath.Join("/repo", "build", "out.tar")},
		{"tilde expands to home", "~/cert.p12", filepath.Join(home, "cert.p12")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.localPath, "~/") && home == "" {
				t.Skip("no home dir in test env")
			}
			if got := resolveUploadSource(tt.localPath, "/repo"); got != tt.want {
				t.Errorf("resolveUploadSource(%q) = %q, want %q", tt.localPath, got, tt.want)
			}
		})
	}
}

// failBackend returns a service whose backend fails the test if reached — used
// to prove a validation error short-circuits before any transfer is attempted.
func failBackend(t *testing.T) *internalrde.Service {
	t.Helper()
	return newTestService(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("backend must not be called on a validation error; got %s", r.URL.Path)
	}))
}

func TestDownloadActionRequiresRemotePath(t *testing.T) {
	action := downloadAction(failBackend(t), "ws-1", "s-1", "/repo")
	// Empty body → no remotePath → clear error, no backend call.
	_, err := action(context.Background(), httptest.NewRequest(http.MethodPost, "/download", nil))
	if err == nil || !strings.Contains(err.Error(), "remotePath is required") {
		t.Fatalf("err = %v, want remotePath required", err)
	}
}

func TestUploadActionRequiresLocalPath(t *testing.T) {
	action := uploadAction(failBackend(t), "ws-1", "s-1", "/repo")
	for _, body := range []string{`{"remoteFolder":"/tmp"}`, ``} {
		req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body))
		_, err := action(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "localPath is required") {
			t.Fatalf("body %q: err = %v, want localPath required", body, err)
		}
	}
}

func TestResolveRemoteUploadFolder(t *testing.T) {
	// Omitted destination defaults to the neutral temp dir, not the session's
	// working directory; an explicit folder is used as-is.
	if got := resolveRemoteUploadFolder(""); got != defaultRemoteUploadDir {
		t.Errorf("empty remoteFolder = %q, want %q", got, defaultRemoteUploadDir)
	}
	if got := resolveRemoteUploadFolder("/home/ubuntu/project"); got != "/home/ubuntu/project" {
		t.Errorf("explicit remoteFolder = %q, want it unchanged", got)
	}
}

func TestHostActionsMessage(t *testing.T) {
	transfers := map[string]internalrde.HostAction{
		internalrde.ActionDownload: {},
		internalrde.ActionUpload:   {},
	}
	if got := hostActionsMessage(transfers); got != "Host actions enabled (Claude can transfer files to and from your machine)" {
		t.Errorf("transfers-only message = %q", got)
	}
	withVNC := map[string]internalrde.HostAction{
		internalrde.ActionDownload: {},
		internalrde.ActionUpload:   {},
		internalrde.ActionOpenVNC:  {},
	}
	got := hostActionsMessage(withVNC)
	if !strings.Contains(got, "transfer files") || !strings.Contains(got, "VNC viewer") {
		t.Errorf("with-VNC message = %q, want both capabilities", got)
	}
}
