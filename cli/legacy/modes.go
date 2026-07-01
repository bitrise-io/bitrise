package legacy

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

// CIModeFlagOverride resolves the --ci flag bound to the CI env var: the flag
// value if passed, otherwise the parsed env value if present, otherwise nil.
func CIModeFlagOverride(cmd *cobra.Command, ciFlagName string) *bool {
	if override := flagBoolOverride(cmd.Root().PersistentFlags(), ciFlagName); override != nil {
		return override
	}
	if val, ok := os.LookupEnv(configs.CIModeEnvKey); ok {
		parsed, _ := strconv.ParseBool(val)
		return pointers.NewBoolPtr(parsed)
	}
	return nil
}

// PRModeFlagOverride resolves the --pr flag. Unlike CI, the PR env vars (PR mode
// and pull request ID) are interpreted by isPRMode together with the inventory,
// so only the flag is resolved here.
func PRModeFlagOverride(cmd *cobra.Command, prFlagName string) *bool {
	return flagBoolOverride(cmd.Root().PersistentFlags(), prFlagName)
}

// SecretFilteringFlagOverride resolves trigger's --secret-filtering flag, which
// was bound to the BITRISE_SECRET_FILTERING env var: the flag value if passed,
// otherwise the parsed env value if present (an empty value is treated as false),
// otherwise nil. A non-bool env value returns an error, matching the previous
// framework's abort. run's flag was not env-bound — see RunSecretFilteringOverride.
func SecretFilteringFlagOverride(cmd *cobra.Command, flagName string) (*bool, error) {
	// The env var was bound to the flag and validated when the flag set was
	// built, before the command-line value was applied, so a non-bool env value
	// aborts even when --secret-filtering is also passed.
	envVal, envSet := os.LookupEnv(configs.IsSecretFilteringKey)
	if envSet && envVal != "" {
		if _, err := strconv.ParseBool(envVal); err != nil {
			return nil, fmt.Errorf("could not parse %q as bool value for $%s", envVal, configs.IsSecretFilteringKey)
		}
	}

	if override := flagBoolOverride(cmd.Flags(), flagName); override != nil {
		return override, nil
	}
	if !envSet {
		return nil, nil
	}
	if envVal == "" {
		return pointers.NewBoolPtr(false), nil
	}
	parsed, _ := strconv.ParseBool(envVal)
	return pointers.NewBoolPtr(parsed), nil
}

// RunSecretFilteringOverride resolves run's --secret-filtering flag, which (unlike
// trigger's) was never bound to the BITRISE_SECRET_FILTERING env var: the flag
// value if passed, otherwise the env matched literally, otherwise nil. A non-bool
// env value is ignored (not aborted), matching the pre-cobra behaviour.
func RunSecretFilteringOverride(cmd *cobra.Command, flagName string) *bool {
	if override := flagBoolOverride(cmd.Flags(), flagName); override != nil {
		return override
	}
	return literalBoolEnv(configs.IsSecretFilteringKey)
}

// SecretEnvsFilteringOverride resolves the secret-envs-filtering mode, which has
// no flag and was never env-bound: the env matched literally, otherwise nil so
// detection falls back to the inventory.
func SecretEnvsFilteringOverride() *bool {
	return literalBoolEnv(configs.IsSecretEnvsFilteringKey)
}

// ValidateGlobalBoolEnvs returns an error when a global bool flag's bound
// environment variable holds a non-bool value, matching the behaviour of the
// previous framework (an empty value is allowed and treated as false).
func ValidateGlobalBoolEnvs() error {
	for _, envKey := range []string{configs.CIModeEnvKey, configs.DebugModeEnvKey} {
		if val, ok := os.LookupEnv(envKey); ok && val != "" {
			if _, err := strconv.ParseBool(val); err != nil {
				return fmt.Errorf("could not parse %q as bool value for $%s", val, envKey)
			}
		}
	}
	return nil
}

// IsDebugMode resolves debug mode before cobra parses the command line, so the
// logger can be configured up front. An explicit truthy --debug enables it;
// otherwise the bound DEBUG env var decides, so a set DEBUG=true still enables
// debug mode even alongside an explicit --debug=false.
//
// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
// The correct precedence is "explicit flag wins" (matching --ci/--pr), but the
// pre-cobra CLI let a set DEBUG=true env override an explicit --debug=false
// (loggerParameters resolved the flag, then fell through to the env unless the
// flag was truthy). Reproduced here so debug mode toggles identically; the next
// major can make the flag win.
//
// The arg scan is intentionally unbounded (no command-token boundary), so a
// plugin's or envman's own trailing --debug also toggles this early logger —
// again matching the pre-cobra CLI. See debugFlagFromArgs.
func IsDebugMode(arguments []string, debugFlagName string) bool {
	if value, set := debugFlagFromArgs(arguments, debugFlagName); set && value {
		return true
	}
	return os.Getenv(configs.DebugModeEnvKey) == "true"
}

// debugFlagFromArgs extracts the --debug global flag from the raw args, ahead of
// cobra, returning set=false when it is absent or carries a non-bool value (in
// which case cobra later reports the error). cobra re-parses the same flag for
// help, analytics and the command itself; this early pass only feeds the logger.
func debugFlagFromArgs(arguments []string, debugFlagName string) (value bool, set bool) {
	for _, argument := range arguments {
		if !IsFlag(debugFlagName, argument) {
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

func IsFlag(name, arg string) bool {
	return arg == "--"+name || arg == "-"+name ||
		strings.HasPrefix(arg, "--"+name+"=") || strings.HasPrefix(arg, "-"+name+"=")
}
