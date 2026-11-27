package tools

import (
	"path/filepath"

	envman "github.com/bitrise-io/bitrise/v3/envman/cli"
)

// shims are the tools the CLI stands in for when it is invoked under their name,
// through the links setup creates in the shims dir.
var shims = []struct {
	name string
	run  func()
}{
	{EnvmanToolName, envman.Run},
}

// RunAsShim runs the built-in copy of the tool argv0 names, if there is one.
func RunAsShim(argv0 string) bool {
	name := filepath.Base(argv0)
	for _, shim := range shims {
		if shim.name == name {
			shim.run()
			return true
		}
	}
	return false
}

// ShimmedTools lists the tools setup links to the CLI.
func ShimmedTools() []string {
	names := make([]string, 0, len(shims))
	for _, shim := range shims {
		names = append(names, shim.name)
	}
	return names
}
