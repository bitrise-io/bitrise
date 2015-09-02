package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
)

const (
	// DefaultIsAlwaysRun ...
	DefaultIsAlwaysRun = false
	// DefaultIsRequiresAdminUser ...
	DefaultIsRequiresAdminUser = false
	// DefaultIsSkippable ...
	DefaultIsSkippable = false
)

// ValidateStepInputOutputModel ...
func ValidateStepInputOutputModel(env envmanModels.EnvironmentItemModel, checkRequiredFields bool) error {
	if err := env.Validate(); err != nil {
		return err
	}

	if checkRequiredFields {
		options, err := env.GetOptions()
		if err != nil {
			return err
		}

		if options.Title == nil || *options.Title == "" {
			return errors.New("Invalid environment: missing or empty title")
		}
	}

	return nil
}

// -------------------
// --- Struct methods

// Normalize ...
func (step StepModel) Normalize() error {
	for _, input := range step.Inputs {
		if err := input.Normalize(); err != nil {
			return err
		}
	}
	for _, output := range step.Outputs {
		if err := output.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSource ...
func (source StepSourceModel) ValidateSource() error {
	if source.Git == "" {
		return errors.New("Invalid step: missing or empty required 'source.git' property")
	}

	if !strings.HasPrefix(source.Git, "http://") && !strings.HasPrefix(source.Git, "https://") {
		return errors.New("Invalid step: step source should start with http:// or https://")
	}
	if !strings.HasSuffix(source.Git, ".git") {
		return errors.New("Invalid step: step source should end with .git")
	}

	if source.Commit == "" {
		return errors.New("Invalid step: missing or empty required 'source.commit' property")
	}
	return nil
}

// Validate ...
func (step StepModel) Validate(checkRequiredFields bool) error {
	if checkRequiredFields {
		if step.Title == nil || *step.Title == "" {
			return errors.New("Invalid step: missing or empty required 'title' property")
		}
		if step.Summary == nil || *step.Summary == "" {
			return errors.New("Invalid step: missing or empty required 'summary' property")
		}
		if step.Website == nil || *step.Website == "" {
			return errors.New("Invalid step: missing or empty required 'website' property")
		}
		if step.PublishedAt == nil || (*step.PublishedAt).Equal(time.Time{}) {
			return errors.New("Invalid step: missing or empty required 'PublishedAt' property")
		}
		if err := step.Source.ValidateSource(); err != nil {
			return err
		}
	}

	for _, input := range step.Inputs {
		var err error
		err = ValidateStepInputOutputModel(input, checkRequiredFields)
		if err != nil {
			return err
		}
	}

	for _, output := range step.Outputs {
		var err error
		err = ValidateStepInputOutputModel(output, checkRequiredFields)
		if err != nil {
			return err
		}
	}

	return nil
}

// FillMissingDefaults ...
func (step *StepModel) FillMissingDefaults() error {
	if step.Title == nil {
		step.Title = pointers.NewStringPtr("")
	}
	if step.Description == nil {
		step.Description = pointers.NewStringPtr("")
	}
	if step.Summary == nil {
		step.Summary = pointers.NewStringPtr("")
	}
	if step.Website == nil {
		step.Website = pointers.NewStringPtr("")
	}
	if step.SourceCodeURL == nil {
		step.SourceCodeURL = pointers.NewStringPtr("")
	}
	if step.SupportURL == nil {
		step.SupportURL = pointers.NewStringPtr("")
	}
	if step.IsRequiresAdminUser == nil {
		step.IsRequiresAdminUser = pointers.NewBoolPtr(DefaultIsRequiresAdminUser)
	}
	if step.IsAlwaysRun == nil {
		step.IsAlwaysRun = pointers.NewBoolPtr(DefaultIsAlwaysRun)
	}
	if step.IsSkippable == nil {
		step.IsSkippable = pointers.NewBoolPtr(DefaultIsSkippable)
	}
	if step.RunIf == nil {
		step.RunIf = pointers.NewStringPtr("")
	}

	for _, input := range step.Inputs {
		err := input.FillMissingDefaults()
		if err != nil {
			return err
		}
	}
	for _, output := range step.Outputs {
		err := output.FillMissingDefaults()
		if err != nil {
			return err
		}
	}
	return nil
}

// IsStepExist ...
func (collection StepCollectionModel) IsStepExist(id string) bool {
	stepHash := collection.Steps
	_, found := stepHash[id]
	if !found {
		return false
	}
	return true
}

// GetStep ...
func (collection StepCollectionModel) GetStep(id, version string) (StepModel, bool) {
	stepHash := collection.Steps
	stepVersions, found := stepHash[id]
	if !found {
		return StepModel{}, false
	}
	step, found := stepVersions.Versions[version]
	if !found {
		return StepModel{}, false
	}
	return step, true
}

// GetDownloadLocations ...
func (collection StepCollectionModel) GetDownloadLocations(id, version string) ([]DownloadLocationModel, error) {
	step, found := collection.GetStep(id, version)
	if found == false {
		return []DownloadLocationModel{}, fmt.Errorf("Collection (%s) doesn't contains step %s (%s)", collection.SteplibSource, id, version)
	}

	locations := []DownloadLocationModel{}
	for _, downloadLocation := range collection.DownloadLocations {
		switch downloadLocation.Type {
		case "zip":
			url := downloadLocation.Src + id + "/" + version + "/step.zip"
			location := DownloadLocationModel{
				Type: downloadLocation.Type,
				Src:  url,
			}
			locations = append(locations, location)
		case "git":
			location := DownloadLocationModel{
				Type: downloadLocation.Type,
				Src:  step.Source.Git,
			}
			locations = append(locations, location)
		default:
			return []DownloadLocationModel{}, fmt.Errorf("[STEPMAN] - Invalid download location (%#v) for step (%#v)", downloadLocation, id)
		}
	}
	if len(locations) < 1 {
		return []DownloadLocationModel{}, fmt.Errorf("[STEPMAN] - No download location found for step (%#v)", id)
	}
	return locations, nil
}

// GetLatestStepVersion ...
func (collection StepCollectionModel) GetLatestStepVersion(id string) (string, error) {
	stepHash := collection.Steps
	stepGroup, found := stepHash[id]
	if !found {
		return "", fmt.Errorf("Collection (%s) doesn't contains step (%s)", collection.SteplibSource, id)
	}

	if stepGroup.LatestVersionNumber == "" {
		return "", fmt.Errorf("Failed to find latest version of step %s", id)
	}

	return stepGroup.LatestVersionNumber, nil
}
