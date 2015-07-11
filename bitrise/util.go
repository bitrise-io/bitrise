package bitrise

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/viktorbenei/depman/pathutil"
)

// ReadWorkflowJSON ...
func ReadWorkflowJSON(pth string) (models.WorkflowModel, error) {
	var workflow models.WorkflowModel

	file, err := os.Open(pth)
	if err != nil {
		return models.WorkflowModel{}, err
	}

	parser := json.NewDecoder(file)
	if err = parser.Decode(&workflow); err != nil {
		return models.WorkflowModel{}, err
	}

	return workflow, nil
}

// NewErrorf ...
func NewErrorf(format string, a ...interface{}) error {
	errStr := fmt.Sprintf(format, a...)
	return errors.New(errStr)
}

// ReadBitriseConfigYML ...
func ReadBitriseConfigYML(pth string) (models.BitriseConfigModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.BitriseConfigModel{}, err
	} else if isExists == false {
		return models.BitriseConfigModel{}, NewErrorf("No file found at path", pth)
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.BitriseConfigModel{}, err
	}
	var bitriseConfig models.BitriseConfigModel
	if err := yaml.Unmarshal(bytes, &bitriseConfig); err != nil {
		return models.BitriseConfigModel{}, err
	}

	return bitriseConfig, nil
}
