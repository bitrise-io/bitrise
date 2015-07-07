package envman

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
)

const (
	IS_EXPAND_KEY string = "is_expand"
	TRUE_KEY      string = "true"
	FALSE_KEY     string = "false"
)

// This is the model of ENVIRONMENT in envman, for methods
type EnvModel struct {
	Key      string
	Value    string
	IsExpand bool
}

// This is the model of ENVIRONMENT in envman, for storing in file
type EnvMapItem map[string]string

type envsYMLModel struct {
	Envs []EnvMapItem `yml:"environments"`
}

// Convert envsYMLModel to envModel array
func (envYML envsYMLModel) convertToEnvModelArray() []EnvModel {
	var envModels []EnvModel
	for _, envMapItem := range envYML.Envs {
		envModel := envMapItem.convertToEnvModel()
		envModels = append(envModels, envModel)
	}
	return envModels
}

func (eMap EnvMapItem) convertToEnvModel() EnvModel {
	var eModel EnvModel
	for key, value := range eMap {
		if key != IS_EXPAND_KEY {
			eModel.Key = key
			eModel.Value = value
		}
	}
	eModel.IsExpand = IsExpand(eMap[IS_EXPAND_KEY])
	return eModel
}

func IsExpand(s string) bool {
	if s == "" {
		return true
	} else {
		expand, err := strconv.ParseBool(s)
		if err != nil {
			log.Errorln("isExpand: Failed to parse input:", err)
			return true
		}
		return expand
	}
}

// Convert envModel array to envsYMLModel
func convertToEnvsYMLModel(eModels []EnvModel) envsYMLModel {
	var envYML envsYMLModel
	var envMaps []EnvMapItem
	for _, eModel := range eModels {
		eMap := eModel.convertToEnvMap()
		envMaps = append(envMaps, eMap)
	}
	envYML.Envs = envMaps
	return envYML
}

func (eModel EnvModel) convertToEnvMap() EnvMapItem {
	eMap := make(EnvMapItem)
	if eModel.IsExpand == false {
		eMap[IS_EXPAND_KEY] = FALSE_KEY
	}
	eMap[eModel.Key] = eModel.Value
	return eMap
}
