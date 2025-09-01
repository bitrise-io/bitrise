package version

// VERSION is the main CLI version number. It's defined at build time using -ldflags
var VERSION = "0.0.0-development"

// BuildNumber is the CI build number that creates the release. It's defined at build time using -ldflags
var BuildNumber = ""

// Commit is the git commit hash used for building the release. It's defined at build time using -ldflags
var Commit = ""
