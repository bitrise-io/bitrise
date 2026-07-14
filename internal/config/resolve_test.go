package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/configs"
)

func TestResolve_DefaultsWhenNothingSet(t *testing.T) {
	r := Resolve(Config{}, Config{}, configs.ConfigModel{})
	assert.Equal(t, Resolved{}, r)
}

func TestResolve_SetupVersionPrecedence(t *testing.T) {
	legacy := configs.ConfigModel{SetupVersion: "legacy"}
	global := Config{SetupVersion: "global"}
	dir := Config{SetupVersion: "dir"}

	// legacy only
	r := Resolve(Config{}, Config{}, legacy)
	assert.Equal(t, "legacy", r.SetupVersion)

	// global overrides legacy
	r = Resolve(global, Config{}, legacy)
	assert.Equal(t, "global", r.SetupVersion)

	// dir overrides global and legacy
	r = Resolve(global, dir, legacy)
	assert.Equal(t, "dir", r.SetupVersion)
}

func TestResolve_LastCLIUpdateCheckPrecedence(t *testing.T) {
	legacyTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	globalTime := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	dirTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	legacy := configs.ConfigModel{LastCLIUpdateCheck: legacyTime}
	global := Config{LastCLIUpdateCheck: globalTime}
	dir := Config{LastCLIUpdateCheck: dirTime}

	r := Resolve(Config{}, Config{}, legacy)
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacyTime))

	r = Resolve(global, Config{}, legacy)
	assert.True(t, r.LastCLIUpdateCheck.Equal(globalTime))

	r = Resolve(global, dir, legacy)
	assert.True(t, r.LastCLIUpdateCheck.Equal(dirTime))
}

func TestResolve_LastPluginUpdateChecksPrecedence(t *testing.T) {
	legacy := configs.ConfigModel{LastPluginUpdateChecks: map[string]time.Time{"legacy-plugin": time.Now()}}
	global := Config{LastPluginUpdateChecks: map[string]time.Time{"global-plugin": time.Now()}}
	dir := Config{LastPluginUpdateChecks: map[string]time.Time{"dir-plugin": time.Now()}}

	r := Resolve(Config{}, Config{}, legacy)
	assert.Contains(t, r.LastPluginUpdateChecks, "legacy-plugin")

	r = Resolve(global, Config{}, legacy)
	assert.Contains(t, r.LastPluginUpdateChecks, "global-plugin")

	r = Resolve(global, dir, legacy)
	assert.Contains(t, r.LastPluginUpdateChecks, "dir-plugin")
}

// TestResolve_LegacyIsTheFallback is the concrete behavior this task is
// about: with no new-layer files present, the pre-existing
// ~/.bitrise/config.json contents surface through Resolved unchanged.
func TestResolve_LegacyIsTheFallback(t *testing.T) {
	legacy := configs.ConfigModel{
		SetupVersion:           "9.9.9",
		LastCLIUpdateCheck:     time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"init": time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)},
	}
	r := Resolve(Config{}, Config{}, legacy)
	assert.Equal(t, legacy.SetupVersion, r.SetupVersion)
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacy.LastCLIUpdateCheck))
	assert.Equal(t, legacy.LastPluginUpdateChecks, r.LastPluginUpdateChecks)
}

func TestContext_RoundTrip(t *testing.T) {
	r := Resolved{SetupVersion: "abc"}
	ctx := WithResolved(t.Context(), r)
	got := FromContext(ctx)
	assert.Equal(t, r, got)
}

func TestFromContext_ZeroWhenAbsent(t *testing.T) {
	got := FromContext(t.Context())
	assert.Equal(t, Resolved{}, got)
}
