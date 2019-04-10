package config

import (
	"fmt"
	"io/ioutil"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"gopkg.in/yaml.v2"
)

const (
	// UserConfigFilePath ...
	UserConfigFilePath = "./.gows.user.yml"

	// SyncModeSymlink ...
	SyncModeSymlink = "symlink"
	// SyncModeCopy ...
	SyncModeCopy = "copy"
	// DefaultSyncMode ...
	DefaultSyncMode = SyncModeSymlink
)

// UserConfigFileAbsPath ...
func UserConfigFileAbsPath() (string, error) {
	return pathutil.AbsPath(UserConfigFilePath)
}

// UserConfigModel - stored in ./.gows.user.yml
type UserConfigModel struct {
	SyncMode string `json:"sync_mode" yaml:"sync_mode"`
}

// CreateDefaultUserConfig ...
func CreateDefaultUserConfig() UserConfigModel {
	return UserConfigModel{
		SyncMode: DefaultSyncMode,
	}
}

// LoadUserConfigFromFile ...
func LoadUserConfigFromFile() (UserConfigModel, error) {
	UserConfigFileAbsPath, err := UserConfigFileAbsPath()
	if err != nil {
		return UserConfigModel{}, fmt.Errorf("Failed to get absolute path of project config: %s", err)
	}

	bytes, err := ioutil.ReadFile(UserConfigFileAbsPath)
	if err != nil {
		return UserConfigModel{}, fmt.Errorf("Failed to read project config file (%s), error: %s", UserConfigFileAbsPath, err)
	}
	var UserConfig UserConfigModel
	if err := yaml.Unmarshal(bytes, &UserConfig); err != nil {
		return UserConfigModel{}, fmt.Errorf("Failed to parse project config (should be valid YML, path: %s), error: %s", UserConfigFileAbsPath, err)
	}

	return UserConfig, nil
}

// SaveUserConfigToFile ...
func SaveUserConfigToFile(projectConf UserConfigModel) error {
	bytes, err := yaml.Marshal(projectConf)
	if err != nil {
		return fmt.Errorf("Failed to parse Project Config (should be valid YML): %s", err)
	}

	err = fileutil.WriteBytesToFile(UserConfigFilePath, bytes)
	if err != nil {
		return fmt.Errorf("Failed to write Project Config into file (%s), error: %s", UserConfigFilePath, err)
	}

	return nil
}
