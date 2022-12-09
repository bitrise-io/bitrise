package models

// StepRunStatus ...
type StepRunStatus int

const (
	StepRunStatusCodeSuccess                StepRunStatus = 0
	StepRunStatusCodeFailed                 StepRunStatus = 1
	StepRunStatusCodeFailedSkippable        StepRunStatus = 2
	StepRunStatusCodeSkipped                StepRunStatus = 3
	StepRunStatusCodeSkippedWithRunIf       StepRunStatus = 4
	StepRunStatusCodePreparationFailed      StepRunStatus = 5
	StepRunStatusAbortedWithCustomTimeout   StepRunStatus = 7 // step times out due to a custom timeout
	StepRunStatusAbortedWithNoOutputTimeout StepRunStatus = 8 // step times out due to no output received (hang)
)

func NewStepRunStatus(status string) StepRunStatus {
	switch status {
	case "success":
		return StepRunStatusCodeSuccess
	case "failed":
		return StepRunStatusCodeFailed
	case "failed_skippable":
		return StepRunStatusCodeFailedSkippable
	case "skipped":
		return StepRunStatusCodeSkipped
	case "skipped_with_run_if":
		return StepRunStatusCodeSkippedWithRunIf
	case "preparation_failed":
		return StepRunStatusCodePreparationFailed
	case "aborted_with_custom_timeout":
		return StepRunStatusAbortedWithCustomTimeout
	case "aborted_with_no_output":
		return StepRunStatusAbortedWithNoOutputTimeout
	default:
		return -1
	}
}

func (s StepRunStatus) String() string {
	switch s {
	case StepRunStatusCodeSuccess:
		return "success"
	case StepRunStatusCodeFailed:
		return "failed"
	case StepRunStatusCodeFailedSkippable:
		return "failed_skippable"
	case StepRunStatusCodeSkipped:
		return "skipped"
	case StepRunStatusCodeSkippedWithRunIf:
		return "skipped_with_run_if"
	case StepRunStatusCodePreparationFailed:
		return "preparation_failed"
	case StepRunStatusAbortedWithCustomTimeout:
		return "aborted_with_custom_timeout"
	case StepRunStatusAbortedWithNoOutputTimeout:
		return "aborted_with_no_output"
	default:
		return "unknown"
	}
}

func (s StepRunStatus) Name() string {
	switch s {
	case StepRunStatusCodeSuccess:
		return ""
	case StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed,
		StepRunStatusCodeFailedSkippable,
		StepRunStatusAbortedWithCustomTimeout,
		StepRunStatusAbortedWithNoOutputTimeout:
		return "Failed"
	case StepRunStatusCodeSkipped,
		StepRunStatusCodeSkippedWithRunIf:
		return "Skipped"
	default:
		return ""
	}
}
