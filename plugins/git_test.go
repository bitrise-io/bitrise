package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterVersionTags(t *testing.T) {
	t.Log("single version tag")
	{
		versionTags := filterVersionTags([]string{"1.0.0"})
		require.Equal(t, 1, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
	}

	t.Log("version tag list")
	{
		versionTags := filterVersionTags([]string{"1.0.0", "1.1.0", "1.1.1"})
		require.Equal(t, 3, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
		require.Equal(t, "1.1.0", versionTags[1].String())
		require.Equal(t, "1.1.1", versionTags[2].String())
	}

	t.Log("non version tag")
	{
		versionTags := filterVersionTags([]string{"release"})
		require.Equal(t, 0, len(versionTags))
	}

	t.Log("version tag + non version tag")
	{
		versionTags := filterVersionTags([]string{"1.0.0", "release"})
		require.Equal(t, 1, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
	}
}
