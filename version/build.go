package version

//todo: Remove this version number change
// VERSION is the main CLI version number. It's defined at build time using -ldflags
var VERSION = "1.99.0-development-json-test"

// BuildNumber is the CI build number that creates the release. It's defined at build time using -ldflags
var BuildNumber = ""

// Commit is the git commit hash used for building the release. It's defined at build time using -ldflags
var Commit = ""
