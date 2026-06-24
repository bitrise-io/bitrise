package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
// These overrides hand-roll the env-var bindings of the pre-cobra (urfave) CLI
// (pflag has no EnvVar equivalent), resolving the flag/env part of each bool mode
// and returning nil when neither is set so the caller falls back to inventory-based
// detection (the isXMode functions). Reading these envs is a kept feature; the
// intentionally inconsistent per-mode semantics are the compat burden the next
// major can unify. Do not unify them now:
//   - --ci was bound to the CI env var (urfave EnvVar): the env is parsed with
//     ParseBool and its mere presence overrides the inventory.
//   - --pr had no env binding: only the flag is resolved here; the PR env vars are
//     interpreted by isPRMode together with the inventory.
//   - trigger's --secret-filtering was bound to BITRISE_SECRET_FILTERING: parsed
//     with ParseBool, a non-bool value aborts.
//   - run's --secret-filtering and the (flag-less) secret-envs-filtering mode were
//     never env-bound: their env vars are matched literally ("true"/"false") with
//     no ParseBool and no abort.

// flagBoolOverride returns the flag's value when it was set on the command line,
// otherwise nil.
func flagBoolOverride(fs *pflag.FlagSet, name string) *bool {
	if !fs.Changed(name) {
		return nil
	}
	v, _ := fs.GetBool(name)
	return pointers.NewBoolPtr(v)
}

// literalBoolEnv resolves a non-env-bound mode's env var: a pointer to true/false
// only when the value is exactly "true"/"false", otherwise nil. It never parses
// other truthy spellings and never aborts, matching the pre-cobra behaviour of
// the env vars that had no urfave EnvVar binding.
func literalBoolEnv(envKey string) *bool {
	switch os.Getenv(envKey) {
	case "true":
		return pointers.NewBoolPtr(true)
	case "false":
		return pointers.NewBoolPtr(false)
	}
	return nil
}

// ciModeFlagOverride resolves the --ci flag bound to the CI env var: the flag
// value if passed, otherwise the parsed env value if present, otherwise nil.
func ciModeFlagOverride(cmd *cobra.Command) *bool {
	if override := flagBoolOverride(cmd.Root().PersistentFlags(), CIKey); override != nil {
		return override
	}
	if val, ok := os.LookupEnv(configs.CIModeEnvKey); ok {
		parsed, _ := strconv.ParseBool(val)
		return pointers.NewBoolPtr(parsed)
	}
	return nil
}

// prModeFlagOverride resolves the --pr flag. Unlike CI, the PR env vars (PR mode
// and pull request ID) are interpreted by isPRMode together with the inventory,
// so only the flag is resolved here.
func prModeFlagOverride(cmd *cobra.Command) *bool {
	return flagBoolOverride(cmd.Root().PersistentFlags(), PRKey)
}

// secretFilteringFlagOverride resolves trigger's --secret-filtering flag, which
// was bound to the BITRISE_SECRET_FILTERING env var: the flag value if passed,
// otherwise the parsed env value if present (an empty value is treated as false),
// otherwise nil. A non-bool env value aborts, matching the previous framework.
// run's flag was not env-bound — see runSecretFilteringOverride.
func secretFilteringFlagOverride(cmd *cobra.Command) *bool {
	// The env var was bound to the flag and validated when the flag set was
	// built, before the command-line value was applied, so a non-bool env value
	// aborts even when --secret-filtering is also passed.
	envVal, envSet := os.LookupEnv(configs.IsSecretFilteringKey)
	if envSet && envVal != "" {
		if _, err := strconv.ParseBool(envVal); err != nil {
			failf("could not parse %q as bool value for $%s", envVal, configs.IsSecretFilteringKey)
		}
	}

	if override := flagBoolOverride(cmd.Flags(), secretFilteringFlag); override != nil {
		return override
	}
	if !envSet {
		return nil
	}
	if envVal == "" {
		return pointers.NewBoolPtr(false)
	}
	parsed, _ := strconv.ParseBool(envVal)
	return pointers.NewBoolPtr(parsed)
}

// runSecretFilteringOverride resolves run's --secret-filtering flag, which (unlike
// trigger's) was never bound to the BITRISE_SECRET_FILTERING env var: the flag
// value if passed, otherwise the env matched literally, otherwise nil. A non-bool
// env value is ignored (not aborted), matching the pre-cobra behaviour.
func runSecretFilteringOverride(cmd *cobra.Command) *bool {
	if override := flagBoolOverride(cmd.Flags(), secretFilteringFlag); override != nil {
		return override
	}
	return literalBoolEnv(configs.IsSecretFilteringKey)
}

// secretEnvsFilteringOverride resolves the secret-envs-filtering mode, which has
// no flag and was never env-bound: the env matched literally, otherwise nil so
// detection falls back to the inventory.
func secretEnvsFilteringOverride() *bool {
	return literalBoolEnv(configs.IsSecretEnvsFilteringKey)
}

// isDebugMode resolves debug mode before cobra parses the command line, so the
// logger can be configured up front. An explicit --debug flag wins (matching
// the --ci/--pr precedence); otherwise the bound DEBUG env var decides.
func isDebugMode(arguments []string) bool {
	if value, set := debugFlagFromArgs(arguments); set {
		return value
	}
	return os.Getenv(configs.DebugModeEnvKey) == "true"
}

// debugFlagFromArgs extracts the --debug global flag from the raw args, ahead of
// cobra, returning set=false when it is absent or carries a non-bool value (in
// which case cobra later reports the error). cobra re-parses the same flag for
// help, analytics and the command itself; this early pass only feeds the logger.
func debugFlagFromArgs(arguments []string) (value bool, set bool) {
	for _, argument := range arguments {
		if !isFlag(DebugModeKey, argument) {
			continue
		}
		// "-flag x" syntax is not supported for boolean flags, so only the bare
		// flag and the "-flag=x" form are accepted.
		if _, raw, ok := strings.Cut(argument, "="); ok {
			if parsed, err := strconv.ParseBool(raw); err == nil {
				value, set = parsed, true
			}
		} else {
			value, set = true, true
		}
	}
	return
}

func isFlag(name, arg string) bool {
	return arg == "--"+name || arg == "-"+name ||
		strings.HasPrefix(arg, "--"+name+"=") || strings.HasPrefix(arg, "-"+name+"=")
}
