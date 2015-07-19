package models

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

// -------------------
// --- File models

// EnvironmentItemYMLSerializedModel ...
type EnvironmentItemYMLSerializedModel struct {
	//
	// ! Should use only one property for key (env_key or mapped_to)
	//
	EnvKey            string   `json:"env_key" yaml:"env_key"`
	MappedTo          string   `json:"mapped_to" yaml:"mapped_to"`
	Value             string   `json:"value" yaml:"value"`
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// StepYMLSerializedModel ....
type StepYMLSerializedModel struct {
	Name                string                              `json:"name" yaml:"name"`
	Description         string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string                              `json:"website" yaml:"website"`
	ForkURL             string                              `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel                     `json:"source" yaml:"source"`
	HostOsTags          []string                            `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string                            `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string                            `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool                                `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []EnvironmentItemYMLSerializedModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemYMLSerializedModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// EnvironmentItemOptionsSerializeModel ...
type EnvironmentItemOptionsSerializeModel struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// EnvironmentItemSerializeModel ...
type EnvironmentItemSerializeModel map[string]interface{}

// StepSourceModel ...
type StepSourceModel struct {
	Git string `json:"git" yaml:"git"`
}

// StepSerializeModel ...
type StepSerializeModel struct {
	Name                string                          `json:"name" yaml:"name"`
	Description         string                          `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string                          `json:"website" yaml:"website"`
	ForkURL             string                          `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel                 `json:"source" yaml:"source"`
	HostOsTags          []string                        `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string                        `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string                        `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool                            `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []EnvironmentItemSerializeModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemSerializeModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepListItemSerializeModel ...
type StepListItemSerializeModel map[string]StepSerializeModel

// WorkflowSerializeModel ...
type WorkflowSerializeModel struct {
	Environments []EnvironmentItemSerializeModel `json:"environments"`
	Steps        []StepListItemSerializeModel    `json:"steps"`
}

// AppSerializeModel ...
type AppSerializeModel struct {
	Environments []EnvironmentItemSerializeModel `json:"environments" yaml:"environments"`
}

// BitriseConfigSerializeModel ...
type BitriseConfigSerializeModel struct {
	FormatVersion string                            `json:"format_version" yaml:"format_version"`
	App           AppSerializeModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowSerializeModel `json:"workflows" yaml:"workflows"`
}

// -------------------
// --- Struct methods

// GetKeyValuePair ...
func (envFile EnvironmentItemSerializeModel) GetKeyValuePair() (string, string, error) {
	if len(envFile) < 3 {
		for key, value := range envFile {
			if key != OptionsKey {
				valueStr, ok := value.(string)
				if ok == false {
					return "", "", fmt.Errorf("Invalid value (key:%#v) (value:%#v)", key, value)
				}
				return key, valueStr, nil
			}
		}
	}
	return "", "", errors.New("Invalid envFile")
}

// // SetFieldOnInterface ...
// func SetFieldOnInterface(obj interface{}, name string, value interface{}) error {
// 	structValue := reflect.ValueOf(obj).Elem()
// 	structFieldValue := structValue.FieldByName(name)
//
// 	if !structFieldValue.IsValid() {
// 		return fmt.Errorf("No such field: %s in obj", name)
// 	}
//
// 	if !structFieldValue.CanSet() {
// 		return fmt.Errorf("Cannot set %s field value", name)
// 	}
//
// 	structFieldType := structFieldValue.Type()
// 	val := reflect.ValueOf(value)
// 	if structFieldType != val.Type() {
// 		return errors.New("Provided value type didn't match obj field type")
// 	}
//
// 	structFieldValue.Set(val)
// 	return nil
// }

// // GetStructTagFieldMap ...
// // Returns a map[string]string containing:
// //  value_specified_for_field_by_tag->NameOfField
// // To get a map of a Struct's fields based on it's "json" tag:
// //  GetStructTagFieldMap(myObj, "json")
// // isOnlyFirstTag: returns only the first tag if the tag value is a comma separated list
// //  ex from: json:"description,omitempty
// //  removes/ignores the 'omitempty' modifier and only returns "description"
// func GetStructTagFieldMap(obj interface{}, tagName string, isOnlyFirstTag bool) map[string]string {
// 	retMap := map[string]string{}
// 	val := reflect.Indirect(reflect.ValueOf(obj))
// 	valType := val.Type()
// 	for i := valType.NumField(); i >= 0; i-- {
// 		tagVal := valType.Field(i).Tag.Get("json")
// 		if tagVal != "" {
// 			if isOnlyFirstTag {
// 				tagVal = strings.Split(tagVal, ",")[0]
// 			}
// 			retMap[tagVal] = valType.Field(i).Name
// 		}
// 	}
// 	return retMap
// }

// ParseFromInterfaceMap ...
func (envSerModel *EnvironmentItemOptionsSerializeModel) ParseFromInterfaceMap(input map[interface{}]interface{}) error {
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
			envSerModel.Title = castedValue
		case "description":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Description = castedValue
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
func (envFile EnvironmentItemSerializeModel) GetOptions() (EnvironmentItemOptionsSerializeModel, error) {
	value, found := envFile[OptionsKey]
	if !found {
		return EnvironmentItemOptionsSerializeModel{}, nil
	}

	envItmCasted, ok := value.(EnvironmentItemOptionsSerializeModel)
	if ok {
		return envItmCasted, nil
	}

	// if it's read from a file (YAML/JSON) then it's most likely not the proper type
	//  so cast it from the generic interface-interface map
	optionsInterfaceMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return EnvironmentItemOptionsSerializeModel{}, fmt.Errorf("Invalid options (value:%#v) - failed to map-interface cast", value)
	}

	options := EnvironmentItemOptionsSerializeModel{}
	err := options.ParseFromInterfaceMap(optionsInterfaceMap)
	if err != nil {
		return EnvironmentItemOptionsSerializeModel{}, err
	}

	log.Debugf("Parsed options: %#v\n", options)

	return options, nil

	// // cast
	// options := EnvironmentItemOptionsSerializeModel{}
	// tagFieldMap := GetStructTagFieldMap(options, "json", true)
	// for k, v := range optionsInterfaceMap {
	// 	serKey, ok := k.(string)
	// 	if serKey == "value_options" {
	// 		options.ValueOptions, ok = v.([]string)
	// 		if !ok {
	// 			return EnvironmentItemOptionsSerializeModel{}, fmt.Errorf("Invalid options (value:%#v) - value_options is not an array of strings", value)
	// 		}
	// 	} else {
	// 		if !ok {
	// 			return EnvironmentItemOptionsSerializeModel{}, fmt.Errorf("Invalid options (value:%#v) - key (%#v) is not a string", value, k)
	// 		}
	// 		relatedStructFieldName, found := tagFieldMap[serKey]
	// 		if !found {
	// 			return EnvironmentItemOptionsSerializeModel{}, fmt.Errorf("Invalid options (value:%#v) - no struct type mapping info found for (key:%s)", value, k)
	// 		}
	// 		if err := SetFieldOnInterface(&options, relatedStructFieldName, v); err != nil {
	// 			return EnvironmentItemOptionsSerializeModel{}, fmt.Errorf("Invalid options (value:%#v) - failed to convert key-value to options model (key:%s) (value:%#v)", value, k, v)
	// 		}
	// 	}
	// }
}

// GetStepIDStepDataPair ...
func (stepListItm StepListItem) GetStepIDStepDataPair() (string, StepModel, error) {
	if len(stepListItm) > 1 {
		return "", StepModel{}, errors.New("StepListItem contains more than 1 key-value pair!")
	}
	for key, value := range stepListItm {
		return key, value, nil
	}
	return "", StepModel{}, errors.New("StepListItem does not contain a key-value pair!")
}
