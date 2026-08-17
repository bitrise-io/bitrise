package config

import (
	"context"
	"time"
)

// Resolved is a Config merged from, highest to lowest precedence:
//  1. Legacy config (~/.bitrise/config.json) — the pre-existing store, kept
//     authoritative so nothing changes for users who already have one.
//  2. Per-directory config (.bitrise-cli.yml, CWD or ancestors)
//  3. Global config file (~/.config/bitrise/cli/config.yml)
//  4. Zero value
//
// APIBaseURL/WebBaseURL are the exception: they skip the per-directory layer
// (see below) — everything else follows the order above.
//
// Resolved embeds Config (rather than being an identical, separately-typed
// copy of its fields) so a new field only needs adding once — but it stays a
// distinct type on purpose: Save takes a Config, and a Resolved carries
// values from all three layers, so passing one straight to Save would risk
// writing per-dir- or legacy-only data into a file that should only ever
// reflect what was actually written to it.
type Resolved struct {
	Config
}

// DefaultAPIBaseURL is the production Bitrise API base URL, used when no
// layer sets api_base_url.
const DefaultAPIBaseURL = "https://api.bitrise.io/v0.1"

// DefaultWebBaseURL is the production Bitrise website base URL, used when no
// layer sets web_base_url.
const DefaultWebBaseURL = "https://app.bitrise.io"

// Resolve merges the legacy, per-directory, and global config layers. The
// caller converts configs.ConfigModel into a Config for legacyCfg, keeping
// this package independent of configs. dirCfg / legacyCfg are zero values
// when their respective files were not found.
//
// APIBaseURL/WebBaseURL deliberately never consult dirCfg: both carry
// credentials (a bearer token, a login password) to whatever host they name,
// and .bitrise-cli.yml is read from the current directory and every
// ancestor with no confirmation — a repo a user merely clones and runs
// `bitrise` inside of could otherwise silently redirect either one to an
// attacker-controlled host. The global file and the legacy file are both
// user-owned (home directory), not repo-owned, so they don't have this
// problem.
func Resolve(legacyCfg, dirCfg, globalCfg Config) Resolved {
	return Resolved{Config: Config{
		SetupVersion:           FirstNonEmptyString(legacyCfg.SetupVersion, dirCfg.SetupVersion, globalCfg.SetupVersion),
		LastCLIUpdateCheck:     firstNonZeroTime(legacyCfg.LastCLIUpdateCheck, dirCfg.LastCLIUpdateCheck, globalCfg.LastCLIUpdateCheck),
		LastPluginUpdateChecks: firstNonEmptyMap(legacyCfg.LastPluginUpdateChecks, dirCfg.LastPluginUpdateChecks, globalCfg.LastPluginUpdateChecks),
		// legacy is always empty for these two (configs.ConfigModel predates
		// the cloud API and has no such field), so this is effectively
		// global > default — no dirCfg, see the doc comment above.
		APIBaseURL: FirstNonEmptyString(legacyCfg.APIBaseURL, globalCfg.APIBaseURL, DefaultAPIBaseURL),
		WebBaseURL: FirstNonEmptyString(legacyCfg.WebBaseURL, globalCfg.WebBaseURL, DefaultWebBaseURL),
		// legacyCfg.AppID is likewise always empty, and there's no sensible
		// default app — effectively dir > global > unset. Both of these DO
		// honor dirCfg, unlike the two URLs above: they name which app and
		// workspace a checkout belongs to, which is exactly what a repo-local
		// file should be able to pin, and neither is a credential.
		AppID:              FirstNonEmptyString(legacyCfg.AppID, dirCfg.AppID, globalCfg.AppID),
		DefaultWorkspaceID: FirstNonEmptyString(legacyCfg.DefaultWorkspaceID, dirCfg.DefaultWorkspaceID, globalCfg.DefaultWorkspaceID),
		// Output/Theme are neither credentials nor URLs, so — like AppID/
		// DefaultWorkspaceID above — they honor dirCfg: a repo may reasonably
		// pin its own output format or color theme.
		Output: FirstNonEmptyString(legacyCfg.Output, dirCfg.Output, globalCfg.Output),
		Theme:  FirstNonEmptyString(legacyCfg.Theme, dirCfg.Theme, globalCfg.Theme),
	}}
}

// FirstNonEmptyString returns the first non-empty value, or "" if all are empty.
func FirstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, v := range values {
		if !v.IsZero() {
			return v
		}
	}
	return time.Time{}
}

// firstNonEmptyMap returns the first non-empty map wholesale — layers don't
// merge per-plugin entries, the higher-precedence layer's map wins entirely.
func firstNonEmptyMap(values ...map[string]time.Time) map[string]time.Time {
	for _, v := range values {
		if len(v) > 0 {
			return v
		}
	}
	return nil
}

type ctxKey struct{}

// WithResolved stores r on ctx so command handlers can read it.
func WithResolved(ctx context.Context, r Resolved) context.Context {
	return context.WithValue(ctx, ctxKey{}, r)
}

// FromContext retrieves Resolved from ctx, or a zero value if absent. A nil
// ctx is treated as absent rather than panicking: cmd.Context() is nil for a
// bare *cobra.Command that was never executed, which is how several tests
// drive RunE directly.
func FromContext(ctx context.Context) Resolved {
	if ctx == nil {
		return Resolved{}
	}
	if r, ok := ctx.Value(ctxKey{}).(Resolved); ok {
		return r
	}
	return Resolved{}
}
