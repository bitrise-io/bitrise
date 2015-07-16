package bitrise

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
)

// NewErrorf ...
func NewErrorf(format string, a ...interface{}) error {
	errStr := fmt.Sprintf(format, a...)
	return errors.New(errStr)
}

// ReadBitriseConfigYML ...
func ReadBitriseConfigYML(pth string) (models.BitriseConfigModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.BitriseConfigModel{}, err
	} else if !isExists {
		return models.BitriseConfigModel{}, NewErrorf("No file found at path", pth)
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.BitriseConfigModel{}, err
	}
	var bitriseConfigYML models.BitriseConfigYMLModel
	if err := yaml.Unmarshal(bytes, &bitriseConfigYML); err != nil {
		return models.BitriseConfigModel{}, err
	}

	return bitriseConfigYML.BitriseConfigModel(), nil
}
