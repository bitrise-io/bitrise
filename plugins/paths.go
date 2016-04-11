package plugins

import (
	"fmt"
	"path"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	ver "github.com/hashicorp/go-version"
)

const (
	pluginsDirName = "plugins"
	pluginSpecName = "spec.yml"

	pluginShName  = "bitrise-plugin.sh"
	pluginYMLName = "bitrise-plugin.yml"
)

var (
	pluginsDir        = ""
	pluginsRoutingPth = ""
)

// -----------------------
// --- Routing
// -----------------------

// CreateAndAddPluginRoute ...
func CreateAndAddPluginRoute(name, source, executable, version, commitHash, triggerEvent string) error {
	newRoute, err := NewPluginRoute(name, source, executable, version, commitHash, triggerEvent)
	if err != nil {
		return err
	}

	return AddPluginRoute(newRoute)
}

// AddPluginRoute ...
func AddPluginRoute(route PluginRoute) error {
	routing, err := readPluginRouting()
	if err != nil {
		return err
	}

	routing.AddRoute(route)

	return writeRoutingToFile(routing)
}

// DeletePluginRoute ...
func DeletePluginRoute(name string) error {
	routing, err := readPluginRouting()
	if err != nil {
		return err
	}

	routing.DeleteRoute(name)

	return writeRoutingToFile(routing)
}

// GetPluginVersion ...
func GetPluginVersion(name string) (*ver.Version, error) {
	route, found, err := ReadPluginRoute(name)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("plugin not installed with name (%s)", name)
	}

	if route.Version == "" {
		return nil, nil
	}

	pluginVersion, err := ver.NewVersion(route.Version)
	if err != nil {
		return nil, err
	}
	if pluginVersion == nil {
		return nil, fmt.Errorf("failed to parse version (%s)", route.Version)
	}

	return pluginVersion, nil
}

// ReadPluginRoute ...
func ReadPluginRoute(name string) (PluginRoute, bool, error) {
	routing, err := readPluginRouting()
	if err != nil {
		return PluginRoute{}, false, err
	}

	route, found := routing.RouteMap[name]
	return route, found, nil
}

func writeRoutingToFile(routing PluginRouting) error {
	bytes, err := yaml.Marshal(routing)
	if err != nil {
		return err
	}

	return fileutil.WriteBytesToFile(pluginsRoutingPth, bytes)
}

func readPluginRouting() (PluginRouting, error) {
	return NewPluginRoutingFromYMLOrEmpty(pluginsRoutingPth)
}

// -----------------------
// --- Paths
// -----------------------

// GetPluginDir ...
func GetPluginDir(name string) string {
	return path.Join(pluginsDir, name)
}

// GetPluginSrcDir ...
func GetPluginSrcDir(name string) string {
	return path.Join(GetPluginDir(name), "src")
}

// GetPluginBinDir ...
func GetPluginBinDir(name string) string {
	return path.Join(GetPluginDir(name), "bin")
}

// GetPluginDataDir ...
func GetPluginDataDir(name string) string {
	return path.Join(GetPluginDir(name), "data")
}

// GetPluginYMLPath ...
func GetPluginYMLPath(name string) string {
	return path.Join(GetPluginSrcDir(name), pluginYMLName)
}

// GetPluginExecutablePath ...
func GetPluginExecutablePath(name string) (string, bool, error) {
	route, found, err := ReadPluginRoute(name)
	if err != nil {
		return "", false, err
	}
	if !found {
		return "", false, fmt.Errorf("plugin not installed with name (%s)", name)
	}

	if route.Executable != "" {
		return path.Join(GetPluginBinDir(name), name), true, nil
	}
	return path.Join(GetPluginSrcDir(name), pluginShName), false, nil
}

// -----------------------
// --- Init
// -----------------------

// InitPaths ...
func InitPaths() error {
	// Plugins dir
	if err := configs.EnsureBitriseConfigDirExists(); err != nil {
		log.Errorf("Failed to ensure bitrise configs dir, err: %s", err)
	}

	bitriseDir := configs.GetBitriseConfigsDirPath()
	tmpPluginsDir := path.Join(bitriseDir, pluginsDirName)

	if err := pathutil.EnsureDirExist(tmpPluginsDir); err != nil {
		return err
	}

	pluginsDir = tmpPluginsDir

	// Plugins routing
	pluginsRoutingPth = path.Join(pluginsDir, pluginSpecName)

	return nil
}
