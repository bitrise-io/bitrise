package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolve_DefaultsWhenNothingSet(t *testing.T) {
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, Resolved{Config: Config{APIBaseURL: DefaultAPIBaseURL}}, r)
}

func TestResolve_APIBaseURLPrecedence(t *testing.T) {
	dir := Config{APIBaseURL: "https://dir.example"}
	global := Config{APIBaseURL: "https://global.example"}

	// no layer set: falls back to the default
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, DefaultAPIBaseURL, r.APIBaseURL)

	// global only
	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "https://global.example", r.APIBaseURL)

	// dir overrides global (legacy has no concept of this field)
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "https://dir.example", r.APIBaseURL)
}

func TestResolve_SetupVersionPrecedence(t *testing.T) {
	legacy := Config{SetupVersion: "legacy"}
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
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "dir", r.SetupVersion)
}

func TestResolve_LastCLIUpdateCheckPrecedence(t *testing.T) {
	legacyTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dirTime := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	globalTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	legacy := Config{LastCLIUpdateCheck: legacyTime}
	dir := Config{LastCLIUpdateCheck: dirTime}
	global := Config{LastCLIUpdateCheck: globalTime}

	r := Resolve(legacy, Config{}, Config{})
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacyTime))

	// legacy still wins over dir and global
	r = Resolve(legacy, dir, global)
	assert.True(t, r.LastCLIUpdateCheck.Equal(legacyTime))

	// without legacy, dir overrides global
	r = Resolve(Config{}, dir, global)
	assert.True(t, r.LastCLIUpdateCheck.Equal(dirTime))
}

func TestResolve_LastPluginUpdateChecksPrecedence(t *testing.T) {
	legacy := Config{LastPluginUpdateChecks: map[string]time.Time{"legacy-plugin": time.Now()}}
	global := Config{LastPluginUpdateChecks: map[string]time.Time{"global-plugin": time.Now()}}

	r := Resolve(legacy, Config{}, Config{})
	assert.Contains(t, r.LastPluginUpdateChecks, "legacy-plugin")

	// legacy still wins over dir and global
	r = Resolve(legacy, dir, global)
	assert.Contains(t, r.LastPluginUpdateChecks, "legacy-plugin")

	// without legacy, dir overrides global
	r = Resolve(Config{}, dir, global)
	assert.Contains(t, r.LastPluginUpdateChecks, "dir-plugin")
}

// TestResolve_LegacyTakesPrecedence asserts the legacy config wins over the
// new per-dir/global layers even when they're also set.
func TestResolve_LegacyTakesPrecedence(t *testing.T) {
	legacy := Config{
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
	r := Resolved{Config: Config{SetupVersion: "abc"}}
	ctx := WithResolved(t.Context(), r)
	got := FromContext(ctx)
	assert.Equal(t, r, got)
}

func TestFromContext_ZeroWhenAbsent(t *testing.T) {
	got := FromContext(t.Context())
	assert.Equal(t, Resolved{}, got)
}
