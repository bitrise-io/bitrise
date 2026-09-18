package claude

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

// transferActionTimeout caps the download/upload host actions. A single archive
// can take minutes to move through cloud storage (the service layer allows 10
// minutes per transfer leg), so these actions need far longer than the bridge's
// 30s default — without it serveAction would cancel a large transfer mid-flight.
const transferActionTimeout = 11 * time.Minute

// defaultRemoteUploadDir is where an upload lands in the session when the caller
// names no destination. A neutral path under /tmp (present on both Linux and
// macOS session images) rather than the session's working directory, so an
// uploaded file isn't dropped into a repo and committed by accident.
const defaultRemoteUploadDir = "/tmp/rde-claude-uploads"

// openVNCURL hands a vnc:// URL to the OS viewer. A package var so tests can
// assert what the open-vnc action would launch without opening anything.
var openVNCURL = cmdutil.OpenVNCURL

// Skill sections documenting each action in the provisioned skill. They are
// appended to the bridge's shared skill header only when their action is
// registered, so the skill never advertises a capability the session lacks.
//
//go:embed skills/open-vnc.md
var openVNCSkillSection string

//go:embed skills/download.md
var downloadSkillSection string

//go:embed skills/upload.md
var uploadSkillSection string

// newHostBridge builds the host bridge for the session with the given allowlist
// of local actions. A pure constructor — the action set (and the decision of
// which actions apply to this session) is made by the caller, so internal/rde
// never depends on the cli packages.
func newHostBridge(svc *internalrde.Service, workspaceID, sessionID string, actions map[string]internalrde.HostAction) *internalrde.HostBridge {
	return &internalrde.HostBridge{
		Service:     svc,
		WorkspaceID: workspaceID,
		SessionID:   sessionID,
		Actions:     actions,
		// Debug is left nil on purpose: the bridge runs concurrently with the
		// full-screen Claude TUI, which holds the terminal in raw mode, so any
		// diagnostic write would corrupt the display (the metadata monitor is
		// silent for the same reason).
	}
}

// localHostActions returns the host actions applicable to this session.
//
// download and upload apply to every session, so they are always registered.
// open-vnc is offered only when the session exposes a VNC endpoint — currently
// macOS sessions; Linux sessions have no VNC, so they get a bridge with just the
// transfer actions (and the in-session Claude is never told about a VNC viewer
// it can't open).
//
// localDir is the local working directory rde claude was launched from (the
// repo root). It is the base for resolving relative transfer paths.
func localHostActions(ctx context.Context, svc *internalrde.Service, workspaceID, sessionID, localDir string) map[string]internalrde.HostAction {
	actions := map[string]internalrde.HostAction{
		internalrde.ActionDownload: {
			Handle:       downloadAction(svc, workspaceID, sessionID, localDir),
			SkillSection: downloadSkillSection,
			Timeout:      transferActionTimeout,
		},
		internalrde.ActionUpload: {
			Handle:       uploadAction(svc, workspaceID, sessionID, localDir),
			SkillSection: uploadSkillSection,
			Timeout:      transferActionTimeout,
		},
	}
	if hasVNC, err := svc.SessionExposesVNC(ctx, workspaceID, sessionID); err == nil && hasVNC {
		actions[internalrde.ActionOpenVNC] = internalrde.HostAction{
			Handle:       openVNCAction(svc, workspaceID, sessionID),
			SkillSection: openVNCSkillSection,
		}
	}
	return actions
}

// hostActionsMessage describes, for the boot log, what the in-session Claude can
// do on the user's machine given the registered actions. Derived from the action
// set so it stays accurate as the set changes per session (e.g. no VNC on Linux).
func hostActionsMessage(actions map[string]internalrde.HostAction) string {
	caps := make([]string, 0, 2)
	if _, ok := actions[internalrde.ActionDownload]; ok {
		caps = append(caps, "transfer files to and from your machine")
	}
	if _, ok := actions[internalrde.ActionOpenVNC]; ok {
		caps = append(caps, "open a VNC viewer on it")
	}
	if len(caps) == 0 {
		return "Host actions enabled"
	}
	return "Host actions enabled (Claude can " + strings.Join(caps, " and ") + ")"
}

// openVNCAction fetches the session's VNC credentials and opens a viewer on the
// user's local machine pointed at the session's desktop. The VNC password is
// read locally and handed straight to the OS handler — it never travels to the
// remote.
func openVNCAction(svc *internalrde.Service, workspaceID, sessionID string) func(context.Context, *http.Request) (any, error) {
	return func(ctx context.Context, _ *http.Request) (any, error) {
		creds, err := svc.GetSessionVNC(ctx, workspaceID, sessionID)
		if err != nil {
			return nil, err
		}
		if err := openVNCURL(ctx, creds.URL); err != nil {
			return nil, fmt.Errorf("open VNC URL: %w", err)
		}
		return map[string]any{"opened": true, "address": creds.Address}, nil
	}
}

