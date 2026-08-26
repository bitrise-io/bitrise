package cmdutil

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandTokenIndex returns the index of the first argument that is not a global
// flag — the command/plugin/positional token. Global flags before this boundary
// configure bitrise; everything from it onward belongs to the command (and, for
// plugins and envman, is forwarded verbatim), so it must not be scanned for or
// stripped of global flags.
//
// fs is the root's persistent flag set. It supplies each global's shorthand and
// whether it takes a value, so this scanner accepts the same spellings cobra
// would: a bare "--output"/"-o" is followed by its value, not by the
// command/plugin token, and that value must be skipped rather than mistaken
// for it.
func CommandTokenIndex(fs *pflag.FlagSet, args []string, globalFlagNames []string) int {
	i := 0
	for i < len(args) {
		_, consumed := matchGlobalFlags(fs, args[i:], globalFlagNames)
		if consumed == 0 {
			return i
		}
		i += consumed
	}
	return len(args)
}

// ApplyGlobalFlagsFromArgs sets the global flags on the plugin/envman dispatch
// paths, where cobra does not parse them. Only the leading args (before the
// command token) are bitrise globals; anything after belongs to the passthrough.
func ApplyGlobalFlagsFromArgs(root *cobra.Command, args []string, globalFlagNames []string) {
	fs := root.PersistentFlags()
	boundary := CommandTokenIndex(fs, args, globalFlagNames)
	for i := 0; i < boundary; {
		assignments, consumed := matchGlobalFlags(fs, args[i:boundary], globalFlagNames)
		if consumed == 0 {
			break // unreachable within the boundary, but keeps the loop well-defined
		}
		for _, a := range assignments {
			_ = fs.Set(a.name, a.value)
		}
		i += consumed
	}
}

type globalFlagAssignment struct{ name, value string }

// matchGlobalFlags reports the assignments args[0] makes — plus, for a value
// flag written with a space, args[1] — and how many leading tokens of args that
// consumed. consumed is 0 when args[0] does not name a global flag at all, in
// which case the assignments are meaningless. A shorthand cluster ("-qo json")
// assigns more than one flag, hence the slice.
func matchGlobalFlags(fs *pflag.FlagSet, args, globalFlagNames []string) ([]globalFlagAssignment, int) {
	arg := args[0]

	if strings.HasPrefix(arg, "--") {
		name, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
		if !slices.Contains(globalFlagNames, name) {
			return nil, 0
		}
		f := fs.Lookup(name)
		if f == nil {
			return nil, 0
		}
		return assignOne(f, value, hasValue, args)
	}

	if !strings.HasPrefix(arg, "-") || arg == "-" {
		return nil, 0
	}

	// Shorthand form. Every character has to name a global flag, since a
	// cluster is all-or-nothing: "-qx" is not two bitrise globals, so the
	// whole token belongs to the command.
	rest := strings.TrimPrefix(arg, "-")
	var out []globalFlagAssignment
	for i := 0; i < len(rest); i++ {
		f := fs.ShorthandLookup(string(rest[i]))
		if f == nil || !slices.Contains(globalFlagNames, f.Name) {
			return nil, 0
		}
		tail := rest[i+1:]
		if f.NoOptDefVal == "" {
			// A value flag ends the cluster: it swallows the rest of the token
			// ("-o=json", "-ojson") or, when the token ends here, the next
			// argument.
			value, hasValue := strings.CutPrefix(tail, "=")
			if !hasValue {
				value, hasValue = tail, tail != ""
			}
			assigned, consumed := assignOne(f, value, hasValue, args)
			return append(out, assigned...), consumed
		}
		if value, hasValue := strings.CutPrefix(tail, "="); hasValue {
			return append(out, globalFlagAssignment{f.Name, value}), 1
		}
		out = append(out, globalFlagAssignment{f.Name, f.NoOptDefVal})
	}
	return out, 1
}

// assignOne resolves a single flag's value: an attached one when hasValue, the
// flag's no-argument default when it has one (bools), otherwise the following
// argument.
func assignOne(f *pflag.Flag, value string, hasValue bool, args []string) ([]globalFlagAssignment, int) {
	switch {
	case hasValue:
		return []globalFlagAssignment{{f.Name, value}}, 1
	case f.NoOptDefVal != "":
		return []globalFlagAssignment{{f.Name, f.NoOptDefVal}}, 1
	case len(args) > 1:
		return []globalFlagAssignment{{f.Name, args[1]}}, 2
	default:
		return []globalFlagAssignment{{f.Name, ""}}, 1
	}
}

// IsFlag reports whether arg is the long flag --name or --name=value. Only the
// double-dash spelling is recognised, matching cobra/pflag: bitrise's long flags
// have no single-dash form.
func IsFlag(name, arg string) bool {
	return arg == "--"+name || strings.HasPrefix(arg, "--"+name+"=")
}
