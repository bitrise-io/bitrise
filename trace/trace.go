package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Logger handles trace event logging
type Logger struct {
	enabled bool
}

// NewLogger creates a new trace logger
func NewLogger(enabled bool) *Logger {
	return &Logger{enabled: enabled}
}

// Event represents a Bitrise trace event
type Event struct {
	Timestamp  int64  `json:"ts"`
	Type       string `json:"type"`
	ProcessID  int    `json:"pid"`
	ThreadID   int    `json:"tid"`
	StepID     string `json:"step_id,omitempty"`
	StepTitle  string `json:"step_title,omitempty"`
	Workflow   string `json:"workflow,omitempty"`
	Status     string `json:"status,omitempty"`
	DurationUs int64  `json:"duration_us,omitempty"`
	BuildID    string `json:"build_id,omitempty"`
	Project    string `json:"project,omitempty"`
}

// Event types
const (
	WorkflowStart = "workflow_start"
	WorkflowEnd   = "workflow_end"
	StepStart     = "step_start"
	StepEnd       = "step_end"
)

// LogWorkflowStart logs a workflow start event
func (l *Logger) LogWorkflowStart(workflow, workflowTitle string) {
	if !l.enabled {
		return
	}

	event := Event{
		Timestamp: time.Now().UnixMicro(),
		Type:      WorkflowStart,
		ProcessID: 1,
		ThreadID:  0,
		Workflow:  workflow,
		StepTitle: workflowTitle,
	}

	l.logEvent(event)
}

// LogWorkflowEnd logs a workflow end event
func (l *Logger) LogWorkflowEnd(workflow, workflowTitle, status string, startTime, endTime time.Time) {
	if !l.enabled {
		return
	}

	duration := endTime.Sub(startTime)

	event := Event{
		Timestamp:  endTime.UnixMicro(),
		Type:       WorkflowEnd,
		ProcessID:  1,
		ThreadID:   0,
		Workflow:   workflow,
		StepTitle:  workflowTitle,
		Status:     status,
		DurationUs: duration.Microseconds(),
	}

	l.logEvent(event)
}

// LogStepStart logs a step start event
func (l *Logger) LogStepStart(stepID, stepTitle, workflow string) {
	if !l.enabled {
		return
	}

	event := Event{
		Timestamp: time.Now().UnixMicro(),
		Type:      StepStart,
		ProcessID: 1,
		ThreadID:  1,
		StepID:    stepID,
		StepTitle: stepTitle,
		Workflow:  workflow,
	}

	l.logEvent(event)
}

// LogStepEnd logs a step end event
func (l *Logger) LogStepEnd(stepID, stepTitle, workflow, status string, startTime, endTime time.Time) {
	if !l.enabled {
		return
	}

	duration := endTime.Sub(startTime)

	event := Event{
		Timestamp:  endTime.UnixMicro(),
		Type:       StepEnd,
		ProcessID:  1,
		ThreadID:   1,
		StepID:     stepID,
		StepTitle:  stepTitle,
		Workflow:   workflow,
		Status:     status,
		DurationUs: duration.Microseconds(),
	}

	l.logEvent(event)
}

// logEvent outputs the trace event in the expected format
func (l *Logger) logEvent(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return // Silently fail to avoid disrupting the build
	}

	fmt.Fprintf(os.Stdout, "BITRISE_TRACE:%s\n", string(data))
}
