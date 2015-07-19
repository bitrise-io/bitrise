package bitrise

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
)

// ReadBitriseConfig ...
func ReadBitriseConfig(pth string) (models.BitriseDataModel, error) {
	log.Debugln("-> ReadBitriseConfig")
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.BitriseDataModel{}, err
	} else if !isExists {
		return models.BitriseDataModel{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.BitriseDataModel{}, err
	}
	var bitriseConfigFile models.BitriseConfigSerializeModel
	if err := yaml.Unmarshal(bytes, &bitriseConfigFile); err != nil {
		return models.BitriseDataModel{}, err
	}

	return bitriseConfigFile.ToBitriseDataModel()
}

// ReadSpecStep ...
func ReadSpecStep(pth string) (models.StepModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.StepModel{}, err
	} else if !isExists {
		return models.StepModel{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.StepModel{}, err
	}
	var stepYML models.StepYMLSerializedModel
	if err := yaml.Unmarshal(bytes, &stepYML); err != nil {
		return models.StepModel{}, err
	}

	return stepYML.ToStepModel()
}

// WriteStringToFile ...
func WriteStringToFile(pth string, fileCont string) error {
	return WriteBytesToFile(pth, []byte(fileCont))
}

// WriteBytesToFile ...
func WriteBytesToFile(pth string, fileCont []byte) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorln("Failed to close file:", err)
		}
	}()

	if _, err := file.Write(fileCont); err != nil {
		return err
	}

	return nil
}
