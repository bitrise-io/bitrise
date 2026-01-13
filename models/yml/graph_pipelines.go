package yml

import (
	"encoding/json"
	"fmt"
	"slices"
)

type GraphPipelineWorkflowModelInput map[string]interface{}

type GraphPipelineRunIfModel struct {
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
}

type GraphPipelineWorkflowModel struct {
	DependsOn       []string                          `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	AbortOnFail     bool                              `json:"abort_on_fail,omitempty" yaml:"abort_on_fail,omitempty"`
	RunIf           GraphPipelineRunIfModel           `json:"run_if,omitempty" yaml:"run_if,omitempty"`
	ShouldAlwaysRun GraphPipelineAlwaysRunMode        `json:"should_always_run,omitempty" yaml:"should_always_run,omitempty"`
	Uses            string                            `json:"uses,omitempty" yaml:"uses,omitempty"`
	Inputs          []GraphPipelineWorkflowModelInput `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Parallel        string                            `json:"parallel,omitempty" yaml:"parallel,omitempty"`
}
type GraphPipelineWorkflowListItemModel map[string]GraphPipelineWorkflowModel

const (
	GraphPipelineAlwaysRunModeOff      GraphPipelineAlwaysRunMode = "off"
	GraphPipelineAlwaysRunModeWorkflow GraphPipelineAlwaysRunMode = "workflow"
)

type GraphPipelineAlwaysRunMode string

func (d *GraphPipelineAlwaysRunMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}

	if err := validateGraphPipelineAlwaysRunMode(value); err != nil {
		return err
	}

	*d = GraphPipelineAlwaysRunMode(value)

	return nil
}

func (d *GraphPipelineAlwaysRunMode) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if err := validateGraphPipelineAlwaysRunMode(value); err != nil {
		return err
	}

	*d = GraphPipelineAlwaysRunMode(value)

	return nil
}

func validateGraphPipelineAlwaysRunMode(value string) error {
	allowedValues := []string{string(GraphPipelineAlwaysRunModeOff), string(GraphPipelineAlwaysRunModeWorkflow)}
	if !slices.Contains(allowedValues, value) {
		return fmt.Errorf("%s is not a valid should_always_run value (%s)", value, allowedValues)
	}

	return nil
}
