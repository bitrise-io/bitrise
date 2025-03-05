package envman

import (
	"encoding/json"
	"os"
	"path"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	envmanConfigFileName         = "configs.json"
	defaultEnvBytesLimitInKB     = 256
	defaultEnvListBytesLimitInKB = 256
)

// ConfigsModel ...
type ConfigsModel struct {
	EnvBytesLimitInKB     int `json:"env_bytes_limit_in_kb,omitempty"`
	EnvListBytesLimitInKB int `json:"env_list_bytes_limit_in_kb,omitempty"`
}

func getEnvmanConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".envman")
}

func getEnvmanConfigsFilePath() string {
	return path.Join(getEnvmanConfigsDirPath(), envmanConfigFileName)
}

func ensureEnvmanConfigDirExists() error {
	confDirPth := getEnvmanConfigsDirPath()
	isExists, err := pathutil.IsDirExists(confDirPth)
	if !isExists || err != nil {
		if err := os.MkdirAll(confDirPth, 0777); err != nil {
			return err
		}
	}
	return nil
}

func createDefaultConfigsModel() ConfigsModel {
	return ConfigsModel{
		EnvBytesLimitInKB:     defaultEnvBytesLimitInKB,
		EnvListBytesLimitInKB: defaultEnvListBytesLimitInKB,
	}
}

// GetConfigs ...
func GetConfigs() (ConfigsModel, error) {
	configPth := getEnvmanConfigsFilePath()
	defaultConfigs := createDefaultConfigsModel()

	if isExist, err := pathutil.IsPathExists(configPth); err != nil {
		return ConfigsModel{}, err
	} else if !isExist {
		return defaultConfigs, nil
	}

	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return ConfigsModel{}, err
	}

	type ConfigsFileMode struct {
		EnvBytesLimitInKB     *int `json:"env_bytes_limit_in_kb,omitempty"`
		EnvListBytesLimitInKB *int `json:"env_list_bytes_limit_in_kb,omitempty"`
	}

	var userConfigs ConfigsFileMode
	if err := json.Unmarshal(bytes, &userConfigs); err != nil {
		return ConfigsModel{}, err
	}

	if userConfigs.EnvBytesLimitInKB != nil {
		defaultConfigs.EnvBytesLimitInKB = *userConfigs.EnvBytesLimitInKB
	}
	if userConfigs.EnvListBytesLimitInKB != nil {
		defaultConfigs.EnvListBytesLimitInKB = *userConfigs.EnvListBytesLimitInKB
	}

	return defaultConfigs, nil
}

// saveConfigs ...
//  only used for unit testing at the moment
func saveConfigs(configModel ConfigsModel) error {
	if err := ensureEnvmanConfigDirExists(); err != nil {
		return err
	}

	bytes, err := json.Marshal(configModel)
	if err != nil {
		return err
	}
	configsPth := getEnvmanConfigsFilePath()
	return fileutil.WriteBytesToFile(configsPth, bytes)
}
