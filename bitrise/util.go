package bitrise

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise-cli/models"
)

// ReadWorkflowJSON ...
func ReadWorkflowJSON(pth string) (models.WorkflowModel, error) {
	var workflow models.WorkflowModel

	file, err := os.Open(pth)
	if err != nil {
		return models.WorkflowModel{}, err
	}

	parser := json.NewDecoder(file)
	if err = parser.Decode(&workflow); err != nil {
		return models.WorkflowModel{}, err
	}

	return workflow, nil
}
