package config

import (
	"context"
	"time"

	"github.com/bitrise-io/bitrise/v2/configs"
)

// Resolved is Config, layered highest to lowest precedence:
//  1. Per-directory config (.bitrise-cli.yml, CWD or ancestors)
//  2. Global config file (~/.config/bitrise/config.yaml)
//  3. Legacy config (~/.bitrise/config.json) — the pre-existing store, kept
//     as the fallback so nothing regresses for users who've never seen the
//     new files.
//  4. Zero value
type Resolved struct {
	SetupVersion           string
	LastCLIUpdateCheck     time.Time
	LastPluginUpdateChecks map[string]time.Time
}

// Resolve merges the global config, per-directory config, and the legacy
// ~/.bitrise/config.json store. dirCfg / legacyCfg are zero values when their
// respective files were not found.
func Resolve(globalCfg, dirCfg Config, legacyCfg configs.ConfigModel) Resolved {
	return Resolved{
		SetupVersion:           firstNonEmptyString(dirCfg.SetupVersion, globalCfg.SetupVersion, legacyCfg.SetupVersion),
		LastCLIUpdateCheck:     firstNonZeroTime(dirCfg.LastCLIUpdateCheck, globalCfg.LastCLIUpdateCheck, legacyCfg.LastCLIUpdateCheck),
		LastPluginUpdateChecks: firstNonEmptyMap(dirCfg.LastPluginUpdateChecks, globalCfg.LastPluginUpdateChecks, legacyCfg.LastPluginUpdateChecks),
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
