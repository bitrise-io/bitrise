package models

import (
	"fmt"
	"strings"
	"time"
)

const (
	FormatVersion = "25"
)

type BuildRunStartModel struct {
	EventName   string    `json:"event_name" yaml:"event_name"`
	ProjectType string    `json:"project_type" yaml:"project_type"`
	StartTime   time.Time `json:"start_time" yaml:"start_time"`
}

type StepError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s StepRunResultsModel) StatusReasonAndErrors() (string, []StepError) {
	switch s.Status {
	case StepRunStatusCodeSuccess:
		return "", nil
	case StepRunStatusCodeSkipped:
		return s.statusReason(), nil
	case StepRunStatusCodeSkippedWithRunIf:
		return s.statusReason(), nil
	case StepRunStatusCodeFailedSkippable:
		return s.statusReason(), s.error()
	case StepRunStatusCodeFailed:
		return "", s.error()
	case StepRunStatusCodePreparationFailed:
		return "", s.error()
	case StepRunStatusAbortedWithCustomTimeout:
		return "", s.error()
	case StepRunStatusAbortedWithNoOutputTimeout:
		return "", s.error()
	default:
		return "", nil
	}
}

func (s StepRunResultsModel) statusReason() string {
	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed,
		StepRunStatusAbortedWithCustomTimeout,
		StepRunStatusAbortedWithNoOutputTimeout:
		return ""
	case StepRunStatusCodeFailedSkippable:
		return `This Step failed, but it was marked as "is_skippable", so the build continued.`
	case StepRunStatusCodeSkipped:
		return `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`
	case StepRunStatusCodeSkippedWithRunIf:
		return fmt.Sprintf(`This Step was skipped, because its "run_if" expression evaluated to false.
The "run_if" expression was: %s`, *s.StepInfo.Step.RunIf)
	}

	return ""
}

func (s StepRunResultsModel) error() []StepError {
	message := ""

	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeSkipped,
		StepRunStatusCodeSkippedWithRunIf:
		return nil
	case StepRunStatusCodeFailedSkippable,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed:
		message = s.ErrorStr
	case StepRunStatusAbortedWithCustomTimeout:
		message = fmt.Sprintf("This Step timed out after %s.", formatStatusReasonTimeInterval(s.Timeout))
	case StepRunStatusAbortedWithNoOutputTimeout:
		message = fmt.Sprintf("This Step failed, because it has not sent any output for %s.", formatStatusReasonTimeInterval(s.NoOutputTimeout))
	}

	return []StepError{{
		Code:    s.ExitCode,
		Message: message,
	}}
}

func formatStatusReasonTimeInterval(timeInterval time.Duration) string {
	remaining := int(timeInterval / time.Second)
	h := int(remaining / 3600)
	remaining = remaining - h*3600
	m := int(remaining / 60)
	remaining = remaining - m*60
	s := remaining

	formattedTimeInterval := ""
	if h > 0 {
		formattedTimeInterval += fmt.Sprintf("%dh ", h)
	}

	if m > 0 {
		formattedTimeInterval += fmt.Sprintf("%dm ", m)
	}

	if s > 0 {
		formattedTimeInterval += fmt.Sprintf("%ds", s)
	}

	formattedTimeInterval = strings.TrimSpace(formattedTimeInterval)

	return formattedTimeInterval
}

type TestResultStepInfo struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Title   string `json:"title" yaml:"title"`
	Number  int    `json:"number" yaml:"number"`
}
