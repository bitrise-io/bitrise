package models

import (
	"time"
)

// TraceEvent represents a single event in a Chrome trace format
type TraceEvent struct {
	Name     string    `json:"name"`
	Cat      string    `json:"cat,omitempty"`
	Ph       string    `json:"ph"`
	Ts       int64     `json:"ts"`
	Dur      int64     `json:"dur,omitempty"`
	Pid      int       `json:"pid"`
	Tid      int       `json:"tid"`
	Args     TraceArgs `json:"args,omitempty"`
	ID       string    `json:"id,omitempty"`
	ParentID string    `json:"parentId,omitempty"`
}

// TraceArgs contains additional data for a trace event
type TraceArgs struct {
	StepTitle    string `json:"stepTitle,omitempty"`
	WorkflowID   string `json:"workflowId,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
	StepID       string `json:"stepId,omitempty"`
	Phase        string `json:"phase,omitempty"`
	Duration     string `json:"duration,omitempty"`
}

// TraceProfile is a collection of trace events in Chrome trace format
type TraceProfile struct {
	TraceEvents     []TraceEvent `json:"traceEvents"`
	DisplayTimeUnit string       `json:"displayTimeUnit"`
}

// NewTraceProfile creates a new trace profile from build performance metrics
func NewTraceProfile(metrics *BuildPerformanceMetrics) *TraceProfile {
	profile := &TraceProfile{
		TraceEvents:     []TraceEvent{},
		DisplayTimeUnit: "ms",
	}

	// Add build begin/end events
	buildStartTs := metrics.StartTime.UnixMicro() / 1000
	buildID := "build_1"

	// Add build begin event
	profile.TraceEvents = append(profile.TraceEvents, TraceEvent{
		Name: "Build",
		Ph:   "b", // Begin event
		Ts:   buildStartTs,
		Pid:  1,
		Tid:  1,
		ID:   buildID,
	})

	// Add build end event
	profile.TraceEvents = append(profile.TraceEvents, TraceEvent{
		Name: "Build",
		Ph:   "e", // End event
		Ts:   metrics.EndTime.UnixMicro() / 1000,
		Pid:  1,
		Tid:  1,
		ID:   buildID,
	})

	// Add workflow events
	for wfIdx, workflow := range metrics.Workflows {
		// workflowID := "workflow_" + workflow.WorkflowID
		workflowStartTs := workflow.StartTime.UnixMicro() / 1000

		// Add workflow duration event
		profile.TraceEvents = append(profile.TraceEvents, TraceEvent{
			Name: workflow.WorkflowTitle,
			Cat:  "workflow",
			Ph:   "X", // Complete event with duration
			Ts:   workflowStartTs,
			Dur:  workflow.TotalTime.Microseconds() / 1000,
			Pid:  1,
			Tid:  wfIdx + 2, // Use different thread IDs for workflows
			Args: TraceArgs{
				WorkflowID:   workflow.WorkflowID,
				WorkflowName: workflow.WorkflowTitle,
				Duration:     FormatDuration(workflow.TotalTime),
			},
		})

		// Add step events
		for stepIdx, step := range workflow.Steps {
			stepStartTs := step.StartTime.UnixMicro() / 1000

			// Add step duration event
			profile.TraceEvents = append(profile.TraceEvents, TraceEvent{
				Name: step.StepTitle,
				Cat:  "step",
				Ph:   "X", // Complete event with duration
				Ts:   stepStartTs,
				Dur:  step.TotalTime.Microseconds() / 1000,
				Pid:  1,
				Tid:  wfIdx + 2, // Same thread ID as parent workflow
				Args: TraceArgs{
					StepTitle:    step.StepTitle,
					StepID:       step.StepID,
					WorkflowID:   workflow.WorkflowID,
					WorkflowName: workflow.WorkflowTitle,
					Duration:     FormatDuration(step.TotalTime),
				},
			})

			// Add phase events
			for phaseIdx, phase := range step.Phases {
				phaseStartTs := step.StartTime.Add(time.Duration(calcPhaseStartOffset(step.Phases, phaseIdx))).UnixMicro() / 1000

				profile.TraceEvents = append(profile.TraceEvents, TraceEvent{
					Name: string(phase.Phase),
					Cat:  "phase",
					Ph:   "X", // Complete event with duration
					Ts:   phaseStartTs,
					Dur:  phase.Duration.Microseconds() / 1000,
					Pid:  2,           // Use different process ID for phases
					Tid:  stepIdx + 2, // Different thread ID for each step's phases
					Args: TraceArgs{
						StepTitle:    step.StepTitle,
						StepID:       step.StepID,
						Phase:        string(phase.Phase),
						WorkflowID:   workflow.WorkflowID,
						WorkflowName: workflow.WorkflowTitle,
						Duration:     FormatDuration(phase.Duration),
					},
				})
			}
		}
	}

	return profile
}

// calcPhaseStartOffset calculates an approximate offset for a phase within a step
// This is an approximation since we don't have exact phase ordering information
func calcPhaseStartOffset(phases []StepPhaseTiming, currentPhaseIdx int) int64 {
	var offset int64 = 0
	for i := 0; i < currentPhaseIdx; i++ {
		offset += phases[i].Duration.Microseconds()
	}
	return offset
}
