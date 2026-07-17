package cmdutil

import (
	"strings"

	"github.com/spf13/cobra"
)

// CommandTokenIndex returns the index of the first argument that is not a global
// flag — the command/plugin/positional token. Global flags before this boundary
// configure bitrise; everything from it onward belongs to the command (and, for
// plugins and envman, is forwarded verbatim), so it must not be scanned for or
// stripped of global flags.
func CommandTokenIndex(args []string, globalFlagNames []string) int {
	for i, a := range args {
		isGlobalFlag := false
		for _, name := range globalFlagNames {
			if IsFlag(name, a) {
				isGlobalFlag = true
				break
			}
		}
		if !isGlobalFlag {
			return i
		}
	}
	return len(args)
}

// ApplyGlobalFlagsFromArgs sets the global flags on the plugin/envman dispatch
// paths, where cobra does not parse them. Only the leading args (before the
// command token) are bitrise globals; anything after belongs to the passthrough.
func ApplyGlobalFlagsFromArgs(root *cobra.Command, args []string, globalFlagNames []string) {
	for _, a := range args[:CommandTokenIndex(args, globalFlagNames)] {
		for _, name := range globalFlagNames {
			switch {
			case a == "--"+name:
				_ = root.PersistentFlags().Set(name, "true")
			case strings.HasPrefix(a, "--"+name+"="):
				_ = root.PersistentFlags().Set(name, strings.TrimPrefix(a, "--"+name+"="))
			}
		}
	}
}

// IsFlag reports whether arg is the long flag --name or --name=value. Only the
// double-dash spelling is recognised, matching cobra/pflag: bitrise's long flags
// have no single-dash form.
func IsFlag(name, arg string) bool {
	return arg == "--"+name || strings.HasPrefix(arg, "--"+name+"=")
}
