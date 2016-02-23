package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Log("simple plugin command")
	{
		args := []string{"bitrise", ":example"}
		pluginName, pluginArgs, isPlugin := ParseArgs(args)
		require.Equal(t, true, isPlugin)
		require.Equal(t, "example", pluginName)
		require.Equal(t, 0, len(pluginArgs))
	}

	t.Log("simple plugin command - with bitrise flags")
	{
		args := []string{"bitrise", "-l", "debug", ":example"}
		pluginName, pluginArgs, isPlugin := ParseArgs(args)
		require.Equal(t, true, isPlugin)
		require.Equal(t, "example", pluginName)
		require.Equal(t, 0, len(pluginArgs))
	}

	t.Log("plugin command - with args")
	{
		args := []string{"bitrise", ":example", "hello", "bitrise"}
		pluginName, pluginArgs, isPlugin := ParseArgs(args)
		require.Equal(t, true, isPlugin)
		require.Equal(t, "example", pluginName)
		require.EqualValues(t, []string{"hello", "bitrise"}, pluginArgs)
	}

	t.Log("plugin command - with falg")
	{
		args := []string{"bitrise", ":example", "hello", "--name", "bitrise"}
		pluginName, pluginArgs, isPlugin := ParseArgs(args)
		require.Equal(t, true, isPlugin)
		require.Equal(t, "example", pluginName)
		require.EqualValues(t, []string{"hello", "--name", "bitrise"}, pluginArgs)
	}

	t.Log("not plugin command")
	{
		args := []string{"bitrise", "hello", "bitrise"}
		pluginName, pluginArgs, isPlugin := ParseArgs(args)
		require.Equal(t, false, isPlugin)
		require.Equal(t, "", pluginName)
		require.Equal(t, 0, len(pluginArgs))
	}
}
