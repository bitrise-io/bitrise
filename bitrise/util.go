package bitrise

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise-cli/models"
)

func ReadWorkFlowJson(pth string) (models.WorkFlowModel, error) {
	var workflow models.WorkFlowModel

	file, err := os.Open(pth)
	if err != nil {
		return models.WorkFlowModel{}, err
	}

	parser := json.NewDecoder(file)
	if err = parser.Decode(&workflow); err != nil {
		return models.WorkFlowModel{}, err
	}

	return workflow, nil
}
