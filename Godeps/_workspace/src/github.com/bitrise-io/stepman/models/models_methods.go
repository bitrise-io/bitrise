package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	envmanModels "github.com/bitrise-io/envman/models"
)

// Validate ...
func Validate(env envmanModels.EnvironmentItemModel) error {
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("Invalid environment: empty env_key")
	}

	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	if options.Title == nil || *options.Title == "" {
		return errors.New("Invalid environment: missing or empty title")
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

// Validate ...
func (step StepModel) Validate() error {
	if step.Title == nil || *step.Title == "" {
		return errors.New("Invalid step: missing or empty required 'title' property")
	}
	if step.Summary == nil || *step.Summary == "" {
		return errors.New("Invalid step: missing or empty required 'summary' property")
	}
	if step.Website == nil || *step.Website == "" {
		return errors.New("Invalid step: missing or empty required 'website' property")
	}
	if step.Source.Git == nil || *step.Source.Git == "" {
		return errors.New("Invalid step: missing or empty required 'source' property")
	}
	for _, input := range step.Inputs {
		err := Validate(input)
		if err != nil {
			return err
		}
	}
	for _, output := range step.Outputs {
		err := Validate(output)
		if err != nil {
			return err
		}
	}
	return nil
}

// FillMissingDeafults ...
func (step *StepModel) FillMissingDeafults() error {
	defaultString := ""

	if step.Description == nil {
		step.Description = &defaultString
	}
	if step.SourceCodeURL == nil {
		step.SourceCodeURL = &defaultString
	}
	if step.SupportURL == nil {
		step.SupportURL = &defaultString
	}
	if step.IsRequiresAdminUser == nil {
		step.IsRequiresAdminUser = &envmanModels.DefaultIsRequiresAdminUser
	}
	if step.IsAlwaysRun == nil {
		step.IsAlwaysRun = &envmanModels.DefaultIsAlwaysRun
	}
	if step.IsSkippable == nil {
		step.IsSkippable = &envmanModels.DefaultIsSkippable
	}
	if step.RunIf == nil {
		step.RunIf = &defaultString
	}

	for _, input := range step.Inputs {
		err := input.FillMissingDeafults()
		if err != nil {
			return err
		}
	}
	for _, output := range step.Outputs {
		err := output.FillMissingDeafults()
		if err != nil {
			return err
		}
	}
	return nil
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
		return []DownloadLocationModel{}, fmt.Errorf("Collection doesn't contains step %s (%s)", id, version)
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
				Src:  *step.Source.Git,
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

// CompareVersions ...
// semantic version (X.Y.Z)
// 1 if version 2 is greater then version 1, -1 if not
// -2 & error if can't compare (if not supported component found)
func CompareVersions(version1, version2 string) (int, error) {
	version1Slice := strings.Split(version1, ".")
	version2Slice := strings.Split(version2, ".")

	lenDiff := len(version1Slice) - len(version2Slice)
	if lenDiff != 0 {
		makeDefVerComps := func(compLen int) []string {
			comps := make([]string, compLen, compLen)
			for idx := len(comps) - 1; idx >= 0; idx-- {
				comps[idx] = "0"
			}
			return comps
		}
		if lenDiff > 0 {
			// v1 slice is longer
			version2Slice = append(version2Slice, makeDefVerComps(lenDiff)...)
		} else {
			// v2 slice is longer
			version1Slice = append(version1Slice, makeDefVerComps(-lenDiff)...)
		}
	}

	cnt := len(version1Slice)
	for i, num := range version1Slice {
		num1, err := strconv.ParseInt(num, 0, 64)
		if err != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err)
			return -2, err
		}

		num2, err2 := strconv.ParseInt(version2Slice[i], 0, 64)
		if err2 != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err2)
			return -2, err
		}

		if num2 > num1 {
			return 1, nil
		}
		if i == cnt-1 {
			// last one
			if num2 == num1 {
				return 0, nil
			}
		}
	}
	return -1, nil
}

// GetLatestStepVersion ...
func (collection StepCollectionModel) GetLatestStepVersion(id string) (string, error) {
	stepHash := collection.Steps
	stepGroup, found := stepHash[id]
	if !found {
		return "", fmt.Errorf("Collection doesn't contains step %s", id)
	}

	if stepGroup.LatestVersionNumber == "" {
		return "", fmt.Errorf("Failed to find latest version of step %s", id)
	}

	return stepGroup.LatestVersionNumber, nil
}
