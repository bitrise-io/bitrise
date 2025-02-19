package version

import (
	"runtime/debug"
	"strings"
)

// IsAlternativeInstallation tracks if the current cli was installed via `go install` or not.
var IsAlternativeInstallation = false

func init() {
	if Commit != "" {
		return
	}

	info, available := debug.ReadBuildInfo()
	if available && info.Main.Version != "" {
		IsAlternativeInstallation = true

		components := strings.Split(info.Main.Version, "-")
		count := len(components)
		if count != 0 {
			Commit = components[count-1]
		}
	}
}
