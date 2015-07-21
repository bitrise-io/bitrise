package models

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

const (
	optionsKey = "opts"
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
	// DefaultIsNotImportant ...
	DefaultIsNotImportant = false
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
	defaultFalse := false
	defaultTrue := true

	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	if options.Description == nil {
		options.Description = &defaultString
	}
	if options.IsRequired == nil {
		options.IsRequired = &defaultFalse
	}
	if options.IsExpand == nil {
		options.IsExpand = &defaultTrue
	}
	if options.IsDontChangeValue == nil {
		options.IsDontChangeValue = &defaultFalse
	}
	return nil
}

// FillMissingDeafults ...
func (step *StepModel) FillMissingDeafults() error {
	defaultString := ""
	defaultFalse := false

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
		step.IsRequiresAdminUser = &defaultFalse
	}
	if step.IsAlwaysRun == nil {
		step.IsAlwaysRun = &defaultFalse
	}
	if step.IsNotImportant == nil {
		step.IsNotImportant = &defaultFalse
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
			step, found := collection.GetStep(id, version)
			if found {
				location := DownloadLocationModel{
					Type: downloadLocation.Type,
					Src:  *step.Source.Git,
				}
				locations = append(locations, location)
			}
		default:
			return []DownloadLocationModel{}, fmt.Errorf("[STEPMAN] - Invalid download location (%#v) for step (%#v)", downloadLocation, id)
		}
	}
	if len(locations) < 1 {
		return []DownloadLocationModel{}, fmt.Errorf("[STEPMAN] - No download location found for step (%#v)", id)
	}
	return locations, nil
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

// MergeWith ...
func (env *EnvironmentItemModel) MergeWith(otherEnv EnvironmentItemModel) error {
	// merge key-value
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}

	otherKey, otherValue, err := otherEnv.GetKeyValuePair()
	if err != nil {
		return err
	}

	if otherKey != key {
		return errors.New("Env keys are diferent")
	}

	(*env)[key] = otherValue

	//merge options
	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	otherOptions, err := otherEnv.GetOptions()
	if err != nil {
		return err
	}

	if otherOptions.Title != nil {
		*options.Title = *otherOptions.Title
	}
	if otherOptions.Description != nil {
		*options.Description = *otherOptions.Description
	}
	if len(otherOptions.ValueOptions) > 0 {
		options.ValueOptions = otherOptions.ValueOptions
	}
	if otherOptions.IsRequired != nil {
		*options.IsRequired = *otherOptions.IsRequired
	}
	if otherOptions.IsExpand != nil {
		*options.IsExpand = *otherOptions.IsExpand
	}
	if otherOptions.IsDontChangeValue != nil {
		*options.IsDontChangeValue = *otherOptions.IsDontChangeValue
	}
	return nil
}

// MergeWith ...
func (step *StepModel) MergeWith(otherStep StepModel) error {
	if otherStep.Title != nil {
		*step.Title = *otherStep.Title
	}
	if otherStep.Description != nil {
		*step.Description = *otherStep.Description
	}
	if otherStep.Summary != nil {
		*step.Summary = *otherStep.Summary
	}
	if otherStep.Website != nil {
		*step.Website = *otherStep.Website
	}
	if otherStep.SourceCodeURL != nil {
		*step.SourceCodeURL = *otherStep.SourceCodeURL
	}
	if otherStep.SupportURL != nil {
		*step.SupportURL = *otherStep.SupportURL
	}
	if otherStep.Source.Git != nil {
		*step.Source.Git = *otherStep.Source.Git
	}
	if len(otherStep.HostOsTags) > 0 {
		step.HostOsTags = otherStep.HostOsTags
	}
	if len(otherStep.ProjectTypeTags) > 0 {
		step.ProjectTypeTags = otherStep.ProjectTypeTags
	}
	if len(otherStep.TypeTags) > 0 {
		step.TypeTags = otherStep.TypeTags
	}
	if otherStep.IsRequiresAdminUser != nil {
		*step.IsRequiresAdminUser = *otherStep.IsRequiresAdminUser
	}
	if otherStep.IsAlwaysRun != nil {
		*step.IsAlwaysRun = *otherStep.IsAlwaysRun
	}
	if otherStep.IsNotImportant != nil {
		*step.IsNotImportant = *otherStep.IsNotImportant
	}

	for _, input := range step.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return err
		}
		otherInput, found := otherStep.getInputByKey(key)
		if found {
			err := input.MergeWith(otherInput)
			if err != nil {
				return err
			}
		}
	}

	for _, output := range step.Outputs {
		key, _, err := output.GetKeyValuePair()
		if err != nil {
			return err
		}
		otherOutput, found := otherStep.getOutputByKey(key)
		if found {
			err := output.MergeWith(otherOutput)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (step StepModel) getInputByKey(key string) (EnvironmentItemModel, bool) {
	for _, input := range step.Inputs {
		k, _, err := input.GetKeyValuePair()
		if err != nil {
			return EnvironmentItemModel{}, false
		}

		if k == key {
			return input, true
		}
	}
	return EnvironmentItemModel{}, false
}

func (step StepModel) getOutputByKey(key string) (EnvironmentItemModel, bool) {
	for _, output := range step.Outputs {
		k, _, err := output.GetKeyValuePair()
		if err != nil {
			return EnvironmentItemModel{}, false
		}

		if k == key {
			return output, true
		}
	}
	return EnvironmentItemModel{}, false
}
