package bitrise

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise-cli/models"
)

func ReadWorkFlowJson(pth string) (models.WorkFlowModel, error) {
	var workflow models.WorkFlowModel

	if file, err := os.Open(pth); err != nil {
		return models.WorkFlowModel{}, err
	} else {
		parser := json.NewDecoder(file)
		if err = parser.Decode(&workflow); err != nil {
			return models.WorkFlowModel{}, err
		}
	}

	return workflow, nil
}
