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

// ReadSpecStep ...
func ReadSpecStep(pth string) (models.StepModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.StepModel{}, err
	} else if !isExists {
		return models.StepModel{}, NewErrorf("No file found at path", pth)
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.StepModel{}, err
	}
	var specStep models.StepModel
	if err := yaml.Unmarshal(bytes, &specStep); err != nil {
		return models.StepModel{}, err
	}

	return specStep, nil
}

// MergeSpecStep ...
func MergeSpecStep(specStep, workflowStep models.StepModel) models.StepModel {
	specStep.ID = mergeString(specStep.ID, workflowStep.ID)
	specStep.SteplibSource = mergeString(specStep.SteplibSource, workflowStep.SteplibSource)
	specStep.VersionTag = mergeString(specStep.VersionTag, workflowStep.VersionTag)
	specStep.Name = mergeString(specStep.Name, workflowStep.Name)
	specStep.Description = mergeString(specStep.Description, workflowStep.Description)
	specStep.Website = mergeString(specStep.Website, workflowStep.Website)
	specStep.ForkURL = mergeString(specStep.ForkURL, workflowStep.ForkURL)
	specStep.Source = mergeStringStringMap(specStep.Source, workflowStep.Source)
	specStep.HostOsTags = mergeStringSlice(specStep.HostOsTags, workflowStep.HostOsTags)
	specStep.ProjectTypeTags = mergeStringSlice(specStep.ProjectTypeTags, workflowStep.ProjectTypeTags)
	specStep.TypeTags = mergeStringSlice(specStep.TypeTags, workflowStep.TypeTags)
	specStep.IsRequiresAdminUser = workflowStep.IsRequiresAdminUser
	specStep.Inputs = mergeInputModels(specStep.Inputs, workflowStep.Inputs)
	specStep.Outputs = mergeInputModels(specStep.Outputs, workflowStep.Outputs)
	return specStep
}

func mergeBoolPtr(reference, override *bool) *bool {
	if override != nil {
		return override
	}
	return reference
}

func mergeInputModel(reference, override models.InputModel) models.InputModel {
	reference.MappedTo = mergeString(reference.MappedTo, override.MappedTo)
	reference.Title = mergeString(reference.Title, override.Title)
	reference.Description = mergeString(reference.Description, override.Description)
	reference.Value = mergeString(reference.Value, override.Value)
	reference.ValueOptions = mergeStringSlice(reference.ValueOptions, override.ValueOptions)
	reference.IsRequired = mergeBoolPtr(reference.IsRequired, override.IsRequired)
	reference.IsExpand = mergeBoolPtr(reference.IsExpand, override.IsExpand)
	reference.IsDontChangeValue = mergeBoolPtr(reference.IsDontChangeValue, override.IsDontChangeValue)
	return reference
}

func mergeInputModels(reference, override []models.InputModel) []models.InputModel {
	for idx, referenceInput := range reference {
		for _, overrideInput := range override {
			if referenceInput.MappedTo == overrideInput.MappedTo {
				reference[idx] = mergeInputModel(referenceInput, overrideInput)
			}
		}
	}
	return reference
}

func mergeStringSlice(reference, override []string) []string {
	slice := []string{}
	copy(slice, reference)
	for _, overrideStr := range override {
		if stringSliceContains(reference, overrideStr) == false {
			slice = append(slice, overrideStr)
		}
	}
	return slice
}

func stringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func mergeStringStringMap(reference, override map[string]string) map[string]string {
	for referenceKey, referenceValue := range reference {
		for overrideKey, overrideValue := range override {
			if referenceKey == overrideKey {
				reference[referenceKey] = mergeString(referenceValue, overrideValue)
			}
		}
	}
	return reference
}

func mergeString(reference, override string) string {
	if override != "" {
		return override
	}
	return reference
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
