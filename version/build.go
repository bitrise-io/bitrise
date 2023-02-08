package version

import (
	"fmt"
	"runtime/debug"
)

// Version is the main CLI version number. It's defined at build time using -ldflags
var Version = ""

// BuildNumber is the CI build number that creates the release. It's defined at build time using -ldflags
var BuildNumber = ""

// Commit is the git commit hash used for building the release. It's defined at build time using -ldflags
var Commit = ""

func init() {
	if Version != "" {
		return
	}

	info, available := debug.ReadBuildInfo()
	if available && info.Main.Version != "" && info.Main.Sum != "" {
		Version = info.Main.Version
		Commit = fmt.Sprintf("%q (mod checksum)", info.Main.Sum)
	} else {
		Version = "local"
	}
}
