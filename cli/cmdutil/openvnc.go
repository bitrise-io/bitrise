package cmdutil

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OpenVNCURL invokes the OS's default URL handler for a vnc:// URL:
//
//   - macOS:    open
//   - Windows:  cmd /c start
//   - other:    xdg-open
//
// Errors carry the handler's combined output so a missing xdg-open or a
// malformed URL is debuggable without rerunning under strace.
//
// The vnc:// prefix is verified before shelling out. The URL is built from
// backend-provided host/port plus URL-escaped credentials, but the call can be
// triggered by a remote signal (the rde claude host bridge), so the prefix
// check is the trust boundary that keeps a non-vnc argument from reaching the
// OS handler.
func OpenVNCURL(ctx context.Context, url string) error {
	if !strings.HasPrefix(url, "vnc://") {
		return fmt.Errorf("refusing to open non-vnc URL %q", url)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "windows":
		// `cmd /c start "" url` — the empty "" is `start`'s window-title
		// placeholder, without it `start` treats the URL as the title.
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", url)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%w: %s", err, string(out))
		}
		return err
	}
	return nil
}
