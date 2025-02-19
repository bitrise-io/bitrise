package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidatePluginFromYML(t *testing.T) {
	version.VERSION = "1.49.3"

	t.Log("simple plugin - with executables")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger: DidFinishRun
executable:
  osx: bin_url
  linux: bin_url
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)
		require.NoError(t, err)

		// Then
		require.NoError(t, validatePlugin(plugin, pth, "bitrise"))

		require.Equal(t, "step", plugin.Name)
		require.Equal(t, "Manage Bitrise CLI steps", plugin.Description)
		require.Equal(t, "DidFinishRun", plugin.TriggerEvent)

		assert.Equal(t, 1, len(plugin.Requirements))
		requirement := plugin.Requirements[0]
		require.Equal(t, "bitrise", requirement.Tool)
		require.Equal(t, "1.3.0", requirement.MinVersion)
		require.Equal(t, "", requirement.MaxVersion)
	}

	t.Log("simple plugin - no trigger, triggers present")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
triggers: 
  - WillStartRun
  - DidFinishRun
executable:
  osx: bin_url
  linux: bin_url
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)
		require.NoError(t, err)

		// Then
		require.Equal(t, "WillStartRun", plugin.TriggerEvents[0])
		require.Equal(t, "DidFinishRun", plugin.TriggerEvents[1])
	}

	t.Log("simple plugin - both trigger and triggers present")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger: DidMagic
triggers: 
  - WillStartRun
  - DidFinishRun
executable:
  osx: bin_url
  linux: bin_url
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)
		require.NoError(t, err)

		// Then
		require.Equal(t, "DidMagic", plugin.TriggerEvent)
		require.Equal(t, "WillStartRun", plugin.TriggerEvents[0])
		require.Equal(t, "DidFinishRun", plugin.TriggerEvents[1])
	}

	t.Log("invalid plugin - no name")
	{
		// Given
		pluginStr := `name: 
description: |-
  Manage Bitrise CLI steps
trigger:
executable:
  osx: bin_url
  linux: bin_url
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)

		// Then
		require.NoError(t, err)
		require.EqualError(t, validatePlugin(plugin, pth, "bitrise"), "missing name")
	}

	t.Log("invalid plugin - no linux executable")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger:
executable:
  osx: bin_url
  linux: 
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)

		// Then
		require.NoError(t, err)
		require.EqualError(t, validatePlugin(plugin, pth, "bitrise"), "both osx and linux executable should be defined, or non of them")
	}

	t.Log("invalid plugin - no osx executable")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger:
executable:
  osx: 
  linux: bin_url
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)

		// Then
		require.NoError(t, err)
		require.EqualError(t, validatePlugin(plugin, pth, "bitrise"), "both osx and linux executable should be defined, or non of them")
	}

	t.Log("invalid plugin - no executables, no bitrise-plugin.sh")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger:
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`
		pth := givenPluginYMLWithContent(pluginStr, t)

		// When
		plugin, err := ParsePluginFromYML(pth)

		// Then
		require.NoError(t, err)
		err = validatePlugin(plugin, pth, "bitrise")
		require.Error(t, err)
		require.Equal(t, true, strings.Contains(err.Error(), "no executable defined, nor bitrise-plugin.sh exist at:"))
	}

	t.Log("simple plugin - with bitrise-plugin.sh")
	{
		// Given
		pluginStr := `name: step
description: |-
  Manage Bitrise CLI steps
trigger:
requirements:
- tool: bitrise
  min_version: 1.3.0
  max_version: ""
`

		pth := givenPluginYMLWithContent(pluginStr, t)
		write(t, "test", filepath.Join(filepath.Dir(pth), "bitrise-plugin.sh"))

		// When
		plugin, err := ParsePluginFromYML(pth)

		// Then
		require.NoError(t, err)

		require.NoError(t, validatePlugin(plugin, pth, "bitrise"))

		require.Equal(t, "step", plugin.Name)
		require.Equal(t, "Manage Bitrise CLI steps", plugin.Description)
		require.Equal(t, 1, len(plugin.Requirements))

		requirement := plugin.Requirements[0]
		require.Equal(t, "bitrise", requirement.Tool)
		require.Equal(t, "1.3.0", requirement.MinVersion)
		require.Equal(t, "", requirement.MaxVersion)
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

func TestNewPluginRoute(t *testing.T) {
	// Given
	pluginStr := `name: Plugin
description: |-
  Manage Bitrise CLI steps
trigger: DidFinishRun
triggers: 
  - Run1
  - Run2
executable:
  osx: bin_url
  linux: bin_url
requirements:
  - tool: bitrise
    min_version: 1.3.0
    max_version: ""
`
	source := "source"
	version := "1.2.3"

	pth := givenPluginYMLWithContent(pluginStr, t)

	// When
	plugin, err := ParsePluginFromYML(pth)
	require.NoError(t, err)
	route, err := NewPluginRoute(plugin, source, version)
	require.NoError(t, err)

	// Then
	assertPluginRouteEqual(t, route, plugin, source, version)
}

func givenPluginYMLWithContent(content string, t *testing.T) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__plugin_test__")
	require.NoError(t, err)

	pth := filepath.Join(tmpDir, "bitrise-plugin.yml")
	write(t, content, pth)

	return pth
}

func write(t *testing.T, content, toPth string) {
	toDir := filepath.Dir(toPth)
	exist, err := pathutil.IsDirExists(toDir)
	require.NoError(t, err)
	if !exist {
		require.NoError(t, os.MkdirAll(toDir, 0700))
	}
	require.NoError(t, fileutil.WriteStringToFile(toPth, content))
}

func assertPluginRouteEqual(t *testing.T, route PluginRoute, plugin Plugin, source, version string) {
	assert.Equal(t, plugin.Name, route.Name)
	assert.Equal(t, source, route.Source)
	assert.Equal(t, version, route.Version)
	assert.Equal(t, plugin.ExecutableURL(), route.Executable)
	assert.Equal(t, plugin.TriggerEvent, route.TriggerEvent)
	assert.Equal(t, plugin.TriggerEvents, route.TriggerEvents)
}
