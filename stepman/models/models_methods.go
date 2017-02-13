package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pointers"
)

const (
	// DefaultIsAlwaysRun ...
	DefaultIsAlwaysRun = false
	// DefaultIsRequiresAdminUser ...
	DefaultIsRequiresAdminUser = false
	// DefaultIsSkippable ...
	DefaultIsSkippable = false
	// DefaultTimeout ...
	DefaultTimeout = 0
)

// String ...
func (stepInfo StepInfoModel) String() string {
	str := ""
	if stepInfo.GroupInfo.DeprecateNotes != "" {
		str += colorstring.Red("This step is deprecated")

		if stepInfo.GroupInfo.RemovalDate != "" {
			str += colorstring.Redf(" and will be removed: %s", stepInfo.GroupInfo.RemovalDate)
		}
		str += "\n"
	}
	str += fmt.Sprintf("%s %s\n", colorstring.Blue("Library:"), stepInfo.Library)
	str += fmt.Sprintf("%s %s\n", colorstring.Blue("ID:"), stepInfo.ID)
	str += fmt.Sprintf("%s %s\n", colorstring.Blue("Version:"), stepInfo.Version)
	str += fmt.Sprintf("%s %s\n", colorstring.Blue("LatestVersion:"), stepInfo.LatestVersion)
	str += fmt.Sprintf("%s\n\n", colorstring.Blue("Definition:"))

	definition, err := fileutil.ReadStringFromFile(stepInfo.DefinitionPth)
	if err != nil {
		str += colorstring.Redf("Failed to read step definition, error: %s", err)
		return str
	}

	str += definition
	return str
}

// JSON ...
func (stepInfo StepInfoModel) JSON() string {
	bytes, err := json.Marshal(stepInfo)
	if err != nil {
		return fmt.Sprintf(`"Failed to marshal step info (%#v), err: %s"`, stepInfo, err)
	}
	return string(bytes)
}

// CreateFromJSON ...
func (stepInfo StepInfoModel) CreateFromJSON(jsonStr string) (StepInfoModel, error) {
	info := StepInfoModel{}
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		return StepInfoModel{}, err
	}
	return info, nil
}

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
func (source StepSourceModel) validateSource() error {
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

// ValidateInputAndOutputEnvs ...
func (step StepModel) ValidateInputAndOutputEnvs(checkRequiredFields bool) error {
	validateEnvs := func(envs []envmanModels.EnvironmentItemModel) error {
		for _, env := range envs {
			key, _, err := env.GetKeyValuePair()
			if err != nil {
				return fmt.Errorf("Invalid environment (%v), err: %s", env, err)
			}

			if err := env.Validate(); err != nil {
				return fmt.Errorf("Invalid environment (%s), err: %s", key, err)
			}

			if checkRequiredFields {
				options, err := env.GetOptions()
				if err != nil {
					return fmt.Errorf("Invalid environment (%s), err: %s", key, err)
				}

				if options.Title == nil || *options.Title == "" {
					return fmt.Errorf("Invalid environment (%s), err: missing or empty title", key)
				}
			}
		}
		return nil
	}

	if err := validateEnvs(append(step.Inputs, step.Outputs...)); err != nil {
		return err
	}

	return nil
}

// AuditBeforeShare ...
func (step StepModel) AuditBeforeShare() error {
	if step.Title == nil || *step.Title == "" {
		return errors.New("Invalid step: missing or empty required 'title' property")
	}
	if step.Summary == nil || *step.Summary == "" {
		return errors.New("Invalid step: missing or empty required 'summary' property")
	}
	if step.Website == nil || *step.Website == "" {
		return errors.New("Invalid step: missing or empty required 'website' property")
	}

	if step.Timeout != nil && *step.Timeout < 0 {
		return errors.New("Invalid step: timeout less then 0")
	}

	if err := step.ValidateInputAndOutputEnvs(true); err != nil {
		return err
	}

	return nil
}

// Audit ...
func (step StepModel) Audit() error {
	if err := step.AuditBeforeShare(); err != nil {
		return err
	}

	if step.PublishedAt == nil || (*step.PublishedAt).Equal(time.Time{}) {
		return errors.New("Invalid step: missing or empty required 'PublishedAt' property")
	}
	if step.Source == nil {
		return errors.New("Invalid step: missing or empty required 'Source' property")
	}
	if err := step.Source.validateSource(); err != nil {
		return err
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
	if step.Timeout == nil {
		step.Timeout = pointers.NewIntPtr(DefaultTimeout)
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
func (collection StepCollectionModel) IsStepExist(id, version string) bool {
	_, found := collection.GetStep(id, version)
	return found
}

// GetStep ...
func (collection StepCollectionModel) GetStep(id, version string) (StepModel, bool) {
	stepVer, isFound := collection.GetStepVersion(id, version)
	return stepVer.Step, isFound
}

// GetStepVersion ...
func (collection StepCollectionModel) GetStepVersion(id, version string) (StepVersionModel, bool) {
	stepHash := collection.Steps
	stepVersions, found := stepHash[id]
	if !found {
		return StepVersionModel{}, false
	}

	if version == "" {
		version = stepVersions.LatestVersionNumber
	}

	step, found := stepVersions.Versions[version]
	if !found {
		return StepVersionModel{}, false
	}

	return StepVersionModel{
		Step:                   step,
		Version:                version,
		LatestAvailableVersion: stepVersions.LatestVersionNumber,
	}, true
}

// GetDownloadLocations ...
func (collection StepCollectionModel) GetDownloadLocations(id, version string) ([]DownloadLocationModel, error) {
	step, found := collection.GetStep(id, version)
	if found == false {
		return []DownloadLocationModel{}, fmt.Errorf("Collection (%s) doesn't contains step %s (%s)", collection.SteplibSource, id, version)
	}

	if step.Source == nil {
		return []DownloadLocationModel{}, errors.New("Missing Source property")
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

// GetBinaryName ...
func (brewDep BrewDepModel) GetBinaryName() string {
	if brewDep.BinName != "" {
		return brewDep.BinName
	}
	return brewDep.Name
}

// GetBinaryName ...
func (aptGetDep AptGetDepModel) GetBinaryName() string {
	if aptGetDep.BinName != "" {
		return aptGetDep.BinName
	}
	return aptGetDep.Name
}
