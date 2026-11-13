package version

// Version is the main CLI version number.
// Injected at build time via -ldflags "-X github.com/bitrise-io/stepman/version.Version=..."
// See goreleaser config file for details.
// Falls back to "dev" when not set (local / go install builds).
var Version = "dev"

// BuildNumber is the CI build number that creates the release. It's defined at build time using -ldflags
var BuildNumber = ""

// Commit is the git commit hash used for building the release. It's defined at build time using -ldflags
var Commit = ""
