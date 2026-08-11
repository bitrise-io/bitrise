package cmdutil

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// FlagApp is the app slug a command acts on.
const FlagApp = "app"

// EnvAppID overrides the app slug when --app isn't passed. Deliberately does
// NOT also accept BITRISE_APP_SLUG the way the reference CLI does: Bitrise
// auto-injects that variable into every build to identify the app the build
// is running for (see analytics/tracker.go, configs/agent_config.go), so
// honoring it here would make a bare `bitrise yml update` step running
// inside app X's build silently target and overwrite app X's own
// bitrise.yml.
const EnvAppID = "BITRISE_APP_ID"

// AddAppFlag registers --app. Registered per-subcommand rather than as a
// persistent parent flag, since some of these commands are also
// re-registered standalone as legacy top-level aliases (see cli/root.go)
// that never attach to the yml parent.
func AddAppFlag(fs *pflag.FlagSet, help string) {
	fs.String(FlagApp, "", help)
}

// ResolveAppSlug returns the app slug from --app, falling back to
// BITRISE_APP_ID, then the app_id set by `bitrise app create` or
// `bitrise config set app_id`.
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
	if v, _ := cmd.Flags().GetString(FlagApp); v != "" {
		return v
	}
	if v := os.Getenv(EnvAppID); v != "" {
		return v
	}
	return config.FromContext(cmd.Context()).AppID
}

// AppSlugRequiredErr returns the standard missing-app-slug error.
func AppSlugRequiredErr() error {
	return errors.New("--app is required")
}
