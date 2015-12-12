package plugins

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pathutil"
)

var pluginsPath string

func getBitriseConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".bitrise")
}

func ensureBitriseConfigDirExists() error {
	confDirPth := getBitriseConfigsDirPath()
	isExists, err := pathutil.IsDirExists(confDirPth)
	if !isExists || err != nil {
		if err := os.MkdirAll(confDirPth, 0777); err != nil {
			return err
		}
	}
	return nil
}

// GetPluginsPath ...
func GetPluginsPath() string {
	if err := ensureBitriseConfigDirExists(); err != nil {
		log.Errorf("Failed to ensure bitrise configs dir, err: %s", err)
	}

	bitriseDir := getBitriseConfigsDirPath()

	return path.Join(bitriseDir, "plugins")
}

// GetPluginPath ...
func GetPluginPath(name, pluginType string) string {
	pluginsPath := GetPluginsPath()
	return path.Join(pluginsPath, pluginType, name)
}
