package cmdutil

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/pflag"
)

// The mode resolvers below return the explicit value of a bool "mode" (set by a
// flag or its bound env var) or nil so the caller falls back to inventory-based
// detection (the isXMode functions). Env handling is uniform: a flag bound to an
// env var via SetFlagEnvVar resolves flag-first, then the env (parsed with
// strconv.ParseBool; an invalid value is an error); an empty or unset env is
// treated as absent. A flag with no env binding (e.g. --pr) resolves to the flag
// value or nil.

// ResolveBoolEnv resolves a bool from an env var: nil when unset or empty, the parsed
// value otherwise. A non-bool value is an error.
func ResolveBoolEnv(envKey string) (*bool, error) {
	raw, ok := os.LookupEnv(envKey)
	if !ok || raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q as bool value for $%s", raw, envKey)
	}
	return pointers.NewBoolPtr(parsed), nil
}

// ResolveBoolFlagOrEnv resolves a bool flag that may be bound to an env var via
// SetFlagEnvVar. Precedence: explicit flag > bound env > nil.
func ResolveBoolFlagOrEnv(fs *pflag.FlagSet, name string) (*bool, error) {
	if fs.Changed(name) {
		v, _ := fs.GetBool(name)
		return pointers.NewBoolPtr(v), nil
	}
	f := fs.Lookup(name)
	if f == nil {
		return nil, nil
	}
	if envs := f.Annotations[EnvVarAnnotation]; len(envs) > 0 {
		return ResolveBoolEnv(envs[0])
	}
	return nil, nil
}
