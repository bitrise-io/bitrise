package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// IsSSHSession reports whether the CLI is running inside an SSH session,
// where OpenBrowser would try to launch a browser on the remote host and the
// OAuth loopback redirect could never reach back to this process.
func IsSSHSession() bool {
	return os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != ""
}
