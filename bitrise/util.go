package bitrise

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise-cli/models"
)

func ReadWorkflowJson(pth string) (models.WorkflowModel, error) {
	var workflow models.WorkflowModel

	if file, err := os.Open(pth); err != nil {
		return models.WorkflowModel{}, err
	} else {
		parser := json.NewDecoder(file)
		if err = parser.Decode(&workflow); err != nil {
			return models.WorkflowModel{}, err
		}
	}

	return workflow, nil
}
