package rde

import (
	"fmt"
	"regexp"
	"strings"
)

// EnvVar is one NAME=VALUE pair destined for the remote command's
// environment.
type EnvVar struct {
	Name  string
	Value string
}

// envNameRe validates environment variable names before they are embedded in
// the export prefix. Anything else is rejected so a crafted "name" coming
// from a repo dotfile can never inject shell syntax into the remote command.
var envNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ResolveExecEnv resolves the env-forwarding entries for an exec: fileEntries
// from the repo dotfile (filePath names it in errors) followed by flagEntries
// from --env. Each entry is either NAME (forward the local value via
// lookupEnv) or NAME=VALUE (a literal). The two sources differ on an unset
// NAME: a dotfile entry is skipped and its name returned in skipped (a shared
// file must not break a teammate who doesn't have the var), while a flag
// entry is a hard error (the caller asked for it explicitly). A flag entry
// overrides a same-named file entry in place, so the final order is
// first-mention order with last-mention values. Values never appear in
// errors — only names.
func ResolveExecEnv(fileEntries []string, filePath string, flagEntries []string, lookupEnv func(string) (string, bool)) (vars []EnvVar, skipped []string, err error) {
	index := map[string]int{}
	set := func(name, value string) {
		if i, ok := index[name]; ok {
			vars[i].Value = value
			return
		}
		index[name] = len(vars)
		vars = append(vars, EnvVar{Name: name, Value: value})
	}

	for _, entry := range fileEntries {
		name, value, hasValue := parseEnvEntry(entry)
		if !envNameRe.MatchString(name) {
			return nil, nil, fmt.Errorf("invalid environment variable name %q in %s", name, filePath)
		}
		if !hasValue {
			local, ok := lookupEnv(name)
			if !ok {
				skipped = append(skipped, name)
				continue
			}
			value = local
		}
		set(name, value)
	}

	for _, entry := range flagEntries {
		name, value, hasValue := parseEnvEntry(entry)
		if !envNameRe.MatchString(name) {
			return nil, nil, fmt.Errorf("invalid environment variable name %q in --env flag", name)
		}
		if !hasValue {
			local, ok := lookupEnv(name)
			if !ok {
				return nil, nil, fmt.Errorf("--env %s: not set in the local environment", name)
			}
			value = local
		}
		set(name, value)
	}

	// A dotfile NAME that was unset locally but later supplied by a flag IS
	// forwarded — drop it from skipped so we don't warn about it.
	if len(skipped) > 0 {
		kept := skipped[:0]
		for _, name := range skipped {
			if _, ok := index[name]; !ok {
				kept = append(kept, name)
			}
		}
		skipped = kept
	}
	return vars, skipped, nil
}

// parseEnvEntry splits entry on the first '='. hasValue is false for a bare
// NAME entry (forward-from-local); NAME= yields an empty literal and
// A=b=c yields value "b=c".
func parseEnvEntry(entry string) (name, value string, hasValue bool) {
	return strings.Cut(entry, "=")
}

// buildEnvExportPrefix renders env as "export A='v'; export B='w'; " — the
// string prepended to the user command before buildLoginShellCmd wraps it in
// the login shell, so the exports run after profile sourcing and win over
// profile-set values. Names are re-validated here (Execute is a public entry
// point; not every caller goes through ResolveExecEnv) and values are always
// single-quoted, which keeps arbitrary bytes — including newlines and the
// interactive shell's history expansion '!' — inert. Empty env yields "".
func buildEnvExportPrefix(env []EnvVar) (string, error) {
	if len(env) == 0 {
		return "", nil
	}
	var b strings.Builder
	for _, v := range env {
		if !envNameRe.MatchString(v.Name) {
			return "", fmt.Errorf("invalid environment variable name %q", v.Name)
		}
		b.WriteString("export ")
		b.WriteString(v.Name)
		b.WriteString("=")
		b.WriteString(shellSingleQuote(v.Value))
		b.WriteString("; ")
	}
	return b.String(), nil
}

// shellSingleQuote wraps s in POSIX single quotes, escaping embedded quotes
// the same way buildLoginShellCmd does. Unlike cmdutil.ShellQuote it always
// quotes — env values may hold arbitrary bytes and never need to stay
// pretty — and it lives here because internal packages must not depend on
// the cli command layer.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
