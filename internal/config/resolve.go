package config

import (
	"context"
	"time"
)

// Resolved is Config, layered highest to lowest precedence:
//  1. Legacy config (~/.bitrise/config.json) — the pre-existing store, kept
//     authoritative so nothing changes for users who already have one.
//  2. Per-directory config (.bitrise-cli.yml, CWD or ancestors)
//  3. Global config file (~/.config/bitrise/cli/config.yml)
//  4. Zero value
type Resolved struct {
	SetupVersion           string
	LastCLIUpdateCheck     time.Time
	LastPluginUpdateChecks map[string]time.Time
}

// Resolve merges the legacy, per-directory, and global config layers. All
// three share the Config shape — the caller converts configs.ConfigModel
// into a Config for legacyCfg, keeping this package independent of configs.
// dirCfg / legacyCfg are zero values when their respective files were not
// found.
func Resolve(legacyCfg, dirCfg, globalCfg Config) Resolved {
	return Resolved{
		SetupVersion:           firstNonEmptyString(legacyCfg.SetupVersion, dirCfg.SetupVersion, globalCfg.SetupVersion),
		LastCLIUpdateCheck:     firstNonZeroTime(legacyCfg.LastCLIUpdateCheck, dirCfg.LastCLIUpdateCheck, globalCfg.LastCLIUpdateCheck),
		LastPluginUpdateChecks: firstNonEmptyMap(legacyCfg.LastPluginUpdateChecks, dirCfg.LastPluginUpdateChecks, globalCfg.LastPluginUpdateChecks),
	}
}

func firstNonEmptyString(values ...string) string {
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
