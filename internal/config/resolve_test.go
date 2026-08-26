package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolve_DefaultsWhenNothingSet(t *testing.T) {
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, Resolved{Config: Config{APIBaseURL: DefaultAPIBaseURL, WebBaseURL: DefaultWebBaseURL}}, r)
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

	// dir is ignored, even when set and global isn't: this field carries a
	// bearer token, and a repo's own .bitrise-cli.yml must not be able to
	// redirect it (see Resolve's doc comment)
	r = Resolve(Config{}, dir, Config{})
	assert.Equal(t, DefaultAPIBaseURL, r.APIBaseURL, "dirCfg must never win for APIBaseURL")

	// dir also doesn't override global
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "https://global.example", r.APIBaseURL)
}

func TestResolve_WebBaseURLPrecedence(t *testing.T) {
	dir := Config{WebBaseURL: "https://dir.example"}
	global := Config{WebBaseURL: "https://global.example"}

	// no layer set: falls back to the default
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, DefaultWebBaseURL, r.WebBaseURL)

	// global only
	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "https://global.example", r.WebBaseURL)

	// dir is ignored, even when set and global isn't: this field carries a
	// login password, and a repo's own .bitrise-cli.yml must not be able to
	// redirect it (see Resolve's doc comment)
	r = Resolve(Config{}, dir, Config{})
	assert.Equal(t, DefaultWebBaseURL, r.WebBaseURL, "dirCfg must never win for WebBaseURL")

	// dir also doesn't override global
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "https://global.example", r.WebBaseURL)
}

func TestResolve_AppIDPrecedence(t *testing.T) {
	dir := Config{AppID: "dir-app"}
	global := Config{AppID: "global-app"}

	// no layer set: stays unset, unlike APIBaseURL there's no default
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, "", r.AppID)

	// global only
	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "global-app", r.AppID)

	// dir overrides global (legacy has no concept of this field)
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "dir-app", r.AppID)
}

func TestResolve_DefaultWorkspaceIDPrecedence(t *testing.T) {
	dir := Config{DefaultWorkspaceID: "dir-ws"}
	global := Config{DefaultWorkspaceID: "global-ws"}

	// no layer set: stays unset, there's no default workspace
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, "", r.DefaultWorkspaceID)

	// global only
	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "global-ws", r.DefaultWorkspaceID)

	// dir overrides global — a repo may pin the workspace it belongs to
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "dir-ws", r.DefaultWorkspaceID)
}

func TestResolve_OutputPrecedence(t *testing.T) {
	dir := Config{Output: "json"}
	global := Config{Output: "yml"}

	// no layer set: stays unset — the raw/auto default is applied downstream
	// by output.ParseFormat/style.ParseTheme, not here
	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, "", r.Output)

	// global only
	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "yml", r.Output)

	// dir overrides global — not a credential or a URL, same treatment as AppID
	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "json", r.Output)
}

func TestResolve_ThemePrecedence(t *testing.T) {
	dir := Config{Theme: "dark"}
	global := Config{Theme: "light"}

	r := Resolve(Config{}, Config{}, Config{})
	assert.Equal(t, "", r.Theme)

	r = Resolve(Config{}, Config{}, global)
	assert.Equal(t, "light", r.Theme)

	r = Resolve(Config{}, dir, global)
	assert.Equal(t, "dark", r.Theme)
}

// TestResolve_LayerPrecedence covers the three fields sharing the generic
// legacy > dir > global precedence (SetupVersion, LastCLIUpdateCheck,
// LastPluginUpdateChecks) in one pass, each via a different underlying
// firstNonEmpty* helper (string/time/map) — one test per precedence
// scenario, not one per field, since the precedence logic itself is
// identical across all three.
func TestResolve_LayerPrecedence(t *testing.T) {
	legacy := Config{
		SetupVersion:           "legacy",
		LastCLIUpdateCheck:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"legacy-plugin": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	dir := Config{
		SetupVersion:           "dir",
		LastCLIUpdateCheck:     time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"dir-plugin": time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	global := Config{
		SetupVersion:           "global",
		LastCLIUpdateCheck:     time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		LastPluginUpdateChecks: map[string]time.Time{"global-plugin": time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	assertResolvedAs := func(t *testing.T, want Config, r Resolved) {
		t.Helper()
		assert.Equal(t, want.SetupVersion, r.SetupVersion)
		assert.True(t, r.LastCLIUpdateCheck.Equal(want.LastCLIUpdateCheck))
		assert.Equal(t, want.LastPluginUpdateChecks, r.LastPluginUpdateChecks)
	}

	t.Run("legacy wins over dir and global", func(t *testing.T) {
		assertResolvedAs(t, legacy, Resolve(legacy, dir, global))
	})
	t.Run("dir wins when legacy absent", func(t *testing.T) {
		assertResolvedAs(t, dir, Resolve(Config{}, dir, global))
	})
	t.Run("global used when legacy and dir absent", func(t *testing.T) {
		assertResolvedAs(t, global, Resolve(Config{}, Config{}, global))
	})
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
