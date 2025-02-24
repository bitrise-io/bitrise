package models

import (
	"time"

	"github.com/bitrise-io/bitrise/v2/exitcode"
)

type BuildRunResultsModel struct {
	WorkflowID           string                `json:"workflow_id" yaml:"workflow_id"`
	EventName            string                `json:"event_name" yaml:"event_name"`
	ProjectType          string                `json:"project_type" yaml:"project_type"`
	StartTime            time.Time             `json:"start_time" yaml:"start_time"`
	StepmanUpdates       map[string]int        `json:"stepman_updates" yaml:"stepman_updates"`
	SuccessSteps         []StepRunResultsModel `json:"success_steps" yaml:"success_steps"`
	FailedSteps          []StepRunResultsModel `json:"failed_steps" yaml:"failed_steps"`
	FailedSkippableSteps []StepRunResultsModel `json:"failed_skippable_steps" yaml:"failed_skippable_steps"`
	SkippedSteps         []StepRunResultsModel `json:"skipped_steps" yaml:"skipped_steps"`
}

func NewBuildRunResultsModel(workflowID string, start time.Time, projectType string) BuildRunResultsModel {
	return BuildRunResultsModel{
		WorkflowID:     workflowID,
		StartTime:      start,
		StepmanUpdates: map[string]int{},
		ProjectType:    projectType,
	}
}

func (buildRes BuildRunResultsModel) IsStepLibUpdated(stepLib string) bool {
	return (buildRes.StepmanUpdates[stepLib] > 0)
}

func (buildRes BuildRunResultsModel) IsBuildFailed() bool {
	return len(buildRes.FailedSteps) > 0
}

func (buildRes BuildRunResultsModel) ExitCode() int {
	if !buildRes.IsBuildFailed() {
		return 0
	}

	if buildRes.isBuildAbortedWithNoOutputTimeout() {
		return exitcode.CLIAbortedWithNoOutputTimeout
	}

	if buildRes.isBuildAbortedWithTimeout() {
		return exitcode.CLIAbortedWithCustomTimeout
	}

	return exitcode.CLIFailed
}

func (buildRes BuildRunResultsModel) HasFailedSkippableSteps() bool {
	return len(buildRes.FailedSkippableSteps) > 0
}

func (buildRes BuildRunResultsModel) ResultsCount() int {
	return len(buildRes.SuccessSteps) + len(buildRes.FailedSteps) + len(buildRes.FailedSkippableSteps) + len(buildRes.SkippedSteps)
}

func (buildRes BuildRunResultsModel) isBuildAbortedWithTimeout() bool {
	for _, stepResult := range buildRes.FailedSteps {
		if stepResult.Status == StepRunStatusAbortedWithCustomTimeout {
			return true
		}
	}

	return false
}

func (buildRes BuildRunResultsModel) isBuildAbortedWithNoOutputTimeout() bool {
	for _, stepResult := range buildRes.FailedSteps {
		if stepResult.Status == StepRunStatusAbortedWithNoOutputTimeout {
			return true
		}
	}

	return false
}

func (buildRes BuildRunResultsModel) unorderedResults() []StepRunResultsModel {
	results := append([]StepRunResultsModel{}, buildRes.SuccessSteps...)
	results = append(results, buildRes.FailedSteps...)
	results = append(results, buildRes.FailedSkippableSteps...)
	return append(results, buildRes.SkippedSteps...)
}

func (buildRes BuildRunResultsModel) OrderedResults() []StepRunResultsModel {
	results := make([]StepRunResultsModel, buildRes.ResultsCount())
	unorderedResults := buildRes.unorderedResults()
	for _, result := range unorderedResults {
		results[result.Idx] = result
	}
	return results
}
