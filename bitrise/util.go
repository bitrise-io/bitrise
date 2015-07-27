package bitrise

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	stepmanModels "github.com/bitrise-io/stepman/models"

	"github.com/bitrise-io/go-pathutil/pathutil"
)

const (
	clrBlack   = "\x1b[30;1m"
	clrRed     = "\x1b[31;1m"
	clrGreen   = "\x1b[32;1m"
	clrYellow  = "\x1b[33;1m"
	clrBlue    = "\x1b[34;1m"
	clrMagenta = "\x1b[35;1m"
	clrCyan    = "\x1b[36;1m"
	clrWhite   = "\x1b[37;1m"
	clrReset   = "\x1b[0m"
)

// Black ...
func Black(s string) string {
	return addColor(s, clrBlack)
}

// Blackf ...
func Blackf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrBlack)
}

// Red ...
func Red(s string) string {
	return addColor(s, clrRed)
}

// Redf ...
func Redf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrRed)
}

// Green ...
func Green(s string) string {
	return addColor(s, clrBlack)
}

// Greenf ...
func Greenf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrGreen)
}

// Yellow ...
func Yellow(s string) string {
	return addColor(s, clrBlack)
}

// Yellowf ...
func Yellowf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrYellow)
}

// Blue ...
func Blue(s string) string {
	return addColor(s, clrBlue)
}

// Bluef ...
func Bluef(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrBlue)
}

// Magenta ...
func Magenta(s string) string {
	return addColor(s, clrMagenta)
}

// Magentaf ...
func Magentaf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrMagenta)
}

// Cyan ...
func Cyan(s string) string {
	return addColor(s, clrCyan)
}

// Cyanf ...
func Cyanf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrCyan)
}

// White ...
func White(s string) string {
	return addColor(s, clrWhite)
}

// Whitef ...
func Whitef(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a), clrWhite)
}

func addColor(s, color string) string {
	return color + s + clrReset
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

	if err := stepModel.Validate(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.FillMissingDeafults(); err != nil {
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
