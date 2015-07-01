package bitrise

import (
	"encoding/json"
	"os"
)

func ReadWorkFlowJson(pth string) (WorkFlowJsonStruct, error) {
	var workflow WorkFlowJsonStruct

	file, err := os.Open(pth)
	if err != nil {
		return WorkFlowJsonStruct{}, err
	}

	parser := json.NewDecoder(file)
	if err = parser.Decode(&workflow); err != nil {
		return WorkFlowJsonStruct{}, err
	}

	return workflow, nil
}
