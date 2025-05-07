package models

import (
	"fmt"
	"sort"
	"time"
)

// StepPhase represents a distinct phase in step execution
type StepPhase string

const (
	// StepPhaseActivation is the phase where step is activated (preparing dependencies, etc.)
	StepPhaseActivation StepPhase = "activation"
	// StepPhaseEnvPreparation is the phase where environment variables are prepared
	StepPhaseEnvPreparation StepPhase = "env_preparation"
	// StepPhaseDependencies is the phase where step dependencies are resolved
	StepPhaseDependencies StepPhase = "dependencies"
	// StepPhaseExecution is the phase where step is actually executing its main function
	StepPhaseExecution StepPhase = "execution"
	// StepPhaseContainer is the phase for container start/stop
	StepPhaseContainer StepPhase = "container"
)

// StepPhaseTiming tracks timing for a specific phase of step execution
type StepPhaseTiming struct {
	Phase    StepPhase `json:"phase"`
	Duration time.Duration `json:"duration_ms"`
}

// StepPerformanceMetrics captures detailed timing data for a step run
type StepPerformanceMetrics struct {
	StepID    string           `json:"step_id"`
	StepTitle string           `json:"step_title"` 
	StartTime time.Time        `json:"start_time"`
	EndTime   time.Time        `json:"end_time"`
	TotalTime time.Duration    `json:"total_time_ms"`
	Phases    []StepPhaseTiming `json:"phases"`
}

// AddPhase adds a phase timing record to the step metrics
func (spm *StepPerformanceMetrics) AddPhase(phase StepPhase, duration time.Duration) {
	spm.Phases = append(spm.Phases, StepPhaseTiming{
		Phase:    phase,
		Duration: duration,
	})
}

// CalculateTotalTime calculates the total time spent in this step
func (spm *StepPerformanceMetrics) CalculateTotalTime() {
	spm.TotalTime = spm.EndTime.Sub(spm.StartTime)
}

// WorkflowPerformanceMetrics captures performance data for an entire workflow execution
type WorkflowPerformanceMetrics struct {
	WorkflowID    string                  `json:"workflow_id"`
	WorkflowTitle string                  `json:"workflow_title"`
	StartTime     time.Time               `json:"start_time"`
	EndTime       time.Time               `json:"end_time"`
	TotalTime     time.Duration           `json:"total_time_ms"`
	Steps         []StepPerformanceMetrics `json:"steps"`
}

// AddStep adds a step's performance metrics to the workflow metrics
func (wpm *WorkflowPerformanceMetrics) AddStep(stepMetrics StepPerformanceMetrics) {
	wpm.Steps = append(wpm.Steps, stepMetrics)
}

// CalculateTotalTime calculates the total time spent in this workflow
func (wpm *WorkflowPerformanceMetrics) CalculateTotalTime() {
	wpm.TotalTime = wpm.EndTime.Sub(wpm.StartTime)
}

// SortStepsByExecution sorts steps by execution time in descending order
func (wpm *WorkflowPerformanceMetrics) SortStepsByExecution() {
	sort.Slice(wpm.Steps, func(i, j int) bool {
		return wpm.Steps[i].TotalTime > wpm.Steps[j].TotalTime
	})
}

// BuildPerformanceMetrics captures performance data for an entire build execution
type BuildPerformanceMetrics struct {
	BuildSlug  string                      `json:"build_slug,omitempty"`
	StartTime  time.Time                   `json:"start_time"`
	EndTime    time.Time                   `json:"end_time"`
	TotalTime  time.Duration               `json:"total_time_ms"`
	Workflows  []WorkflowPerformanceMetrics `json:"workflows"`
}

// AddWorkflow adds a workflow's performance metrics to the build metrics
func (bpm *BuildPerformanceMetrics) AddWorkflow(workflowMetrics WorkflowPerformanceMetrics) {
	bpm.Workflows = append(bpm.Workflows, workflowMetrics)
}

// CalculateTotalTime calculates the total time spent in this build
func (bpm *BuildPerformanceMetrics) CalculateTotalTime() {
	bpm.TotalTime = bpm.EndTime.Sub(bpm.StartTime)
}

// GetTopNSlowestSteps returns the top N slowest steps across all workflows
func (bpm *BuildPerformanceMetrics) GetTopNSlowestSteps(n int) []StepPerformanceMetrics {
	var allSteps []StepPerformanceMetrics
	for _, wf := range bpm.Workflows {
		allSteps = append(allSteps, wf.Steps...)
	}

	sort.Slice(allSteps, func(i, j int) bool {
		return allSteps[i].TotalTime > allSteps[j].TotalTime
	})

	if len(allSteps) < n {
		return allSteps
	}
	return allSteps[:n]
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if d.Minutes() >= 1 {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		ms := int(d.Milliseconds()) % 1000
		return fmt.Sprintf("%dm %ds %dms", m, s, ms)
	}
	if d.Seconds() >= 1 {
		s := int(d.Seconds())
		ms := int(d.Milliseconds()) % 1000
		return fmt.Sprintf("%ds %dms", s, ms)
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}