// downloadRequest is the JSON body the in-session Claude POSTs to /download.
type downloadRequest struct {
	RemotePath   string `json:"remotePath"`
	LocalDest    string `json:"localDest"`
	OnlyContents bool   `json:"onlyContents"`
}

// downloadAction pulls a file or directory from the session onto the user's
// local machine. The transfer runs from the local process (session→cloud
// storage→local), not over the bridge connection. When the caller names no
// destination the archive lands in a per-session temp dir so freshly-pulled
// files never end up inside the repo working tree.
func downloadAction(svc *internalrde.Service, workspaceID, sessionID, localDir string) func(context.Context, *http.Request) (any, error) {
	return func(ctx context.Context, r *http.Request) (any, error) {
		var req downloadRequest
		if err := decodeJSONBody(r, &req); err != nil {
			return nil, err
		}
		if req.RemotePath == "" {
			return nil, fmt.Errorf("remotePath is required")
		}
		dest := resolveDownloadDest(req.LocalDest, localDir, sessionID)
		if err := svc.DownloadFile(ctx, workspaceID, sessionID, req.RemotePath, dest, req.OnlyContents); err != nil {
			return nil, err
		}
		return map[string]any{"downloaded": true, "localPath": absOrSelf(dest)}, nil
	}
}

// uploadRequest is the JSON body the in-session Claude POSTs to /upload.
type uploadRequest struct {
	LocalPath    string `json:"localPath"`
	RemoteFolder string `json:"remoteFolder"`
}

// uploadAction pushes a local file or directory into the session. As with
// download the transfer runs from the local process (local→cloud storage→
// session). A relative localPath is resolved against the launch working dir;
// when no remoteFolder is given the file lands in a neutral temp dir on the
// session rather than its working directory.
func uploadAction(svc *internalrde.Service, workspaceID, sessionID, localDir string) func(context.Context, *http.Request) (any, error) {
	return func(ctx context.Context, r *http.Request) (any, error) {
		var req uploadRequest
		if err := decodeJSONBody(r, &req); err != nil {
			return nil, err
		}
		if req.LocalPath == "" {
			return nil, fmt.Errorf("localPath is required")
		}
		source := resolveUploadSource(req.LocalPath, localDir)
		dest := resolveRemoteUploadFolder(req.RemoteFolder)
		if err := svc.UploadFile(ctx, workspaceID, sessionID, source, dest); err != nil {
			return nil, err
		}
		return map[string]any{"uploaded": true, "localPath": absOrSelf(source), "remoteFolder": dest}, nil
	}
}

// resolveDownloadDest decides where a download's archive is extracted locally:
//   - empty       → a dedicated per-session dir under the OS temp dir, so pulled
//     files stay out of the repo and can't be committed by accident
//   - absolute    → used as-is
//   - relative    → joined onto the launch working dir
func resolveDownloadDest(localDest, localDir, sessionID string) string {
	if localDest == "" {
		return filepath.Join(os.TempDir(), "rde-claude", sessionID)
	}
	localDest = expandTilde(localDest)
	if filepath.IsAbs(localDest) {
		return localDest
	}
	return filepath.Join(localDir, localDest)
}

// resolveRemoteUploadFolder picks the session-side destination for an upload:
// the caller's remoteFolder when given, otherwise the neutral temp dir so the
// file doesn't land in the session's working directory.
func resolveRemoteUploadFolder(remoteFolder string) string {
	if remoteFolder == "" {
		return defaultRemoteUploadDir
	}
	return remoteFolder
}

// resolveUploadSource resolves the local file/dir to upload: absolute paths are
// used as-is, relative ones are joined onto the launch working dir.
func resolveUploadSource(localPath, localDir string) string {
	localPath = expandTilde(localPath)
	if filepath.IsAbs(localPath) {
		return localPath
	}
	return filepath.Join(localDir, localPath)
}

// expandTilde replaces a leading ~ (alone or as ~/...) with the local user's
// home directory. Paths reach us as JSON, not through a shell, so a literal ~
// would otherwise be treated as a relative path component — surprising when the
// user says "download it to ~/Downloads".
func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// absOrSelf returns the absolute form of path, falling back to path itself if it
// can't be resolved. Used so the result reported back to Claude (and the user)
// is an unambiguous path.
func absOrSelf(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// decodeJSONBody decodes the request body into dst. An empty body is not an
// error — dst is left at its zero values so the caller's own required-field
// checks produce the clearer message.
func decodeJSONBody(r *http.Request, dst any) error {
	if r.Body == nil {
		return nil
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("parse request body: %w", err)
	}
	return nil
}
