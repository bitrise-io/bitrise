package bitrise

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

//=======================================
// Constants
//=======================================

const (
	bitriseVersionSetupStateFileName = "setup.version"
	bitriseConfigFileName            = "config.yml"
)

const bitriseConfigTemplate = `opt_out_analytics: {{.OptOutAnalytics}}`

//=======================================
// Models
//=======================================

// ConfigModel ...
type ConfigModel struct {
	OptOutAnalytics bool
}

type configFileModel struct {
	OptOutAnalytics string `yaml:"opt_out_analytics"`
}

// NewConfigFromBytes ...
func NewConfigFromBytes(bytes []byte) (ConfigModel, error) {
	var fileConfig configFileModel
	if err := yaml.Unmarshal(bytes, &fileConfig); err != nil {
		return ConfigModel{}, err
	}

	if fileConfig.OptOutAnalytics != "" {
		if fileConfig.OptOutAnalytics != "false" && fileConfig.OptOutAnalytics != "true" {
			return ConfigModel{}, fmt.Errorf("Invalid config: opt_out_analytics value should be (\"false\" / \"true\"), actual: (%s)", fileConfig.OptOutAnalytics)
		}
	}

	config := ConfigModel{
		OptOutAnalytics: false,
	}
	if fileConfig.OptOutAnalytics == "true" {
		config.OptOutAnalytics = true
	}

	return config, nil
}

//=======================================
// Utility
//=======================================

func getBitriseConfigVersionSetupFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), bitriseVersionSetupStateFileName)
}

func getBitriseConfigFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), bitriseConfigFileName)
}

func readConfig() (ConfigModel, error) {
	configPth := getBitriseConfigFilePath()
	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return ConfigModel{}, err
	}

	return NewConfigFromBytes(bytes)
}

func ensureConfigExist() error {
	configExist := true
	configPth := getBitriseConfigFilePath()
	if exist, err := pathutil.IsPathExists(configPth); err != nil {
		return err
	} else if !exist {
		configExist = false
	}

	if configExist {
		if _, err := readConfig(); err != nil {
			return err
		}
		return nil
	}

	config := ConfigModel{
		OptOutAnalytics: false,
	}

	if err := SaveConfig(config); err != nil {
		return err
	}

	return nil
}

//=======================================
// Main
//=======================================

// GetBitriseConfigsDirPath ...
func GetBitriseConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".bitrise")
}

// EnsureBitriseConfigDirExists ...
func EnsureBitriseConfigDirExists() error {
	confDirPth := GetBitriseConfigsDirPath()
	return pathutil.EnsureDirExist(confDirPth)
}

// CheckIsSetupWasDoneForVersion ...
func CheckIsSetupWasDoneForVersion(ver string) bool {
	configPth := getBitriseConfigVersionSetupFilePath()
	cont, err := fileutil.ReadStringFromFile(configPth)
	if err != nil {
		return false
	}
	return (cont == ver)
}

// SaveSetupSuccessForVersion ...
func SaveSetupSuccessForVersion(ver string) error {
	if err := EnsureBitriseConfigDirExists(); err != nil {
		return err
	}
	configPth := getBitriseConfigVersionSetupFilePath()
	return fileutil.WriteStringToFile(configPth, ver)
}

// ReadConfig ...
func ReadConfig() (ConfigModel, error) {
	if err := ensureConfigExist(); err != nil {
		return ConfigModel{}, err
	}

	return readConfig()
}

// SaveConfig ...
func SaveConfig(config ConfigModel) error {
	// Converte config to file config
	fileConfig := configFileModel{
		OptOutAnalytics: "false",
	}

	if config.OptOutAnalytics {
		fileConfig.OptOutAnalytics = "false"
	}

	// Write config to file
	configTemplate, err := template.New("config").Parse(bitriseConfigTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(getBitriseConfigFilePath())
	if err != nil {
		return err
	}
	fileWriter := bufio.NewWriter(file)

	err = configTemplate.Execute(fileWriter, config)
	if err != nil {
		return err
	}

	if err = fileWriter.Flush(); err != nil {
		return err
	}

	return nil
}
