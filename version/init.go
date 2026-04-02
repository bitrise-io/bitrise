package version

import (
	"runtime/debug"
	"strings"
)

// VERSION is the main CLI version number.
// Injected at build time via -ldflags "-X github.com/bitrise-io/bitrise/v2/version.VERSION=$(git describe --tags --exact-match)"
// Falls back to "dev" when not set (local / go install builds).
var VERSION = "dev"

// BuildNumber is the CI build number that creates the release.
// Injected at build time via -ldflags "-X github.com/bitrise-io/bitrise/v2/version.BuildNumber=..."
var BuildNumber = ""

// Commit is the git commit hash used for building the release.
// Injected at build time via -ldflags "-X github.com/bitrise-io/bitrise/v2/version.Commit=..."
var Commit = ""

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
