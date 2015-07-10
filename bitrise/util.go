package bitrise

import (
	"encoding/json"
	"os"

	models "github.com/bitrise-io/bitrise-cli/models/models_0_9_0"
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
