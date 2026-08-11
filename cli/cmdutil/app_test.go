package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	assert.Empty(t, LookupAppSlug(cmd))

	t.Setenv(EnvAppID, "env-app-slug")
	assert.Equal(t, "env-app-slug", LookupAppSlug(cmd))

	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))
	assert.Equal(t, "flag-app-slug", LookupAppSlug(cmd))
}

func TestResolveAppSlug_FromConfig(t *testing.T) {
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
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")
	require.NoError(t, cmd.Flags().Set(FlagApp, "flag-app-slug"))

	slug, err := ResolveAppSlugArg(cmd, nil)
	require.NoError(t, err)
	assert.Equal(t, "flag-app-slug", slug)

	_, err = ResolveAppSlugArg(&cobra.Command{}, nil)
	require.EqualError(t, err, "--app is required")
}

func TestResolveAppSlug_LegacyAppSlugEnvNotHonored(t *testing.T) {
	// BITRISE_APP_SLUG is auto-injected by Bitrise into every build to
	// identify the app the build is running for — it must never be read
	// here, or a step running inside app X's build could silently target
	// app X's own bitrise.yml. See EnvAppID's doc comment.
	t.Setenv("BITRISE_APP_SLUG", "ci-injected-app-slug")
	cmd := &cobra.Command{}
	AddAppFlag(cmd.Flags(), "app slug")

	_, err := ResolveAppSlug(cmd)
	require.EqualError(t, err, "--app is required")
}
