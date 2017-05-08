package plugins

import (
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	pluginInputPayloadKey        = "BITRISE_PLUGIN_INPUT_PAYLOAD"
	pluginInputBitriseVersionKey = "BITRISE_PLUGIN_INPUT_BITRISE_VERSION"
	pluginInputTriggerEventKey   = "BITRISE_PLUGIN_INPUT_TRIGGER"
	pluginInputPluginModeKey     = "BITRISE_PLUGIN_INPUT_PLUGIN_MODE"
	pluginInputDataDirKey        = "BITRISE_PLUGIN_INPUT_DATA_DIR"

	bitrisePluginOutputEnvKey = "BITRISE_PLUGIN_OUTPUT"
)

const bitrisePluginPrefix = ":"

const (
	triggerMode PluginMode = "trigger"
	commandMode PluginMode = "command"
)

// PluginMode ...
type PluginMode string

// PluginInput ...
type PluginInput map[string]string

// ParseArgs ...
func ParseArgs(args []string) (string, []string, bool) {
	log.Debugf("args: %v", args)

	if len(args) == 0 {
		return "", []string{}, false
	}

	pluginName := ""
	pluginArgs := []string{}
	for idx, arg := range args {

		if strings.Contains(arg, bitrisePluginPrefix) {
			pluginSplits := strings.Split(arg, ":")

			if len(pluginSplits) != 2 {
				return "", []string{}, false
			}

			pluginName = pluginSplits[1]
			if len(args) > idx {
				pluginArgs = args[idx+1 : len(args)]
			}
			return pluginName, pluginArgs, true
		}
	}

	return "", []string{}, false
}

// CheckForNewVersion ...
func CheckForNewVersion(plugin Plugin) (string, error) {
	route, found, err := ReadPluginRoute(plugin.Name)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("no route found for already loaded plugin (%s)", plugin.Name)
	}
	if route.Version == "" {
		// local plugin, can not update
		return "", nil
	}

	pluginSrcDir := GetPluginSrcDir(plugin.Name)

	gitDirPath := filepath.Join(pluginSrcDir, ".git")
	if exist, err := pathutil.IsPathExists(gitDirPath); err != nil {
		return "", fmt.Errorf("failed to check if .git folder exist at (%s), error: %s", gitDirPath, err)
	} else if !exist {
		return "", fmt.Errorf(".git folder not exist at (%s), error: %s", gitDirPath, err)
	}

	versions, err := GitVersionTags(pluginSrcDir)
	if err != nil {
		return "", err
	}

	log.Debug("")
	log.Debugf("versions: %v", versions)

	if len(versions) == 0 {
		return "", nil
	}

	latestVersion := versions[len(versions)-1]

	currentVersion, err := GetPluginVersion(plugin.Name)
	if err != nil {
		return "", fmt.Errorf("failed to check installed plugin (%s) version, error: %s", plugin.Name, err)
	}

	log.Debug("")
	log.Debugf("latestVersion: %v", latestVersion)
	log.Debugf("currentVersion: %v", currentVersion)

	if currentVersion == nil {
		log.Debugf("Installed plugin (%s) is from local source, nothing to check...", plugin.Name)
		return "", nil
	}

	if latestVersion.GreaterThan(currentVersion) {
		return latestVersion.String(), nil
	}

	return "", nil
}

// LoadPlugin ...
func LoadPlugin(name string) (Plugin, bool, error) {
	pluginDir := GetPluginDir(name)

	if exists, err := pathutil.IsDirExists(pluginDir); err != nil {
		return Plugin{}, false, fmt.Errorf("Failed to check dir (%s), err: %s", pluginDir, err)
	} else if !exists {
		return Plugin{}, false, nil
	}

	pluginYMLPath := GetPluginYMLPath(name)

	plugin, err := ParseAndValidatePluginFromYML(pluginYMLPath)
	if err != nil {
		return Plugin{}, true, err
	}

	return plugin, true, nil
}

// InstalledPluginList ...
func InstalledPluginList() ([]Plugin, error) {
	routing, err := readPluginRouting()
	if err != nil {
		return []Plugin{}, err
	}

	pluginList := []Plugin{}

	for name := range routing.RouteMap {
		if plugin, found, err := LoadPlugin(name); err != nil {
			return []Plugin{}, err
		} else if !found {
			return []Plugin{}, fmt.Errorf("Plugin (%s) found in route, but could not load it", name)
		} else {
			pluginList = append(pluginList, plugin)
		}
	}

	return pluginList, nil
}
