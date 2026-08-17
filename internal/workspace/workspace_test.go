package workspace

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestSole(t *testing.T) {
	t.Run("exactly one is returned", func(t *testing.T) {
		got, err := Sole([]bitriseapi.Organization{{Slug: "only-ws", Name: "Only"}})
		require.NoError(t, err)
		assert.Equal(t, "only-ws", got.Slug)
	})

	t.Run("none is an error naming the way out", func(t *testing.T) {
		_, err := Sole(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no workspaces found")
		assert.Contains(t, err.Error(), "--workspace")
	})

	t.Run("several lists them sorted, with both ways to pin one", func(t *testing.T) {
		_, err := Sole([]bitriseapi.Organization{
			{Slug: "b-ws", Name: "Bravo"},
			{Slug: "a-ws", Name: "Alpha"},
		})
		require.Error(t, err)
		msg := err.Error()
		assert.Contains(t, msg, "multiple workspaces available")
		assert.Contains(t, msg, EnvWorkspaceID)
		assert.Contains(t, msg, "bitrise config set default_workspace_id")
		assert.Less(t, strings.Index(msg, "Alpha"), strings.Index(msg, "Bravo"))
	})
}

func TestSort_NamedFirstThenCaseInsensitiveByName(t *testing.T) {
	orgs := []bitriseapi.Organization{
		{Slug: "z-ws"},
		{Slug: "c-ws", Name: "charlie"},
		{Slug: "a-ws"},
		{Slug: "b-ws", Name: "Bravo"},
	}
	got := Sort(orgs)

	var slugs []string
	for _, o := range got {
		slugs = append(slugs, o.Slug)
	}
	assert.Equal(t, []string{"b-ws", "c-ws", "a-ws", "z-ws"}, slugs)
	assert.Equal(t, "z-ws", orgs[0].Slug, "Sort must not reorder its argument")
}

func TestList_RendersNameSlugAndFallsBackToSlug(t *testing.T) {
	got := List([]bitriseapi.Organization{
		{Slug: "nameless"},
		{Slug: "a-ws", Name: "Alpha"},
	})
	assert.Equal(t, "  Alpha (a-ws)\n  nameless", got)
}
