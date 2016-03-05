package plugins

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/depman/pathutil"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	ver "github.com/hashicorp/go-version"
)

//=======================================
// Plugin
//=======================================

// NewPluginFromBytes ...
func NewPluginFromBytes(bytes []byte) (plugin Plugin, err error) {
	if err = yaml.Unmarshal(bytes, &plugin); err != nil {
		return Plugin{}, err
	}
	if err := plugin.Validate(); err != nil {
		return Plugin{}, err
	}
	return plugin, nil
}

// NewPluginFromYML ...
func NewPluginFromYML(ymlPth string) (Plugin, error) {
	if isExists, err := pathutil.IsPathExists(ymlPth); err != nil {
		return Plugin{}, err
	} else if !isExists {
		return Plugin{}, fmt.Errorf("Plugin yml path (%s) doesn't exist", ymlPth)
	}

	bytes, err := fileutil.ReadBytesFromFile(ymlPth)
	if err != nil {
		return Plugin{}, err
	}

	return NewPluginFromBytes(bytes)
}

// Validate ...
func (plugin Plugin) Validate() error {
	if plugin.Name == "" {
		return fmt.Errorf("invalid plugin: missing required name")
	}
	return nil
}

func (plugin Plugin) String() string {
	pluginStr := colorstring.Green(plugin.Name)
	pluginStr += fmt.Sprintf("\n  Description: %s", plugin.Description)
	return pluginStr
}

func systemOsName() (string, error) {
	osOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("uname", "-s")
	if err != nil {
		return "", err
	}
	return strip(osOut), nil
}

// ExecutableURL ...
func (plugin Plugin) ExecutableURL() string {
	systemOS, err := systemOsName()
	if err != nil {
		return ""
	}

	switch systemOS {
	case "Darwin":
		return plugin.Executable.Osx
	case "Linux":
		return plugin.Executable.Linux
	default:
		return ""
	}
}

//=======================================
// Sorting

// SortByName ...
func SortByName(plugins []Plugin) {
	byName := func(p1, p2 *Plugin) bool {
		return p1.Name < p2.Name
	}

	sortBy(byName).sort(plugins)
}

type sortBy func(p1, p2 *Plugin) bool

func (by sortBy) sort(plugins []Plugin) {
	ps := &pluginSorter{
		plugins: plugins,
		sortBy:  by,
	}
	sort.Sort(ps)
}

type pluginSorter struct {
	plugins []Plugin
	sortBy  sortBy
}

//=======================================
// sort.Interface

func (s *pluginSorter) Len() int {
	return len(s.plugins)
}

func (s *pluginSorter) Swap(i, j int) {
	s.plugins[i], s.plugins[j] = s.plugins[j], s.plugins[i]
}

func (s *pluginSorter) Less(i, j int) bool {
	return s.sortBy(&s.plugins[i], &s.plugins[j])
}

//=======================================
// PluginRoute
//=======================================

// NewPluginRoute ...
func NewPluginRoute(name, source, executable, version, commitHash, triggerEvent string) (PluginRoute, error) {
	route := PluginRoute{
		Name:         name,
		Source:       source,
		Executable:   executable,
		Version:      version,
		CommitHash:   commitHash,
		TriggerEvent: triggerEvent,
	}
	if err := route.Validate(); err != nil {
		return PluginRoute{}, err
	}
	return route, nil
}

// Validate ...
func (route PluginRoute) Validate() error {
	if route.Name == "" {
		return fmt.Errorf("invalid route: missing required name")
	}
	if route.Source == "" {
		return fmt.Errorf("invalid route: missing required source")
	}
	if route.Version != "" {
		if _, err := ver.NewVersion(route.Version); err != nil {
			return fmt.Errorf("invalid route: invalid version (%s)", route.Version)
		}
	}
	return nil
}

//=======================================
// PluginRouting
//=======================================

// NewPluginRouting ...
func NewPluginRouting() PluginRouting {
	return PluginRouting{RouteMap: map[string]PluginRoute{}}
}

// NewPluginRoutingFromBytes ...
func NewPluginRoutingFromBytes(bytes []byte) (PluginRouting, error) {
	var routing PluginRouting
	if err := yaml.Unmarshal(bytes, &routing); err != nil {
		return PluginRouting{}, err
	}
	if err := routing.Validate(); err != nil {
		return PluginRouting{}, err
	}
	return routing, nil
}

// NewPluginRoutingFromYMLOrEmpty ...
func NewPluginRoutingFromYMLOrEmpty(ymlPth string) (PluginRouting, error) {
	if exist, err := pathutil.IsPathExists(ymlPth); err != nil {
		return PluginRouting{}, err
	} else if exist {
		bytes, err := fileutil.ReadBytesFromFile(ymlPth)
		if err != nil {
			return PluginRouting{}, err
		}

		return NewPluginRoutingFromBytes(bytes)
	}

	return NewPluginRouting(), nil
}

// Validate ...
func (routing PluginRouting) Validate() error {
	for name, route := range routing.RouteMap {
		if name == "" {
			return fmt.Errorf("invalid routing: missing required route's key")
		}
		if name != route.Name {
			return fmt.Errorf("invalid routing: route's key (%s) should equal to route's name (%s)", name, route.Name)
		}
		if err := route.Validate(); err != nil {
			return fmt.Errorf("invalid routing: invalid plugin: %s", err)
		}
	}
	return nil
}

// AddRoute ...
func (routing *PluginRouting) AddRoute(route PluginRoute) {
	routing.RouteMap[route.Name] = route
}

// DeleteRoute ...
func (routing *PluginRouting) DeleteRoute(routeName string) {
	delete(routing.RouteMap, routeName)
}
