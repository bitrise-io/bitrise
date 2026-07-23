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

// Resolve merges the legacy, per-directory, and global config layers. The
// caller converts configs.ConfigModel into a Config for legacyCfg, keeping
// this package independent of configs. dirCfg / legacyCfg are zero values
// when their respective files were not found.
func Resolve(legacyCfg, dirCfg, globalCfg Config) Resolved {
	return Resolved{Config: Config{
		SetupVersion:           FirstNonEmptyString(legacyCfg.SetupVersion, dirCfg.SetupVersion, globalCfg.SetupVersion),
		LastCLIUpdateCheck:     firstNonZeroTime(legacyCfg.LastCLIUpdateCheck, dirCfg.LastCLIUpdateCheck, globalCfg.LastCLIUpdateCheck),
		LastPluginUpdateChecks: firstNonEmptyMap(legacyCfg.LastPluginUpdateChecks, dirCfg.LastPluginUpdateChecks, globalCfg.LastPluginUpdateChecks),
		// legacyCfg.APIBaseURL is always empty (configs.ConfigModel predates
		// the cloud API and has no such field), so this is effectively
		// dir > global > default.
		APIBaseURL: FirstNonEmptyString(legacyCfg.APIBaseURL, dirCfg.APIBaseURL, globalCfg.APIBaseURL, DefaultAPIBaseURL),
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

// FromContext retrieves Resolved from ctx, or a zero value if absent.
func FromContext(ctx context.Context) Resolved {
	if r, ok := ctx.Value(ctxKey{}).(Resolved); ok {
		return r
	}
	return Resolved{}
}
