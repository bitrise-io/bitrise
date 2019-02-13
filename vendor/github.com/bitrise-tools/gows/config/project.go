package config

import (
	"fmt"
	"io/ioutil"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"gopkg.in/yaml.v2"
)

const (
	// ProjectConfigFilePath ...
	ProjectConfigFilePath = "./gows.yml"
)

// ProjectConfigFileAbsPath ...
func ProjectConfigFileAbsPath() (string, error) {
	return pathutil.AbsPath(ProjectConfigFilePath)
}

// ProjectConfigModel - stored in ./gows.yml
type ProjectConfigModel struct {
	PackageName string `json:"package_name" yaml:"package_name"`
}

// LoadProjectConfigFromFile ...
func LoadProjectConfigFromFile() (ProjectConfigModel, error) {
	projectConfigFileAbsPath, err := ProjectConfigFileAbsPath()
	if err != nil {
		return ProjectConfigModel{}, fmt.Errorf("Failed to get absolute path of project config: %s", err)
	}

	bytes, err := ioutil.ReadFile(projectConfigFileAbsPath)
	if err != nil {
		return ProjectConfigModel{}, fmt.Errorf("Failed to read project config file (%s), error: %s", projectConfigFileAbsPath, err)
	}
	var projectConfig ProjectConfigModel
	if err := yaml.Unmarshal(bytes, &projectConfig); err != nil {
		return ProjectConfigModel{}, fmt.Errorf("Failed to parse project config (should be valid YML, path: %s), error: %s", projectConfigFileAbsPath, err)
	}

	return projectConfig, nil
}

// SaveProjectConfigToFile ...
func SaveProjectConfigToFile(projectConf ProjectConfigModel) error {
	bytes, err := yaml.Marshal(projectConf)
	if err != nil {
		return fmt.Errorf("Failed to parse Project Config (should be valid YML): %s", err)
	}

	err = fileutil.WriteBytesToFile(ProjectConfigFilePath, bytes)
	if err != nil {
		return fmt.Errorf("Failed to write Project Config into file (%s), error: %s", ProjectConfigFilePath, err)
	}

	return nil
}
