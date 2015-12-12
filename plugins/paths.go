package plugins

import (
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
)

// GetPluginsDir ...
func GetPluginsDir() string {
	if err := bitrise.EnsureBitriseConfigDirExists(); err != nil {
		log.Errorf("Failed to ensure bitrise configs dir, err: %s", err)
	}

	bitriseDir := bitrise.GetBitriseConfigsDirPath()
	pluginsDir := path.Join(bitriseDir, "plugins")

	if err := bitrise.EnsureDir(pluginsDir); err != nil {
		log.Errorf("Failed to ensure path (%s), err: %s", pluginsDir, err)
		return ""
	}

	return pluginsDir
}

// GetPluginPath ...
func GetPluginPath(name, pluginType string) string {
	pluginsPath := GetPluginsDir()
	pluginTypeDir := path.Join(pluginsPath, pluginType)

	if err := bitrise.EnsureDir(pluginTypeDir); err != nil {
		log.Errorf("Failed to ensure path (%s), err: %s", pluginTypeDir, err)
		return ""
	}

	return path.Join(pluginTypeDir, name)
}
