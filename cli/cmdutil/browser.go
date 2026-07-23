package cmdutil

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url) //nolint:gosec // G204: fixed command, URL from API
	case "linux":
		cmd = exec.Command("xdg-open", url) //nolint:gosec // G204: fixed command, URL from API
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec // G204: fixed command, URL from API
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
