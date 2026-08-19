// Package legacy holds the pre-cobra (urfave → cobra) migration compatibility
// shims for the bitrise CLI: legacy argument normalization, dispatch parsing and
// the hand-rolled flag/env "mode" resolution. They are isolated here, away from
// the command wiring in package cli, so the compatibility surface is easy to spot
// and the next major version can drop the package wholesale. It depends only on
// cobra/pflag and the configs package — never on cli.
package legacy

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandTokenIndex returns the index of the first argument that is not a global
// flag — the command/plugin/positional token. Global flags before this boundary
// configure bitrise; everything from it onward belongs to the command (and, for
// plugins and envman, is forwarded verbatim), so it must not be scanned for or
// stripped of global flags.
func CommandTokenIndex(args []string, globalFlagNames []string) int {
	for i, a := range args {
		if !isGlobalFlagArg(a, globalFlagNames) {
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

func isGlobalFlagArg(a string, globalFlagNames []string) bool {
	for _, name := range globalFlagNames {
		if IsFlag(name, a) {
			return true
		}
	}
	return false
}

func IsKnownCommand(root *cobra.Command, name string) bool {
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

// NormalizeLegacyArgs rewrites single-dash long flags (e.g. `-config`) to their
// double-dash form, so both spellings are accepted: urfave/Go-flag took `-config`
// ≡ `--config`, but pflag accepts only `--config` / `-c`. The next major can drop
// this (with its normalizeArg / knownLongFlagNames helpers) and accept only the
// double-dash form. Passthrough args are left untouched: everything after `--` and
// everything from the envman command onwards (envman forwards its args verbatim
// and uses single-dash long flags of its own).
func NormalizeLegacyArgs(args []string, root *cobra.Command) []string {
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

// EnableUnknownFlagPassthrough sets FParseErrWhitelist.UnknownFlags on the whole
// command tree. urfave/cli left an unrecognised flag that followed a positional
// argument in the argument list and ignored it (e.g. `bitrise run wf --unknown`
// still ran the workflow); pflag rejects unknown flags outright. The flag is
// per-command (cobra does not inherit it), and the command that ultimately runs is
// the one that parses, so it must be set on every command, not just the root. The
// next major version, which reworks the command surface, can tighten this.
func EnableUnknownFlagPassthrough(cmd *cobra.Command) {
	cmd.FParseErrWhitelist.UnknownFlags = true
	for _, sub := range cmd.Commands() {
		EnableUnknownFlagPassthrough(sub)
	}
}
