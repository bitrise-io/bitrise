package bitrise

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// TimeToFormattedSeconds ...
func TimeToFormattedSeconds(t time.Duration, postfix string) string {
	sec := t.Seconds()
	if sec > 10.0 {
		return fmt.Sprintf("%.f%s", sec, postfix)
	} else if sec < 1.0 {
		return fmt.Sprintf("%.2f%s", sec, postfix)
	}
	return fmt.Sprintf("%.1f%s", sec, postfix)
}

// SaveConfigToFile ...
func SaveConfigToFile(pth string, bitriseConf models.BitriseDataModel) error {
	contBytes, err := generateYAML(bitriseConf)
	if err != nil {
		return err
	}
	if err := WriteBytesToFile(pth, contBytes); err != nil {
		return err
	}

	log.Println()
	log.Infoln("=> Init success!")
	log.Infoln("File created at path:", pth)

	return nil
}

func generateYAML(v interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

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
	var bitriseData models.BitriseDataModel
	if err := yaml.Unmarshal(bytes, &bitriseData); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.Normalize(); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.Validate(); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.FillMissingDefaults(); err != nil {
		return models.BitriseDataModel{}, err
	}

	return bitriseData, nil
}

// ReadSpecStep ...
func ReadSpecStep(pth string) (stepmanModels.StepModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return stepmanModels.StepModel{}, err
	} else if !isExists {
		return stepmanModels.StepModel{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return stepmanModels.StepModel{}, err
	}

	var stepModel stepmanModels.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.Normalize(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.ValidateStep(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	return stepModel, nil
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

// IsVersionBetween ...
//  returns true if it's between the lower and upper limit
//  or in case it matches the lower or the upper limit
func IsVersionBetween(verBase, verLower, verUpper string) (bool, error) {
	r1, err := stepmanModels.CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	r2, err := stepmanModels.CompareVersions(verBase, verUpper)
	if err != nil {
		return false, err
	}
	if r2 == -1 {
		return false, nil
	}

	return true, nil
}

// IsVersionGreaterOrEqual ...
//  returns true if verBase is greater or equal to verLower
func IsVersionGreaterOrEqual(verBase, verLower string) (bool, error) {
	r1, err := stepmanModels.CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	return true, nil
}
