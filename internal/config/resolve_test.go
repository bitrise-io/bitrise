package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/configs"
)

func TestResolve_DefaultsWhenNothingSet(t *testing.T) {
	r := Resolve(configs.ConfigModel{}, Config{}, Config{})
	assert.Equal(t, Resolved{}, r)
}

func TestResolve_SetupVersionPrecedence(t *testing.T) {
	legacy := configs.ConfigModel{SetupVersion: "legacy"}
	dir := Config{SetupVersion: "dir"}
	global := Config{SetupVersion: "global"}

	// legacy only
	r := Resolve(legacy, Config{}, Config{})
	assert.Equal(t, "legacy", r.SetupVersion)

	// legacy still wins over dir
	r = Resolve(legacy, dir, Config{})
	assert.Equal(t, "legacy", r.SetupVersion)

	// legacy still wins over dir and global
	r = Resolve(legacy, dir, global)
	assert.Equal(t, "legacy", r.SetupVersion)

	// without legacy, dir overrides global
	r = Resolve(configs.ConfigModel{}, dir, global)
	assert.Equal(t, "dir", r.SetupVersion)
}

func TestResolve_LastCLIUpdateCheckPrecedence(t *testing.T) {
	legacyTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dirTime := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	globalTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	legacy := configs.ConfigModel{LastCLIUpdateCheck: legacyTime}
	dir := Config{LastCLIUpdateCheck: dirTime}
	global := Config{LastCLIUpdateCheck: globalTime}

	r := Resolve(legacy, Config{}, Config{})
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacyTime))

	// legacy still wins over dir and global
	r = Resolve(legacy, dir, global)
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacyTime))

	// without legacy, dir overrides global
	r = Resolve(configs.ConfigModel{}, dir, global)
	assert.True(t, r.LastCLIUpdateCheck.Equal(dirTime))
}

func TestResolve_LastPluginUpdateChecksPrecedence(t *testing.T) {
	legacy := configs.ConfigModel{LastPluginUpdateChecks: map[string]time.Time{"legacy-plugin": time.Now()}}
	dir := Config{LastPluginUpdateChecks: map[string]time.Time{"dir-plugin": time.Now()}}
	global := Config{LastPluginUpdateChecks: map[string]time.Time{"global-plugin": time.Now()}}

	r := Resolve(legacy, Config{}, Config{})
	assert.Contains(t, r.LastPluginUpdateChecks, "legacy-plugin")

	// legacy still wins over dir and global
	r = Resolve(legacy, dir, global)
	assert.Contains(t, r.LastPluginUpdateChecks, "legacy-plugin")

	// without legacy, dir overrides global
	r = Resolve(configs.ConfigModel{}, dir, global)
	assert.Contains(t, r.LastPluginUpdateChecks, "dir-plugin")
}

// TestResolve_LegacyTakesPrecedence is the concrete behavior this task is
// about: per the RFC, the pre-existing ~/.bitrise/config.json is read first
// and wins over the new per-dir/global layers, even when they're also set.
func TestResolve_LegacyTakesPrecedence(t *testing.T) {
	legacy := configs.ConfigModel{
		SetupVersion:           "9.9.9",
		LastCLIUpdateCheck:     time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"init": time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)},
	}
	dir := Config{
		SetupVersion:           "1.0.0",
		LastCLIUpdateCheck:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"dir-plugin": time.Now()},
	}
	global := Config{
		SetupVersion:           "2.0.0",
		LastCLIUpdateCheck:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"global-plugin": time.Now()},
	}
	r := Resolve(legacy, dir, global)
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
