package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	optionsKey = "opts"
)

var (
	//DefaultIsRequired ...
	DefaultIsRequired = false
	// DefaultIsExpand ...
	DefaultIsExpand = true
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue = false
	// DefaultIsAlwaysRun ...
	DefaultIsAlwaysRun = false
	// DefaultIsRequiresAdminUser ...
	DefaultIsRequiresAdminUser = false
	// DefaultIsSkippable ...
	DefaultIsSkippable = false
)

// -------------------
// --- Struct methods

// Normalize ...
func (step StepModel) Normalize() error {
	for _, input := range step.Inputs {
		opts, err := input.GetOptions()
		if err != nil {
			return err
		}
		input[optionsKey] = opts
	}
	for _, output := range step.Outputs {
		opts, err := output.GetOptions()
		if err != nil {
			return err
		}
		output[optionsKey] = opts
	}
	return nil
}

// Validate ...
func (env EnvironmentItemModel) Validate() error {
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("Invalid environment: empty env_key")
	}

	log.Debugln("-> validate")
	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	if options.Title == nil || *options.Title == "" {
		return errors.New("Invalid environment: missing or empty title")
	}

	return nil
}

// Validate ...
func (step StepModel) Validate() error {
	if step.Title == nil || *step.Title == "" {
		return errors.New("Invalid step: missing or empty title")
	}
	if step.Summary == nil || *step.Summary == "" {
		return errors.New("Invalid step: missing or empty summary")
	}
	if step.Website == nil || *step.Website == "" {
		return errors.New("Invalid step: missing or empty website")
	}
	if step.Source.Git == nil || *step.Source.Git == "" {
		return errors.New("Invalid step: missing or empty source")
	}
	for _, input := range step.Inputs {
		err := input.Validate()
		if err != nil {
			return err
		}
	}
	for _, output := range step.Outputs {
		err := output.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// FillMissingDeafults ...
func (env *EnvironmentItemModel) FillMissingDeafults() error {
	defaultString := ""

	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	if options.Description == nil {
		options.Description = &defaultString
	}
	if options.IsRequired == nil {
		options.IsRequired = &DefaultIsRequired
	}
	if options.IsExpand == nil {
		options.IsExpand = &DefaultIsExpand
	}
	if options.IsDontChangeValue == nil {
		options.IsDontChangeValue = &DefaultIsDontChangeValue
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
		step.IsRequiresAdminUser = &DefaultIsRequiresAdminUser
	}
	if step.IsAlwaysRun == nil {
		step.IsAlwaysRun = &DefaultIsAlwaysRun
	}
	if step.IsSkippable == nil {
		step.IsSkippable = &DefaultIsSkippable
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

// GetKeyValuePair ...
func (env EnvironmentItemModel) GetKeyValuePair() (string, string, error) {
	if len(env) > 2 {
		return "", "", errors.New("Invalid env: more then 2 fields ")
	}

	retKey := ""
	retValue := ""

	for key, value := range env {
		if key != optionsKey {
			if retKey != "" {
				return "", "", errors.New("Invalid env: more then 1 key-value field found!")
			}

			valueStr, ok := value.(string)
			if !ok {
				if value == nil {
					valueStr = ""
				} else {
					return "", "", fmt.Errorf("Invalid value (key:%#v) (value:%#v)", key, value)
				}
			}

			retKey = key
			retValue = valueStr
		}
	}

	if retKey == "" {
		return "", "", errors.New("Invalid env: no envKey specified!")
	}

	return retKey, retValue, nil
}

// ParseFromInterfaceMap ...
func (envSerModel *EnvironmentItemOptionsModel) ParseFromInterfaceMap(input map[interface{}]interface{}) error {
	for key, value := range input {
		keyStr, ok := key.(string)
		if !ok {
			return fmt.Errorf("Invalid key, should be a string: %#v", key)
		}
		log.Debugf("  ** processing (key:%#v) (value:%#v) (envSerModel:%#v)", key, value, envSerModel)
		switch keyStr {
		case "title":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Title = &castedValue
		case "description":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Description = &castedValue
		case "value_options":
			castedValue, ok := value.([]string)
			if !ok {
				// try with []interface{} instead and cast the
				//  items to string
				castedValue = []string{}
				interfArr, ok := value.([]interface{})
				if !ok {
					return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
				}
				for _, interfItm := range interfArr {
					castedItm, ok := interfItm.(string)
					if !ok {
						return fmt.Errorf("Invalid value in value_options (%#v), not a string: %#v", interfArr, interfItm)
					}
					castedValue = append(castedValue, castedItm)
				}
			}
			envSerModel.ValueOptions = castedValue
		case "is_required":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsRequired = &castedValue
		case "is_expand":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsExpand = &castedValue
		case "is_dont_change_value":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsDontChangeValue = &castedValue
		default:
			return fmt.Errorf("Not supported key found in options: %#v", key)
		}
	}
	return nil
}

// GetOptions ...
func (env EnvironmentItemModel) GetOptions() (EnvironmentItemOptionsModel, error) {
	value, found := env[optionsKey]
	if !found {
		return EnvironmentItemOptionsModel{}, nil
	}

	envItmCasted, ok := value.(EnvironmentItemOptionsModel)
	if ok {
		return envItmCasted, nil
	}

	log.Debugf(" * processing env:%#v", env)

	// if it's read from a file (YAML/JSON) then it's most likely not the proper type
	//  so cast it from the generic interface-interface map
	optionsInterfaceMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return EnvironmentItemOptionsModel{}, fmt.Errorf("Invalid options (value:%#v) - failed to map-interface cast", value)
	}

	options := EnvironmentItemOptionsModel{}
	err := options.ParseFromInterfaceMap(optionsInterfaceMap)
	if err != nil {
		return EnvironmentItemOptionsModel{}, err
	}

	log.Debugf("Parsed options: %#v\n", options)

	return options, nil
}
