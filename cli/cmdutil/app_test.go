package cmdutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestResolveAppSlug_FromFlag(t *testing.T) {
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "app-slug"))

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "app-slug", slug)
}

func TestResolveAppSlug_MissingFlag(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	_, err := ResolveAppSlug(cmd)
	require.EqualError(t, err, "--app is required")
}

func TestResolveAppSlug_FromEnv(t *testing.T) {
	t.Setenv(EnvAppID, "env-app-slug")
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "env-app-slug", slug)
}

func TestResolveAppSlug_FlagTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv(EnvAppID, "env-app-slug")
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "flag-app-slug", slug)
}

func TestLookupAppSlug_EmptyInsteadOfError(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	assert.Empty(t, LookupAppSlug(cmd))

	t.Setenv(EnvAppID, "env-app-slug")
	assert.Equal(t, "env-app-slug", LookupAppSlug(cmd))

	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))
	assert.Equal(t, "flag-app-slug", LookupAppSlug(cmd))
}

func TestResolveAppSlug_FromConfig(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{AppID: "config-app-slug"}}))

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "config-app-slug", slug)
}

func TestResolveAppSlug_EnvTakesPrecedenceOverConfig(t *testing.T) {
	t.Setenv(EnvAppID, "env-app-slug")
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{AppID: "config-app-slug"}}))

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "env-app-slug", slug)
}

func TestResolveAppSlugArg_PositionalArgTakesPrecedence(t *testing.T) {
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))

	slug, err := ResolveAppSlugArg(cmd, []string{"arg-app-slug"})
	require.NoError(t, err)
	assert.Equal(t, "arg-app-slug", slug)
}

func TestResolveAppSlugArg_FallsBackToResolveAppSlug(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))

	slug, err := ResolveAppSlugArg(cmd, nil)
	require.NoError(t, err)
	assert.Equal(t, "flag-app-slug", slug)

	_, err = ResolveAppSlugArg(&cobra.Command{}, nil)
	require.EqualError(t, err, "--app is required")
}

// BITRISE_APP_SLUG is auto-injected by Bitrise into every build, so honoring it
// is what lets a bare command inside a build target the app it runs for.
func TestResolveAppSlug_LegacyAppSlugEnvHonored(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "ci-injected-app-slug")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "ci-injected-app-slug", slug)
}

func TestResolveAppSlug_EnvTakesPrecedenceOverLegacyEnv(t *testing.T) {
	t.Setenv(EnvAppID, "env-app-slug")
	t.Setenv(EnvAppIDLegacy, "ci-injected-app-slug")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "env-app-slug", slug)
}

func TestResolveAppSlug_LegacyEnvTakesPrecedenceOverConfig(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "ci-injected-app-slug")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{AppID: "config-app-slug"}}))

	slug, err := ResolveAppSlug(cmd)
	require.NoError(t, err)
	assert.Equal(t, "ci-injected-app-slug", slug)
}

func TestResolveAndLookupAppSlug_ResolvesNameToSlug(t *testing.T) {
	client, calls := appsClient(t, `{"data":[{"slug":"app-123","title":"My iOS App"}],"paging":{}}`)

	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "My iOS App"))

	slug, err := ResolveAndLookupAppSlug(cmd, client)
	require.NoError(t, err)
	assert.Equal(t, "app-123", slug)
	assert.Equal(t, []string{"My iOS App"}, *calls)
}

// TestResolveAndLookupAppSlug_PassesSlugThrough covers the contract that keeps
// every existing caller working: a value the API reports no title match for is
// handed on as a literal slug rather than erroring here.
func TestResolveAndLookupAppSlug_PassesSlugThrough(t *testing.T) {
	client, _ := appsClient(t, `{"data":[],"paging":{}}`)

	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "3f8a91c02d4e"))

	slug, err := ResolveAndLookupAppSlug(cmd, client)
	require.NoError(t, err)
	assert.Equal(t, "3f8a91c02d4e", slug)
}

// TestResolveAndLookupAppSlug_AmbientSourceSkipsResolution pins that only a
// --app value is resolved by name. BITRISE_APP_ID (like BITRISE_APP_SLUG and
// the config app_id) is always already a canonical slug — Bitrise injects it
// verbatim into every build — so resolving it would just be a wasted API
// call, and the value passes through unchanged instead.
func TestResolveAndLookupAppSlug_AmbientSourceSkipsResolution(t *testing.T) {
	t.Setenv(EnvAppID, "env-app-slug")
	client, calls := appsClient(t, `{"data":[],"paging":{}}`)

	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	AddAppFlag(cmd.Flags(), "app slug")

	slug, err := ResolveAndLookupAppSlug(cmd, client)
	require.NoError(t, err)
	assert.Equal(t, "env-app-slug", slug)
	assert.Empty(t, *calls, "an ambient source must not be resolved by name")
}

func TestAppSlugIsUserProvided(t *testing.T) {
	t.Setenv(EnvAppIDLegacy, "")

	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	assert.False(t, AppSlugIsUserProvided(cmd, nil), "no flag, no arg, no ambient source")

	t.Setenv(EnvAppID, "env-app-slug")
	assert.False(t, AppSlugIsUserProvided(cmd, nil), "an ambient env var is not user-provided")

	assert.True(t, AppSlugIsUserProvided(cmd, []string{"arg-app-slug"}), "a positional arg is user-provided")

	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))
	assert.True(t, AppSlugIsUserProvided(cmd, nil), "the --app flag is user-provided")
}

func TestResolveAndLookupAppSlug_AmbiguousNameErrors(t *testing.T) {
	client, _ := appsClient(t, `{"data":[{"slug":"a1","title":"Dup"},{"slug":"a2","title":"Dup"}],"paging":{}}`)

	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "Dup"))

	_, err := ResolveAndLookupAppSlug(cmd, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

// TestNewResolver_CachesWithinOneInvocation pins the reason NewResolver builds
// a cache at all: the same name resolved twice through one resolver costs one
// API call, not two.
func TestNewResolver_CachesWithinOneInvocation(t *testing.T) {
	client, calls := appsClient(t, `{"data":[{"slug":"app-123","title":"My iOS App"}],"paging":{}}`)

	r := NewResolver(client)
	for range 2 {
		slug, err := r.AppSlug(t.Context(), "My iOS App")
		require.NoError(t, err)
		assert.Equal(t, "app-123", slug)
	}
	assert.Equal(t, []string{"My iOS App"}, *calls)
}

// appsClient returns a client whose GET /apps answers with body, plus the
// title= values it was queried with, so tests can assert on call count.
func appsClient(t *testing.T, body string) (*bitriseapi.Client, *[]string) {
	t.Helper()
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Query().Get("title"))
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	client, err := bitriseapi.New(srv.URL, "token")
	require.NoError(t, err)
	return client, &calls
}
