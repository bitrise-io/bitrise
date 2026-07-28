package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
