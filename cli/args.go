package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandTokenIndex returns the index of the first argument that is not a global
// flag — the command/plugin/positional token. Global flags before this boundary
// configure bitrise; everything from it onward belongs to the command (and, for
// plugins and envman, is forwarded verbatim), so it must not be scanned for or
// stripped of global flags.
func commandTokenIndex(args []string) int {
	for i, a := range args {
		if !isGlobalFlagArg(a) {
			return i
		}
	}
	return len(args)
}

// applyGlobalFlagsFromArgs sets the global flags on the plugin/envman dispatch
// paths, where cobra does not parse them. Only the leading args (before the
// command token) are bitrise globals; anything after belongs to the passthrough.
func applyGlobalFlagsFromArgs(root *cobra.Command, args []string) {
	for _, a := range args[:commandTokenIndex(args)] {
		for _, name := range globalFlagNames {
			switch {
			case a == "--"+name || a == "-"+name:
				_ = root.PersistentFlags().Set(name, "true")
			case strings.HasPrefix(a, "--"+name+"="):
				_ = root.PersistentFlags().Set(name, strings.TrimPrefix(a, "--"+name+"="))
			case strings.HasPrefix(a, "-"+name+"="):
				_ = root.PersistentFlags().Set(name, strings.TrimPrefix(a, "-"+name+"="))
			}
		}
	}
}

func isGlobalFlagArg(a string) bool {
	for _, name := range globalFlagNames {
		if isFlag(name, a) {
			return true
		}
	}
	return false
}

func isKnownCommand(root *cobra.Command, name string) bool {
	if name == "help" {
		return true
	}
	for _, c := range root.Commands() {
		if c.Name() == name {
			return true
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}

// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
// normalizeLegacyArgs rewrites single-dash long flags (e.g. `-config`) to their
// double-dash form, so both spellings are accepted: urfave/Go-flag took `-config`
// ≡ `--config`, but pflag accepts only `--config` / `-c`. The next major can drop
// this (with its normalizeArg / knownLongFlagNames helpers) and accept only the
// double-dash form. Passthrough args are left untouched: everything after `--` and
// everything from the envman command onwards (envman forwards its args verbatim
// and uses single-dash long flags of its own).
func normalizeLegacyArgs(args []string, root *cobra.Command) []string {
	known := knownLongFlagNames(root)
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" || a == "envman" {
			out = append(out, args[i:]...)
			break
		}
		out = append(out, normalizeArg(a, known))
	}
	return out
}

func normalizeArg(arg string, known map[string]bool) string {
	if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
		return arg
	}
	name := strings.TrimPrefix(arg, "-")
	if eq := strings.IndexByte(name, '='); eq >= 0 {
		name = name[:eq]
	}
	if len(name) >= 2 && known[name] {
		return "-" + arg
	}
	return arg
}

func knownLongFlagNames(root *cobra.Command) map[string]bool {
	names := map[string]bool{"help": true, "version": true}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		c.PersistentFlags().VisitAll(func(f *pflag.Flag) { names[f.Name] = true })
		c.Flags().VisitAll(func(f *pflag.Flag) { names[f.Name] = true })
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	return names
}
