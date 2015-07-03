package bitrise

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise-cli/models"
)

func ReadWorkFlowJson(pth string) (models.WorkFlowJsonStruct, error) {
	var workflow models.WorkFlowJsonStruct

	file, err := os.Open(pth)
	if err != nil {
		return models.WorkFlowJsonStruct{}, err
	}

	parser := json.NewDecoder(file)
	if err = parser.Decode(&workflow); err != nil {
		return models.WorkFlowJsonStruct{}, err
	}

	return workflow, nil
}
