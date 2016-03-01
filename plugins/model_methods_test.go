package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPluginFromBytes(t *testing.T) {
	t.Log("simple plugin")
	{
		pluginStr := `name: Name
description: Description
requirements:
- tool: Tool
  min_version: 1.0.0
  max_version: 1.0.2
`

		plugin, err := NewPluginFromBytes([]byte(pluginStr))
		require.NoError(t, err)
		require.Equal(t, "Name", plugin.Name)
		require.Equal(t, "Description", plugin.Description)
		require.Equal(t, 1, len(plugin.Requirements))

		requirement := plugin.Requirements[0]
		require.Equal(t, "Tool", requirement.Tool)
		require.Equal(t, "1.0.0", requirement.MinVersion)
		require.Equal(t, "1.0.2", requirement.MaxVersion)
	}
}

func TestValidate(t *testing.T) {
	t.Log("invalid plugin - no name")
	{
		pluginStr := `name: ""
description: Description
requirements:
- tool: Tool
  min_version: 1.0.0
  max_version: 1.0.2
`

		_, err := NewPluginFromBytes([]byte(pluginStr))
		require.Error(t, err)
	}
}

func TestString(t *testing.T) {
	t.Log("simple plugin")
	{
		pluginStr := `name: Name
description: Description
requirements:
- tool: Tool
  min_version: 1.0.0
  max_version: 1.0.2
`

		plugin, err := NewPluginFromBytes([]byte(pluginStr))
		require.NoError(t, err)

		desiredPrintablePlugin := "\x1b[32;1mName\x1b[0m\n  Description: Description"
		printablePlugin := plugin.String()
		require.Equal(t, desiredPrintablePlugin, printablePlugin)
	}
}

func TestSortByName(t *testing.T) {
	t.Log("single plugin")
	{
		pluginA := Plugin{Name: "A"}

		plugins := []Plugin{pluginA}

		SortByName(plugins)
		require.Equal(t, "A", plugins[0].Name)
	}

	t.Log("simple sort")
	{
		pluginA := Plugin{Name: "A"}
		pluginB := Plugin{Name: "B"}
		pluginC := Plugin{Name: "C"}

		plugins := []Plugin{pluginC, pluginA, pluginB}

		SortByName(plugins)
		require.Equal(t, "A", plugins[0].Name)
		require.Equal(t, "B", plugins[1].Name)
		require.Equal(t, "C", plugins[2].Name)
	}
}

func TestNewPluginRoutingFromBytes(t *testing.T) {
	t.Log("simple routing")
	{
		routingStr := `route_map:
  name:
    name: name
    source: source
    version: "1.0.0"
    commit_hash: hash
    executable: "./test"
`

		routing, err := NewPluginRoutingFromBytes([]byte(routingStr))
		require.NoError(t, err)

		route, found := routing.RouteMap["name"]
		require.Equal(t, true, found)
		require.Equal(t, "name", route.Name)
		require.Equal(t, "source", route.Source)
		require.Equal(t, "1.0.0", route.Version)
		require.Equal(t, "hash", route.CommitHash)
		require.Equal(t, "./test", route.Executable)
	}
}

func TestValidateRouting(t *testing.T) {
	t.Log("simple routing")
	{
		routing := PluginRouting{
			RouteMap: map[string]PluginRoute{
				"test": PluginRoute{
					Name:       "test",
					Source:     "source",
					Version:    "1.0.0",
					CommitHash: "hash",
					Executable: "./executable",
				},
			},
		}

		require.NoError(t, routing.Validate())
	}

	t.Log("invalid routing - missing required route's key")
	{
		routing := PluginRouting{
			RouteMap: map[string]PluginRoute{
				"": PluginRoute{
					Name:       "test",
					Source:     "source",
					Version:    "1.0.0",
					CommitHash: "hash",
					Executable: "./executable",
				},
			},
		}

		require.Error(t, routing.Validate())
	}

	t.Log("invalid routing - route's key, route's name missmatch")
	{
		routing := PluginRouting{
			RouteMap: map[string]PluginRoute{
				"test1": PluginRoute{
					Name:       "test2",
					Source:     "source",
					Version:    "1.0.0",
					CommitHash: "hash",
					Executable: "./executable",
				},
			},
		}

		require.Error(t, routing.Validate())
	}
}

/*
// AddRoute ...
func (routing *PluginRouting) AddRoute(route PluginRoute) {
	routing.RouteMap[route.Name] = route
}

// DeleteRoute ...
func (routing *PluginRouting) DeleteRoute(routeName string) {
	delete(routing.RouteMap, routeName)
}
*/

func TestAddRoute(t *testing.T) {
	t.Log("simple add")
	{
		routing := PluginRouting{
			RouteMap: map[string]PluginRoute{
				"test1": PluginRoute{
					Name:       "test1",
					Source:     "source1",
					Version:    "1.0.1",
					CommitHash: "hash1",
					Executable: "./executable1",
				},
			},
		}

		route := PluginRoute{
			Name:       "test2",
			Source:     "source2",
			Version:    "1.0.2",
			CommitHash: "hash2",
			Executable: "./executable2",
		}

		routing.AddRoute(route)

		route, found := routing.RouteMap["test1"]
		require.Equal(t, true, found)
		require.Equal(t, "test1", route.Name)
		require.Equal(t, "source1", route.Source)
		require.Equal(t, "1.0.1", route.Version)
		require.Equal(t, "hash1", route.CommitHash)
		require.Equal(t, "./executable1", route.Executable)

		route, found = routing.RouteMap["test2"]
		require.Equal(t, true, found)
		require.Equal(t, "test2", route.Name)
		require.Equal(t, "source2", route.Source)
		require.Equal(t, "1.0.2", route.Version)
		require.Equal(t, "hash2", route.CommitHash)
		require.Equal(t, "./executable2", route.Executable)
	}
}

func DeleteRoute(t *testing.T) {
	t.Log("simple delete")
	{
		routing := PluginRouting{
			RouteMap: map[string]PluginRoute{
				"test1": PluginRoute{
					Name:       "test1",
					Source:     "source1",
					Version:    "1.0.1",
					CommitHash: "hash1",
					Executable: "./executable1",
				},
				"test2": PluginRoute{
					Name:       "test2",
					Source:     "source2",
					Version:    "1.0.2",
					CommitHash: "hash2",
					Executable: "./executable2",
				},
			},
		}

		routing.DeleteRoute("test2")

		route, found := routing.RouteMap["test1"]
		require.Equal(t, true, found)
		require.Equal(t, "test1", route.Name)
		require.Equal(t, "source1", route.Source)
		require.Equal(t, "1.0.1", route.Version)
		require.Equal(t, "hash1", route.CommitHash)
		require.Equal(t, "./executable1", route.Executable)

		route, found = routing.RouteMap["test2"]
		require.Equal(t, false, found)
	}
}
