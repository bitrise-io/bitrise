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
	out, wantsNextArg, ok := walkShorthands(arg[1:], func(c string) *pflag.Flag {
		f := fs.ShorthandLookup(c)
		if f == nil || !slices.Contains(globalFlagNames, f.Name) {
			return nil
		}
		return f
	})
	switch {
	case !ok:
		return nil, 0
	case wantsNextArg && len(args) > 1:
		out[len(out)-1].value = args[1]
		return out, 2
	default:
		// A required value with nothing left to take it is pflag's error to
		// report; the empty assignment keeps the boundary scan moving.
		return out, 1
	}
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

// DetectSingleDashLongFlag reports the first argument in args that spells a
// long flag name of cmd with a single dash instead of "--" (e.g. "-config").
// Left to pflag, this either silently misparses as a shorthand cluster with
// an attached value (when the leading character happens to be a registered
// shorthand, e.g. "-config" becomes "-c" with value "onfig") or is rejected
// with a cryptic "unknown shorthand flag" error — so callers should reject it
// outright rather than let cobra parse it. cmd should be the resolved target
// command (e.g. via (*cobra.Command).Find), since a flag name is only
// meaningful relative to the command it is reachable on.
//
// Tokens that pflag would consume as a preceding flag's value are skipped:
// pflag takes the next argument verbatim even when it starts with a dash, so
// `--commit-message -tag` legitimately sets the message to "-tag" and must
// not be mistaken for a misspelled --tag.
func DetectSingleDashLongFlag(cmd *cobra.Command, args []string) (arg, flagName string, found bool) {
	byName, byShorthand := reachableFlags(cmd)

	skipValue := false
	for _, a := range args {
		if skipValue {
			// Checked before the terminator: pflag only treats "--" as one
			// when it reads it as a fresh token, so as a flag's value it is
			// taken literally and scanning has to continue past it.
			skipValue = false
			continue
		}
		if a == "--" {
			break
		}
		switch {
		case strings.HasPrefix(a, "--"):
			name, _, attached := strings.Cut(a[2:], "=")
			skipValue = !attached && takesValue(byName[name])
		case a != "-" && strings.HasPrefix(a, "-"):
			name, _, _ := strings.Cut(a[1:], "=")
			if flagName, ok := singleDashLongFlag(name, byName, byShorthand); ok {
				return a, flagName, true
			}
			_, wantsNextArg, ok := walkShorthands(a[1:], func(c string) *pflag.Flag { return byShorthand[c] })
			skipValue = ok && wantsNextArg
		}
	}
	return "", "", false
}

// singleDashLongFlag reports the long flag name token spells, after peeling any
// leading bool shorthands. pflag walks a cluster one character at a time and a
// bool consumes nothing, so "-qconfig" carries on to --config's shorthand and
// silently becomes --config=onfig — the same misparse as the bare "-config"
// this guard exists to catch, just behind an incidental bool. -h is registered
// on every command, which makes that prefix reachable everywhere.
func singleDashLongFlag(token string, byName, byShorthand map[string]*pflag.Flag) (string, bool) {
	for i := 0; i < len(token); i++ {
		if byName[token[i:]] != nil {
			return token[i:], true
		}
		// Anything but a bool shorthand ends the peel: an unregistered
		// character makes pflag reject the token, and a value-taking one
		// swallows the rest as its value.
		if f := byShorthand[token[i:i+1]]; f == nil || takesValue(f) {
			break
		}
	}
	return "", false
}

// reachableFlags indexes every flag reachable on cmd — its own plus every
// ancestor's persistent flags — by long name and by shorthand, mirroring what
// pflag actually sees once cobra merges them at parse time.
func reachableFlags(cmd *cobra.Command) (byName, byShorthand map[string]*pflag.Flag) {
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultVersionFlag()

	byName, byShorthand = map[string]*pflag.Flag{}, map[string]*pflag.Flag{}
	index := func(f *pflag.Flag) {
		byName[f.Name] = f
		if f.Shorthand != "" {
			byShorthand[f.Shorthand] = f
		}
	}
	cmd.Flags().VisitAll(index)
	cmd.InheritedFlags().VisitAll(index)
	return byName, byShorthand
}

// walkShorthands decomposes the characters after a single dash exactly as
// pflag's parseSingleShortArg does, so every caller here shares one model of
// that grammar rather than each keeping its own. In order: "-f=arg" wins but
// only when something follows the "=", then a flag with a NoOptDefVal takes
// that and parsing continues on the next character, then "-farg" swallows the
// remainder of the token, and finally a bare "-f" needs the next argument.
//
// ok is false as soon as a character names no flag, which is pflag's own hard
// error — callers decide what that means for them. wantsNextArg reports the
// last case, where the value is the following argument.
func walkShorthands(cluster string, lookup func(string) *pflag.Flag) (out []globalFlagAssignment, wantsNextArg, ok bool) {
	for cluster != "" {
		f := lookup(cluster[:1])
		if f == nil {
			return nil, false, false
		}
		switch {
		case len(cluster) > 2 && cluster[1] == '=':
			return append(out, globalFlagAssignment{f.Name, cluster[2:]}), false, true
		case f.NoOptDefVal != "":
			out = append(out, globalFlagAssignment{f.Name, f.NoOptDefVal})
			cluster = cluster[1:]
		case len(cluster) > 1:
			return append(out, globalFlagAssignment{f.Name, cluster[1:]}), false, true
		default:
			return append(out, globalFlagAssignment{f.Name, ""}), true, true
		}
	}
	return out, false, true
}

// takesValue reports whether f needs a separate argument for its value. A
// flag with a NoOptDefVal (every bool, and anything declared with one) can be
// written bare, so it never consumes the next argument.
func takesValue(f *pflag.Flag) bool {
	return f != nil && f.NoOptDefVal == ""
}
