package models

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

const (
	optionsKey string = "opts"
)

// -------------------
// --- Common models

// EnvironmentItemOptionsModel ...
type EnvironmentItemOptionsModel struct {
	Title             *string  `json:"title,omitempty" yaml:"title,omitempty"`
	Description       *string  `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// EnvironmentItemModel ...
type EnvironmentItemModel map[string]interface{}

// StepSourceModel ...
type StepSourceModel struct {
	Git *string `json:"git,omitempty" yaml:"git,omitempty"`
}

// StepModel ...
type StepModel struct {
	Title               *string                `json:"title,omitempty" yaml:"title,omitempty"`
	Description         *string                `json:"description,omitempty" yaml:"description,omitempty"`
	Summary             *string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Website             *string                `json:"website,omitempty" yaml:"website,omitempty"`
	SourceCodeURL       *string                `json:"source_code_url,omitempty" yaml:"source_code_url,omitempty"`
	SupportURL          *string                `json:"support_url,omitempty" yaml:"support_url,omitempty"`
	Source              StepSourceModel        `json:"source,omitempty" yaml:"source,omitempty"`
	HostOsTags          []string               `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string               `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string               `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser *bool                  `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	IsAlwaysRun         *bool                  `json:"is_always_run,omitempty" yaml:"is_always_run,omitempty"`
	IsNotImportant      *bool                  `json:"is_not_important,omitempty" yaml:"is_not_important,omitempty"`
	Inputs              []EnvironmentItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// -------------------
// --- Steplib models

// StepGroupModel ...
type StepGroupModel struct {
	ID       string               `json:"id"`
	Versions map[string]StepModel `json:"versions"`
	Latest   StepModel            `json:"latest"`
}

// StepHash ...
type StepHash map[string]StepGroupModel

// DownloadLocationModel ...
type DownloadLocationModel struct {
	Type string `json:"type"`
	Src  string `json:"src"`
}

// StepCollectionModel ...
type StepCollectionModel struct {
	FormatVersion        string                  `json:"format_version" yaml:"format_version"`
	GeneratedAtTimeStamp int64                   `json:"generated_at_timestamp" yaml:"generated_at_timestamp"`
	Steps                StepHash                `json:"steps" yaml:"steps"`
	SteplibSource        string                  `json:"steplib_source" yaml:"steplib_source"`
	DownloadLocations    []DownloadLocationModel `json:"download_locations" yaml:"download_locations"`
}

// WorkFlowModel ...
type WorkFlowModel struct {
	FormatVersion string      `json:"format_version"`
	Environments  []string    `json:"environments"`
	Steps         []StepModel `json:"steps"`
}

// -------------------
// --- Bitrise-cli models

// StepListItemModel ...
type StepListItemModel map[string]StepModel

// WorkflowModel ...
type WorkflowModel struct {
	Environments []EnvironmentItemModel `json:"environments"`
	Steps        []StepListItemModel    `json:"steps"`
}

// AppModel ...
type AppModel struct {
	Environments []EnvironmentItemModel `json:"environments" yaml:"environments"`
}

// BitriseConfigModel ...
type BitriseConfigModel struct {
	FormatVersion string                   `json:"format_version" yaml:"format_version"`
	App           AppModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowModel `json:"workflows" yaml:"workflows"`
}

// -------------------
// --- Struct methods

// Validate ...
func (env EnvironmentItemModel) Validate() error {
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

// GetStep ...
func (collection StepCollectionModel) GetStep(id, version string) (bool, StepModel) {
	log.Debugln("-> GetStep")
	stepHash := collection.Steps
	for stepID, stepGroup := range stepHash {
		if stepID == id {
			for stepVersion, step := range stepGroup.Versions {
				if stepVersion == version {
					return true, step
				}

			}
		}
	}

	return false, StepModel{}
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
			found, step := collection.GetStep(id, version)
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
	if len(env) < 3 {
		retKey := ""
		retValue := ""

		for key, value := range env {
			if key != optionsKey {
				if retKey != "" {
					return "", "", errors.New("Invalid env: more then 1 key-value field found!")
				}

				valueStr, ok := value.(string)
				if ok == false {
					return "", "", fmt.Errorf("Invalid value (key:%#v) (value:%#v)", key, value)
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

	return "", "", errors.New("Invalid env: more then 2 fields ")
}

// ParseFromInterfaceMap ...
func (envSerModel *EnvironmentItemOptionsModel) ParseFromInterfaceMap(input map[interface{}]interface{}) error {
	for key, value := range input {
		keyStr, ok := key.(string)
		if !ok {
			return fmt.Errorf("Invalid key, should be a string: %#v", key)
		}
		switch keyStr {
		case "title":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			*envSerModel.Title = castedValue
		case "description":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			*envSerModel.Description = castedValue
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
	if len(env) > 2 {
		return EnvironmentItemOptionsModel{}, errors.New("Invalid env: more then 2 field")
	}

	optsShouldExist := false
	if len(env) == 2 {
		optsShouldExist = true
	}

	value, found := env[optionsKey]
	if !found {
		if optsShouldExist {
			return EnvironmentItemOptionsModel{}, errors.New("Invalid env: 2 fields but, no opts found")
		}
		return EnvironmentItemOptionsModel{}, nil
	}

	envItmCasted, ok := value.(EnvironmentItemOptionsModel)
	if ok {
		return envItmCasted, nil
	}

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
