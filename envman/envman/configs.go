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
	defaultEnvBytesLimitInKB     = 20
	defaultEnvListBytesLimitInKB = 100
)

// ConfigsModel ...
type ConfigsModel struct {
	EnvBytesLimitInKB     int `json:"env_bytes_limit_in_kb,omitempty" yaml:"env_bytes_limit_in_kb,omitempty"`
	EnvListBytesLimitInKB int `json:"env_list_bytes_limit_in_kb,omitempty" yaml:"env_list_bytes_limit_in_kb,omitempty"`
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

// CheckIfConfigsSaved ...
func CheckIfConfigsSaved() bool {
	configPth := getEnvmanConfigsFilePath()
	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return false
	}

	var configs ConfigsModel
	if err := json.Unmarshal(bytes, &configs); err != nil {
		return false
	}
	return true
}

// GetConfigs ...
func GetConfigs() (ConfigsModel, error) {
	configPth := getEnvmanConfigsFilePath()
	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return ConfigsModel{}, err
	}

	type ConfigsFileMode struct {
		EnvBytesLimitInKB     *int `json:"env_bytes_limit_in_kb,omitempty" yaml:"env_bytes_limit_in_kb,omitempty"`
		EnvListBytesLimitInKB *int `json:"env_list_bytes_limit_in_kb,omitempty" yaml:"env_list_bytes_limit_in_kb,omitempty"`
	}

	var configs ConfigsFileMode
	if err := json.Unmarshal(bytes, &configs); err != nil {
		return ConfigsModel{}, err
	}

	defaultConfigs := ConfigsModel{
		EnvBytesLimitInKB:     defaultEnvBytesLimitInKB,
		EnvListBytesLimitInKB: defaultEnvListBytesLimitInKB,
	}

	if configs.EnvBytesLimitInKB != nil {
		defaultConfigs.EnvBytesLimitInKB = *configs.EnvBytesLimitInKB
	}
	if configs.EnvListBytesLimitInKB != nil {
		defaultConfigs.EnvListBytesLimitInKB = *configs.EnvListBytesLimitInKB
	}

	return defaultConfigs, nil
}

// SaveDefaultConfigs ...
func SaveDefaultConfigs() error {
	if err := ensureEnvmanConfigDirExists(); err != nil {
		return err
	}

	defaultConfigs := ConfigsModel{
		EnvBytesLimitInKB:     defaultEnvBytesLimitInKB,
		EnvListBytesLimitInKB: defaultEnvListBytesLimitInKB,
	}
	bytes, err := json.Marshal(defaultConfigs)
	if err != nil {
		return err
	}
	configsPth := getEnvmanConfigsFilePath()
	return fileutil.WriteBytesToFile(configsPth, bytes)
}
