package envman

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/models"
)

const (
	// IsExpandKey ...
	IsExpandKey string = "is_expand"
	// TrueKey ...
	TrueKey string = "true"
	// FalseKey ...
	FalseKey string = "false"
)

type envsYMLModel struct {
	Envs []models.EnvironmentItemModel `yml:"environments"`
}

// // Convert envsYMLModel to envModel array
// func (envYML envsYMLModel) convertToEnvModelArray() []models.EnvironmentItemModel {
// 	var envModels []models.EnvironmentItemModel
// 	for _, env := range envYML.Envs {
// 		envModels = append(envModels, env)
// 	}
// 	return envModels
// }

// func (eMap EnvMapItem) convertToEnvModel() models.EnvironmentItemModel {
// 	var eModel EnvModel
// 	for key, value := range eMap {
// 		if key != IsExpandKey {
// 			eModel.Key = key
// 			eModel.Value = value
// 		}
// 	}
// 	eModel.IsExpand = ParseBool(eMap[IsExpandKey], true)
// 	return eModel
// }

// ParseBool ...
func ParseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}

	lowercased := strings.ToLower(s)
	if lowercased == "yes" || lowercased == "y" {
		return true
	}
	if lowercased == "no" || lowercased == "n" {
		return false
	}

	value, err := strconv.ParseBool(s)
	if err != nil {
		log.Errorln("[ENVMAN] - isExpand: Failed to parse input:", err)
		return defaultValue
	}
	return value
}

// // Convert envModel array to envsYMLModel
// func convertToEnvsYMLModel(eModels []EnvModel) envsYMLModel {
// 	var envYML envsYMLModel
// 	var envMaps []EnvMapItem
// 	for _, eModel := range eModels {
// 		eMap := eModel.convertToEnvMap()
// 		envMaps = append(envMaps, eMap)
// 	}
// 	envYML.Envs = envMaps
// 	return envYML
// }

// func (eModel EnvModel) convertToEnvMap() EnvMapItem {
// 	eMap := make(EnvMapItem)
// 	if !eModel.IsExpand {
// 		eMap[IsExpandKey] = FalseKey
// 	}
// 	eMap[eModel.Key] = eModel.Value
// 	return eMap
// }
