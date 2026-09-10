package cmdutil

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/cache"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/internal/resolve"
)

// FlagApp is the app slug a command acts on.
const FlagApp = "app"

// EnvAppID overrides the app slug when --app isn't passed.
const EnvAppID = "BITRISE_APP_ID"

// EnvAppIDLegacy is the pre-rename name, still accepted below EnvAppID.
// Bitrise auto-injects it into every build to identify the app the build runs
// for (see analytics/tracker.go, configs/agent_config.go), so honoring it
// means a bare command inside app X's build targets app X — including
// `bitrise yml update`, which then overwrites that app's own bitrise.yml.
// That ambient targeting is intentional and matches the released CLI.
const EnvAppIDLegacy = "BITRISE_APP_SLUG"

// AddAppFlag registers --app. Registered per-subcommand rather than as a
// persistent parent flag, since some of these commands are also
// re-registered standalone as legacy top-level aliases (see cli/root.go)
// that never attach to the yml parent.
func AddAppFlag(fs *pflag.FlagSet, help string) {
	fs.String(FlagApp, "", help)
}

// flagAppSlug returns just the --app flag's value, or "" if unset.
func flagAppSlug(cmd *cobra.Command) string {
	v, _ := cmd.Flags().GetString(FlagApp)
	return v
}

// ambientAppSlug returns the app slug from BITRISE_APP_ID, then
// BITRISE_APP_SLUG, then the app_id set by `bitrise app create` or `bitrise
// config set app_id` — all of which are always already a canonical slug
// (Bitrise-injected or previously resolved by this CLI), never a display
// name a user typed, so callers never resolve these by name.
func ambientAppSlug(cmd *cobra.Command) string {
	if v := os.Getenv(EnvAppID); v != "" {
		return v
	}
	if v := os.Getenv(EnvAppIDLegacy); v != "" {
		return v
	}
	return config.FromContext(cmd.Context()).AppID
}

// ResolveAppSlug returns the app slug from --app, falling back to
// BITRISE_APP_ID, then BITRISE_APP_SLUG, then the app_id set by
// `bitrise app create` or `bitrise config set app_id`.
func ResolveAppSlug(cmd *cobra.Command) (string, error) {
	if slug := LookupAppSlug(cmd); slug != "" {
		return slug, nil
	}
	return "", AppSlugRequiredErr()
}

// ResolveAppSlugArg is ResolveAppSlug for commands that also accept the app
// slug as a positional argument (e.g. `app view [APP_ID]`), which takes
// precedence over --app/BITRISE_APP_ID/config when given.
func ResolveAppSlugArg(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	return ResolveAppSlug(cmd)
}

// LookupAppSlug resolves the app slug the same way as ResolveAppSlug but
// returns an empty string instead of an error when neither source is set —
// for commands where the app is optional (e.g. `yml validate --app`).
func LookupAppSlug(cmd *cobra.Command) string {
	if v := flagAppSlug(cmd); v != "" {
		return v
	}
	return ambientAppSlug(cmd)
}

// AppSlugIsUserProvided reports whether the app slug ResolveAppSlug /
// ResolveAppSlugArg / LookupAppSlug would return for cmd/args came from
// something the user typed (--app, or a positional argument in args) rather
// than an ambient source (BITRISE_APP_ID/BITRISE_APP_SLUG, saved config).
// Only a user-provided value can be a display name — ambient ones are always
// already a canonical slug — so callers that resolve a name to a slug
// themselves (rather than through ResolveAndLookupAppSlug) use this to skip
// that lookup, and its API call, for ambient values.
func AppSlugIsUserProvided(cmd *cobra.Command, args []string) bool {
	if len(args) > 0 && args[0] != "" {
		return true
	}
	return flagAppSlug(cmd) != ""
}

// AppSlugRequiredErr returns the standard missing-app-slug error.
func AppSlugRequiredErr() error {
	return errors.New("--app is required")
}

// NewResolver returns a Resolver wired to client and a fresh in-memory cache,
// so repeated lookups within one command invocation hit the API once.
func NewResolver(client *bitriseapi.Client) *resolve.Resolver {
	return resolve.New(client, cache.New())
}

// ResolveAndLookupAppSlug reads the app slug from --app, falling back to
// BITRISE_APP_ID, then BITRISE_APP_SLUG, then config app_id. Only a --app
// value goes through the name lookup (a targeted GET /apps?title=<value>
// query) — the ambient sources are already a canonical slug, so resolving
// those would just be a wasted API call.
func ResolveAndLookupAppSlug(cmd *cobra.Command, client *bitriseapi.Client) (string, error) {
	if v := flagAppSlug(cmd); v != "" {
		return NewResolver(client).AppSlug(cmd.Context(), v)
	}
	if v := ambientAppSlug(cmd); v != "" {
		return v, nil
	}
	return "", AppSlugRequiredErr()
}
