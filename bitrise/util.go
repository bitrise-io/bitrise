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

	var specStep stepmanModels.StepModel
	if err := yaml.Unmarshal(bytes, &specStep); err != nil {
		return models.StepModel{}, err
	}

	return convertStepmanToBitriseStepModel(specStep)
}

func convertStepmanToBitriseStepModel(specStep stepmanModels.StepModel) (models.StepModel, error) {
	inputs := []models.EnvironmentItemModel{}
	for _, specEnv := range specStep.Inputs {
		env, err := convertStepmanToBitriseEnvironmentItemModel(specEnv)
		if err != nil {
			return models.StepModel{}, err
		}
		inputs = append(inputs, env)
	}

	outputs := []models.EnvironmentItemModel{}
	for _, specEnv := range specStep.Outputs {
		env, err := convertStepmanToBitriseEnvironmentItemModel(specEnv)
		if err != nil {
			return models.StepModel{}, err
		}
		outputs = append(outputs, env)
	}

	step := models.StepModel{
		Name:        specStep.Name,
		Description: specStep.Description,
		Website:     specStep.Website,
		ForkURL:     specStep.ForkURL,
		Source: models.StepSourceModel{
			Git: specStep.Source.Git,
		},
		HostOsTags:          specStep.HostOsTags,
		ProjectTypeTags:     specStep.ProjectTypeTags,
		TypeTags:            specStep.TypeTags,
		IsRequiresAdminUser: *specStep.IsRequiresAdminUser,
		Inputs:              inputs,
		Outputs:             outputs,
	}

	return step, nil
}

// ToEnvironmentItemModel ...
func convertStepmanToBitriseEnvironmentItemModel(specEnv stepmanModels.EnvironmentItemModel) (models.EnvironmentItemModel, error) {
	isRequired := models.DefaultIsRequired
	if specEnv.IsRequired != nil {
		isRequired = *specEnv.IsRequired
	}

	isExpand := models.DefaultIsExpand
	if specEnv.IsExpand != nil {
		isExpand = *specEnv.IsExpand
	}

	isDontChnageValue := models.DefaultIsDontChangeValue
	if specEnv.IsDontChangeValue != nil {
		isDontChnageValue = *specEnv.IsDontChangeValue
	}

	env := models.EnvironmentItemModel{
		EnvKey:            specEnv.EnvKey,
		Value:             specEnv.Value,
		Title:             specEnv.Title,
		Description:       specEnv.Description,
		ValueOptions:      specEnv.ValueOptions,
		IsRequired:        isRequired,
		IsExpand:          isExpand,
		IsDontChangeValue: isDontChnageValue,
	}

	return env, nil
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
