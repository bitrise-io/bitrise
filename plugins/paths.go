package plugins

import (
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
)

// GetPluginsDir ...
func GetPluginsDir() (string, error) {
	if err := bitrise.EnsureBitriseConfigDirExists(); err != nil {
		log.Errorf("Failed to ensure bitrise configs dir, err: %s", err)
	}

	bitriseDir := bitrise.GetBitriseConfigsDirPath()
	pluginsDir := path.Join(bitriseDir, "plugins")

	if err := bitrise.EnsureDir(pluginsDir); err != nil {
		return "", err
	}

	return pluginsDir, nil
}

// GetPluginPath ...
func GetPluginPath(name, pluginType string) (string, error) {
	pluginsPath, err := GetPluginsDir()
	if err != nil {
		return "", err
	}

	pluginTypeDir := path.Join(pluginsPath, pluginType)

	if err := bitrise.EnsureDir(pluginTypeDir); err != nil {
		return "", err
	}

	return path.Join(pluginTypeDir, name), nil
}
